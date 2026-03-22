package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	syncproto "github.com/Ko4etov/gophkeeper/internal/proto/sync"
	"github.com/Ko4etov/gophkeeper/internal/server/auth"
	"github.com/Ko4etov/gophkeeper/internal/server/service/logger"
	"github.com/Ko4etov/gophkeeper/internal/server/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SyncService struct {
	syncproto.UnimplementedSyncServiceServer
	storage *storage.Storage
}

func NewSyncService(storage *storage.Storage) *SyncService {
	return &SyncService{
		storage: storage,
	}
}

// Sync - основной метод синхронизации
func (s *SyncService) Sync(stream syncproto.SyncService_SyncServer) error {
	ctx, cancel := context.WithTimeout(stream.Context(), 3*time.Minute)
	defer cancel()

	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "unauthorized")
	}

	req, err := stream.Recv()
	if err != nil {
		return err
	}

	snapshot := req.GetSnapshot()
    if snapshot == nil {
        return status.Error(codes.InvalidArgument, "first message must be snapshot")
    }

	serverRecords, err := s.storage.GetAllRecordsMeta(ctx, userID)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get server records: %v", err)
	}

	serverMap := make(map[string]*storage.RecordMeta)
	for _, record := range serverRecords {
		serverMap[record.ClientID] = record
	}

	clientMetaMap := make(map[string]*syncproto.EntryMeta)
	for _, meta := range snapshot.Metas {
		clientMetaMap[meta.Id] = meta
	}

	var needDownload []string
	var needDelete []string
	var needFromClient []string

	for id, serverRecord := range serverMap {
		clientMeta, exists := clientMetaMap[id]

		if !exists {
			needDownload = append(needDownload, id)
			continue
		}

		if !isTsZero(clientMeta.DeletedAt) {
			// Клиент пометил как удаленную
			if err := s.storage.DeleteRecord(ctx, userID, id); err != nil {
				logger.Logger.Errorf("Failed to delete record %s: %v", id, err)
			} else {
				logger.Logger.Infof("Deleted record on server: %s", id)
			}
			needDelete = append(needDelete, id)
			continue
		}

		if clientMeta.Version > serverRecord.Version {
			needFromClient = append(needFromClient, id)
		} else if serverRecord.Version > clientMeta.Version {
			needDownload = append(needDownload, id)
		}

		delete(clientMetaMap, id)
	}

	for id, clientMeta := range clientMetaMap {
		if !isTsZero(clientMeta.DeletedAt) {
			continue
		}

		if isTsZero(clientMeta.SyncedAt) {
			needFromClient = append(needFromClient, id)
		} else {
			needDelete = append(needDelete, id)
		}
	}

	instructions := &syncproto.SyncInstructions{
		NeedDownload:   needDownload,
		NeedDelete:     needDelete,
		NeedFromClient: needFromClient,
	}

	if err := stream.Send(&syncproto.SyncResponse{
		Payload: &syncproto.SyncResponse_Instructions{
			Instructions: instructions,
		},
	}); err != nil {
		return err
	}

	for i := 0; i < len(needFromClient); i++ {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		entry := msg.GetEntry()
		if entry == nil {
			return status.Error(codes.InvalidArgument, "expected entry")
		}

		record, err := s.convertProtoEntryToModel(entry)
		if err != nil {
			logger.Logger.Errorf("Failed to convert entry %s: %v", entry.Id, err)
			continue
		}

		if _, exists := serverMap[entry.Id]; exists {
			err := s.storage.UpdateRecord(ctx, userID, record)
			if err != nil {
				logger.Logger.Info("cant update record: %w", err)
			}
		} else {
			err := s.storage.CreateRecord(ctx, userID, record)
			if err != nil {
				logger.Logger.Errorf("cant create record: %w", err)
			}
		}
	}

	if len(needDownload) > 0 {
		fullRecords, err := s.storage.GetFullRecordsByIDs(ctx, userID, needDownload)
		if err != nil {
			logger.Logger.Errorf("Failed to get full records: %v", err)
		} else {
			for _, record := range fullRecords {
				entry := s.convertToProtoEntry(record)
				if entry == nil {
					continue
				}

				if err := stream.Send(&syncproto.SyncResponse{
					Payload: &syncproto.SyncResponse_Entry{
						Entry: entry,
					},
				}); err != nil {
					return err
				}
				logger.Logger.Infof("Sent entry to client: %s (version %d)", record.ClientID, record.Version)
			}
		}
	}

	syncedRecords := len(needDownload) + len(needDelete) + len(needFromClient)

	return stream.Send(&syncproto.SyncResponse{
		Payload: &syncproto.SyncResponse_Complete{
			Complete: &syncproto.SyncComplete{
				SyncedCount: int32(syncedRecords),
			},
		},
	})
}

