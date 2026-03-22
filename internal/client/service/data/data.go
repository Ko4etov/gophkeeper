// Package data предоставляет сервис для работы с данными пользователя.
package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	grpcclient "github.com/Ko4etov/gophkeeper/internal/client/grpc_client"
	"github.com/Ko4etov/gophkeeper/internal/client/service/storage"
	"github.com/Ko4etov/gophkeeper/internal/models"
	"github.com/google/uuid"
)

// DataService управляет операциями с данными пользователя.
type DataService struct {
	grpcclient *grpcclient.GrpcClient
	storage    *storage.Storage
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewDataService создает новый сервис данных.
func NewDataService(ctx context.Context, grpcclient *grpcclient.GrpcClient, storage *storage.Storage) (*DataService, error) {
	serviceCtx, cancel := context.WithCancel(ctx)

	return &DataService{
		grpcclient: grpcclient,
		storage:    storage,
		ctx:        serviceCtx,
		cancel:     cancel,
	}, nil
}

// SaveEntry сохраняет запись в локальное хранилище.
func (c *DataService) SaveEntry(email string, entry *models.DataEntry) error {
	return c.storage.SaveEntry(email, entry)
}

// AddLogin добавляет запись логина/пароля.
func (c *DataService) AddLogin(name, login, password, url, notes string, tags []string, currentUserEmail string) error {
	entry := &models.DataEntry{
		Meta: models.DataMeta{
			ID:        uuid.New().String(),
			Name:      name,
			Tags:      tags,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
			DataType:  models.TypeLoginPassword,
		},
		Data: models.LoginPasswordData{
			Login:    login,
			Password: password,
			URL:      url,
			Notes:    notes,
		},
	}

	return c.storage.SaveEntry(currentUserEmail, entry)
}

// AddCard добавляет запись банковской карты.
func (c *DataService) AddCard(name string, number string, holder string, month int, year int, cvv string, cardType string, bank string, tags []string, currentUserEmail string) error {
	entry := &models.DataEntry{
		Meta: models.DataMeta{
			ID:        uuid.New().String(),
			Name:      name,
			Tags:      tags,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
			DataType:  models.TypeBankCard,
		},
		Data: models.BankCardData{
			CardNumber:  number,
			CardHolder:  holder,
			ExpiryMonth: month,
			ExpiryYear:  year,
			CVV:         cvv,
			CardType:    cardType,
			BankName:    bank,
		},
	}

	return c.storage.SaveEntry(currentUserEmail, entry)
}

// AddText добавляет текстовую запись.
func (c *DataService) AddText(name, content, format string, tags []string, currentUserEmail string) error {
	entry := &models.DataEntry{
		Meta: models.DataMeta{
			ID:        uuid.New().String(),
			Name:      name,
			Tags:      tags,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
			DataType:  models.TypeText,
		},
		Data: models.TextData{
			Content: content,
			Format:  format,
		},
	}

	return c.storage.SaveEntry(currentUserEmail, entry)
}

// AddBinary добавляет бинарную запись.
func (c *DataService) AddBinary(name, filename string, content []byte, mimeType string, tags []string, currentUserEmail string) error {
	entry := &models.DataEntry{
		Meta: models.DataMeta{
			ID:        uuid.New().String(),
			Name:      name,
			Tags:      tags,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
			DataType:  models.TypeBinary,
		},
		Data: models.BinaryData{
			Filename: filename,
			Content:  content,
			MimeType: mimeType,
			Size:     int64(len(content)),
		},
	}

	return c.storage.SaveEntry(currentUserEmail, entry)
}

// ListMeta возвращает метаданные всех записей (без данных).
func (c *DataService) ListMeta(currentUserEmail string) ([]*models.DataMeta, error) {
	return c.storage.ListEntriesMeta(currentUserEmail)
}

// Get возвращает полную запись по ID.
func (c *DataService) Get(currentUserEmail string, id string) (*models.DataEntry, error) {
	return c.storage.GetEntry(currentUserEmail, id)
}

// GetLogin возвращает запись логина/пароля по ID.
func (c *DataService) GetLogin(id string, currentUserEmail string) (*models.DataEntry, error) {
	entry, err := c.storage.GetEntry(currentUserEmail, id)
	if err != nil {
		return nil, err
	}
	if entry.Meta.DataType != models.TypeLoginPassword {
		return nil, ErrWrongDataType
	}
	return entry, nil
}

// GetCard возвращает запись банковской карты по ID.
func (c *DataService) GetCard(id string, currentUserEmail string) (*models.DataEntry, error) {
	entry, err := c.storage.GetEntry(currentUserEmail, id)
	if err != nil {
		return nil, err
	}
	if entry.Meta.DataType != models.TypeBankCard {
		return nil, ErrWrongDataType
	}
	return entry, nil
}

// GetText возвращает текстовую запись по ID.
func (c *DataService) GetText(id string, currentUserEmail string) (*models.DataEntry, error) {
	entry, err := c.storage.GetEntry(currentUserEmail, id)
	if err != nil {
		return nil, err
	}
	if entry.Meta.DataType != models.TypeText {
		return nil, ErrWrongDataType
	}
	return entry, nil
}

// GetBinary возвращает бинарную запись по ID.
func (c *DataService) GetBinary(id string, currentUserEmail string) (*models.DataEntry, error) {
	entry, err := c.storage.GetEntry(currentUserEmail, id)
	if err != nil {
		return nil, err
	}
	if entry.Meta.DataType != models.TypeBinary {
		return nil, ErrWrongDataType
	}
	return entry, nil
}

// UpdateLogin обновляет существующую запись логина/пароля.
func (c *DataService) UpdateLogin(id, name, login, password, url, notes string, tags []string, currentUserEmail string) error {
	existing, err := c.storage.GetEntry(currentUserEmail, id)
	if err != nil {
		return err
	}

	if existing.Meta.DataType != models.TypeLoginPassword {
		return ErrWrongDataType
	}

	entry := &models.DataEntry{
		Meta: models.DataMeta{
			ID:          id,
			Name:        name,
			Description: existing.Meta.Description,
			Tags:        tags,
			CreatedAt:   existing.Meta.CreatedAt,
			UpdatedAt:   time.Now(),
			Version:     existing.Meta.Version + 1,
			DataType:    models.TypeLoginPassword,
		},
		Data: models.LoginPasswordData{
			Login:    login,
			Password: password,
			URL:      url,
			Notes:    notes,
		},
	}

	return c.storage.SaveEntry(currentUserEmail, entry)
}

// UpdateCard обновляет существующую запись банковской карты.
func (c *DataService) UpdateCard(id, name, number, holder string, month, year int, cvv, cardType, bank string, tags []string, currentUserEmail string) error {
	existing, err := c.storage.GetEntry(currentUserEmail, id)
	if err != nil {
		return err
	}

	if existing.Meta.DataType != models.TypeBankCard {
		return ErrWrongDataType
	}

	entry := &models.DataEntry{
		Meta: models.DataMeta{
			ID:          id,
			Name:        name,
			Description: existing.Meta.Description,
			Tags:        tags,
			CreatedAt:   existing.Meta.CreatedAt,
			UpdatedAt:   time.Now(),
			Version:     existing.Meta.Version + 1,
			DataType:    models.TypeBankCard,
		},
		Data: models.BankCardData{
			CardNumber:  number,
			CardHolder:  holder,
			ExpiryMonth: month,
			ExpiryYear:  year,
			CVV:         cvv,
			CardType:    cardType,
			BankName:    bank,
		},
	}

	return c.storage.SaveEntry(currentUserEmail, entry)
}

// UpdateText обновляет существующую текстовую запись.
func (c *DataService) UpdateText(id, name, content, format string, tags []string, currentUserEmail string) error {
	existing, err := c.storage.GetEntry(currentUserEmail, id)
	if err != nil {
		return err
	}

	if existing.Meta.DataType != models.TypeText {
		return ErrWrongDataType
	}

	entry := &models.DataEntry{
		Meta: models.DataMeta{
			ID:          id,
			Name:        name,
			Description: existing.Meta.Description,
			Tags:        tags,
			CreatedAt:   existing.Meta.CreatedAt,
			UpdatedAt:   time.Now(),
			Version:     existing.Meta.Version + 1,
			DataType:    models.TypeText,
		},
		Data: models.TextData{
			Content: content,
			Format:  format,
		},
	}

	return c.storage.SaveEntry(currentUserEmail, entry)
}

// UpdateBinaryMetadata обновляет метаданные бинарного файла.
func (c *DataService) UpdateBinaryMetadata(id, name, filename, mimeType string, tags []string, currentUserEmail string) error {
	existing, err := c.storage.GetEntry(currentUserEmail, id)
	if err != nil {
		return err
	}

	if existing.Meta.DataType != models.TypeBinary {
		return ErrWrongDataType
	}

	binaryData, ok := existing.Data.(models.BinaryData)
	if !ok {
		return fmt.Errorf("invalid binary data")
	}

	binaryData.Filename = filename
	binaryData.MimeType = mimeType

	entry := &models.DataEntry{
		Meta: models.DataMeta{
			ID:          id,
			Name:        name,
			Description: existing.Meta.Description,
			Tags:        tags,
			CreatedAt:   existing.Meta.CreatedAt,
			UpdatedAt:   time.Now(),
			Version:     existing.Meta.Version + 1,
			DataType:    models.TypeBinary,
			DeletedAt:   existing.Meta.DeletedAt,
			SyncedAt:    existing.Meta.SyncedAt,
		},
		Data: binaryData,
	}

	return c.storage.SaveEntry(currentUserEmail, entry)
}

// UpdateBinaryFull обновляет и метаданные, и содержимое бинарного файла.
func (c *DataService) UpdateBinaryFull(id, name, filename string, content []byte, mimeType string, tags []string, currentUserEmail string) error {
	existing, err := c.storage.GetEntry(currentUserEmail, id)
	if err != nil {
		return err
	}

	if existing.Meta.DataType != models.TypeBinary {
		return ErrWrongDataType
	}

	entry := &models.DataEntry{
		Meta: models.DataMeta{
			ID:        id,
			Name:      name,
			Tags:      tags,
			CreatedAt: existing.Meta.CreatedAt,
			UpdatedAt: time.Now(),
			Version:   existing.Meta.Version + 1,
			DataType:  models.TypeBinary,
		},
		Data: models.BinaryData{
			Filename: filename,
			Content:  content,
			MimeType: mimeType,
			Size:     int64(len(content)),
		},
	}

	return c.storage.SaveEntry(currentUserEmail, entry)
}

// Delete удаляет запись по ID (мягкое удаление).
func (c *DataService) Delete(id string, currentUserEmail string) error {
	existing, err := c.storage.GetEntry(currentUserEmail, id)
	if err != nil {
		return err
	}

	existing.Meta.DeletedAt = time.Now()
	existing.Meta.UpdatedAt = time.Now()

	return c.storage.SaveEntry(currentUserEmail, existing)
}

// BatchDelete удаляет несколько записей по списку ID.
func (c *DataService) BatchDelete(email string, ids []string) error {
	return c.storage.BatchDelete(email, ids)
}

// GetSalt возвращает соль для пользователя.
func (s *DataService) GetSalt(currentUserEmail string) ([]byte, error) {
	return s.storage.GetSalt(currentUserEmail)
}

// SaveSalt сохраняет соль для пользователя.
func (s *DataService) SaveSalt(salt []byte, currentUserEmail string) error {
	return s.storage.SaveSalt(currentUserEmail, salt)
}

// ErrWrongDataType возникает при попытке доступа к данным неверного типа.
var ErrWrongDataType = errors.New("wrong data type for this operation")