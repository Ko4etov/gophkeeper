// internal/client/service/sync/manager.go
package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	grpcclient "github.com/Ko4etov/gophkeeper/internal/client/grpc_client"
	"github.com/Ko4etov/gophkeeper/internal/client/service/auth"
	"github.com/Ko4etov/gophkeeper/internal/client/service/data"
	"github.com/Ko4etov/gophkeeper/internal/models"
	protosync "github.com/Ko4etov/gophkeeper/internal/proto/sync"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SyncManager управляет синхронизацией данных с сервером
type SyncManager struct {
	grpcClient  *grpcclient.GrpcClient
	dataService *data.DataService
	authService *auth.AuthService
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex
	isSyncing   bool
	lastSync    time.Time
	forceSyncCh chan struct{}
}

// NewSyncManager создает новый менеджер синхронизации
func NewSyncManager(
	grpcClient *grpcclient.GrpcClient,
	dataService *data.DataService,
	authService *auth.AuthService,
) *SyncManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &SyncManager{
		grpcClient:  grpcClient,
		dataService: dataService,
		authService: authService,
		ctx:         ctx,
		cancel:      cancel,
		forceSyncCh: make(chan struct{}, 1),
	}
}

// Start запускает фоновую синхронизацию
func (m *SyncManager) Start(interval time.Duration) {
	m.wg.Add(1)
	go m.backgroundSync(interval)
}

// Stop останавливает фоновую синхронизацию
func (m *SyncManager) Stop() {
	m.cancel()
	m.wg.Wait()
}

// backgroundSync выполняет фоновую синхронизацию
func (m *SyncManager) backgroundSync(interval time.Duration) {
	defer m.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Первая синхронизация через 5 секунд
	time.Sleep(5 * time.Second)
	m.trySync()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.trySync()
		case <-m.forceSyncCh:
			m.trySync()
		}
	}
}

// trySync пытается выполнить синхронизацию
func (m *SyncManager) trySync() {
	m.mu.Lock()
	if m.isSyncing {
		m.mu.Unlock()
		return
	}
	m.isSyncing = true
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.isSyncing = false
		m.mu.Unlock()
	}()

	user := m.authService.GetUser()
	if user == nil {
		return
	}

	if err := m.sync(user.Email); err != nil {
		fmt.Printf("Sync failed: %v\n", err)
	}

	m.mu.Lock()
	m.lastSync = time.Now()
	m.mu.Unlock()
}

// ForceSync принудительно запускает синхронизацию
func (m *SyncManager) ForceSync() {
	select {
	case m.forceSyncCh <- struct{}{}:
	default:
	}
}

// sync выполняет основную логику синхронизации
func (m *SyncManager) sync(email string) error {
	user := m.authService.GetUser()
	if user == nil {
		return fmt.Errorf("not logged in")
	}

	ctx := context.Background()

	stream, err := m.grpcClient.Sync(ctx, user.Token)
	if err != nil {
		return fmt.Errorf("failed to open sync stream: %w", err)
	}
	defer func() {
		_ = stream.CloseSend()
	}()

	localMetas, err := m.dataService.ListMeta(email)
	if err != nil {
		return fmt.Errorf("failed to get local metas: %w", err)
	}

	var metas []*protosync.EntryMeta
	for _, meta := range localMetas {
		meta := &protosync.EntryMeta{
			Id:        meta.ID,
			Version:   int32(meta.Version),
			UpdatedAt: timestamppb.New(meta.UpdatedAt),
			DeletedAt: timestamppb.New(meta.DeletedAt),
			SyncedAt:  timestamppb.New(meta.SyncedAt),
		}
		metas = append(metas, meta)
	}

	if err := stream.Send(&protosync.SyncRequest{
		Payload: &protosync.SyncRequest_Snapshot{
			Snapshot: &protosync.ClientSnapshot{
				ClientId: user.ID,
				Metas:    metas,
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to send snapshot: %w", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive instructions: %w", err)
	}

	instructions := resp.GetInstructions()
	if instructions == nil {
		return fmt.Errorf("expected instructions, got %T", resp.Payload)
	}

	if len(instructions.NeedDelete) > 0 {
		if err := m.dataService.BatchDelete(email, instructions.NeedDelete); err != nil {
			fmt.Printf("Failed to delete records: %v\n", err)
		}
	}

	if len(instructions.NeedFromClient) > 0 {
		for _, id := range instructions.NeedFromClient {
			encryptedData, err := m.dataService.GetEncryptedEntry(email, id)
			if err != nil {
				fmt.Printf("Failed to get encrypted entry %s: %v\n", id, err)
				continue
			}

			// Получаем метаданные для записи
			entry, err := m.dataService.Get(email, id)
			if err != nil {
				fmt.Printf("Failed to get entry %s: %v\n", id, err)
				continue
			}

			protoEntry := m.convertModelToProtoWithEncryptedData(entry, user.ID, encryptedData)
			if protoEntry == nil {
				continue
			}

			if err := stream.Send(&protosync.SyncRequest{
				Payload: &protosync.SyncRequest_Entry{
					Entry: protoEntry,
				},
			}); err != nil {
				return fmt.Errorf("failed to send entry %s: %w", id, err)
			}

			entry.Meta.SyncedAt = time.Now()
			m.dataService.SaveEntry(user.Email, entry)
		}
	}

	for _, id := range instructions.NeedDownload {
		resp, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("failed to receive entry %s: %w", id, err)
		}

		entry := resp.GetEntry()
		if entry == nil {
			return fmt.Errorf("expected entry, got %T", resp.Payload)
		}

		m.saveProtoEntryLocally(email, entry)
	}

	finalResp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive completion: %w", err)
	}

	complete := finalResp.GetComplete()
	if complete == nil {
		return fmt.Errorf("expected complete, got %T", finalResp.Payload)
	}

	fmt.Printf("Sync completed: %d entries synced\n", complete.SyncedCount)
	return nil
}