func isTsZero(ts *timestamppb.Timestamp) bool {
    if ts == nil {
        return true
    }
    if ts.Seconds == -62135596800 && ts.Nanos == 0 {
        return true
    }
    return ts.Seconds == 0 && ts.Nanos == 0
}

// extractDataJSON извлекает JSON из entry
func (s *SyncService) extractDataJSON(entry *syncproto.Entry) (json.RawMessage, error) {
	switch data := entry.Data.(type) {
	case *syncproto.Entry_LoginData:
		return json.Marshal(data.LoginData)
	case *syncproto.Entry_CardData:
		return json.Marshal(data.CardData)
	case *syncproto.Entry_TextData:
		return json.Marshal(data.TextData)
	case *syncproto.Entry_BinaryData:
		return json.Marshal(data.BinaryData)
	default:
		return nil, nil
	}
}

// convertToProtoEntry конвертирует ServerRecord в proto Entry
func (s *SyncService) convertToProtoEntry(record *storage.Record) *syncproto.Entry {
	entry := &syncproto.Entry{
		Id:          record.ClientID,
		UserId:      record.UserID,
		Name:        record.Name,
		Tags:        record.Tags,
		DataType:    record.DataType,
		Version:     record.Version,
		CreatedAt:   timestamppb.New(record.CreatedAt),
		UpdatedAt:   timestamppb.New(record.UpdatedAt),
	}

	switch record.DataType {
	case "login_password":
		var data syncproto.LoginPasswordData
		if err := json.Unmarshal(record.Data, &data); err == nil {
			entry.Data = &syncproto.Entry_LoginData{LoginData: &data}
		}
	case "bank_card":
		var data syncproto.BankCardData
		if err := json.Unmarshal(record.Data, &data); err == nil {
			entry.Data = &syncproto.Entry_CardData{CardData: &data}
		}
	case "text":
		var data syncproto.TextData
		if err := json.Unmarshal(record.Data, &data); err == nil {
			entry.Data = &syncproto.Entry_TextData{TextData: &data}
		}
	case "binary":
		var data syncproto.BinaryData
		if err := json.Unmarshal(record.Data, &data); err == nil {
			entry.Data = &syncproto.Entry_BinaryData{BinaryData: &data}
		}
	}

	return entry
}

func (s *SyncService) convertProtoEntryToModel(entry *syncproto.Entry) (*storage.Record, error) {
	record := &storage.Record{
		ClientID:  entry.Id,
		UserID:    entry.UserId,
		DataType:  entry.DataType,
		Name:      entry.Name,
		Tags:      entry.Tags,
		Version:   entry.Version,
		CreatedAt: entry.CreatedAt.AsTime(),
		UpdatedAt: entry.UpdatedAt.AsTime(),
	}

	// Извлекаем и сохраняем Data в зависимости от типа
	var err error
	switch data := entry.Data.(type) {
	case *syncproto.Entry_LoginData:
		record.Data, err = json.Marshal(data.LoginData)
	case *syncproto.Entry_CardData:
		record.Data, err = json.Marshal(data.CardData)
	case *syncproto.Entry_TextData:
		record.Data, err = json.Marshal(data.TextData)
	case *syncproto.Entry_BinaryData:
		record.Data, err = json.Marshal(data.BinaryData)
	default:
		return nil, fmt.Errorf("unknown data type for entry %s", entry.Id)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal data for entry %s: %w", entry.Id, err)
	}

	return record, nil
}