// convertModelToProto преобразует models.DataEntry в protosync.Entry
func (m *SyncManager) convertModelToProto(entry *models.DataEntry, userID string) *protosync.Entry {
	encryptedData, err := m.dataService.GetEncryptedEntry(userID, entry.Meta.ID)
	if err != nil {
		fmt.Printf("Failed to get encrypted data for %s: %v\n", entry.Meta.ID, err)
		return nil
	}

	return &protosync.Entry{
		Id:            entry.Meta.ID,
		UserId:        userID,
		Name:          entry.Meta.Name,
		Tags:          entry.Meta.Tags,
		DataType:      string(entry.Meta.DataType),
		CreatedAt:     timestamppb.New(entry.Meta.CreatedAt),
		UpdatedAt:     timestamppb.New(entry.Meta.UpdatedAt),
		Version:       int32(entry.Meta.Version),
		EncryptedData: encryptedData, // ✅ отправляем зашифрованные данные
	}
}

// saveProtoEntryLocally сохраняет protosync.Entry в локальное хранилище
func (m *SyncManager) saveProtoEntryLocally(email string, protoEntry *protosync.Entry) {
    entry := &models.DataEntry{
        Meta: models.DataMeta{
            ID:        protoEntry.Id,
            Name:      protoEntry.Name,
            Tags:      protoEntry.Tags,
            CreatedAt: protoEntry.CreatedAt.AsTime(),
            UpdatedAt: protoEntry.UpdatedAt.AsTime(),
            Version:   int(protoEntry.Version),
            DataType:  models.DataType(protoEntry.DataType),
        },
    }

    // Сохраняем зашифрованные данные напрямую (storage сам расшифрует при чтении)
    // Передаем encrypted_data как есть
    if err := m.dataService.SaveEncryptedEntry(email, entry, protoEntry.EncryptedData); err != nil {
        fmt.Printf("Failed to save entry %s: %v\n", protoEntry.Id, err)
    }
}

func (m *SyncManager) convertModelToProtoWithEncryptedData(entry *models.DataEntry, userID string, encryptedData string) *protosync.Entry {
	protoEntry := &protosync.Entry{
		Id:            entry.Meta.ID,
		UserId:        userID,
		Name:          entry.Meta.Name,
		Tags:          entry.Meta.Tags,
		DataType:      string(entry.Meta.DataType),
		CreatedAt:     timestamppb.New(entry.Meta.CreatedAt),
		UpdatedAt:     timestamppb.New(entry.Meta.UpdatedAt),
		Version:       int32(entry.Meta.Version),
		EncryptedData: encryptedData,
	}

	return protoEntry
}

// GetLastSync возвращает время последней синхронизации
func (m *SyncManager) GetLastSync() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastSync
}

// IsSyncing возвращает статус синхронизации
func (m *SyncManager) IsSyncing() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isSyncing
}
