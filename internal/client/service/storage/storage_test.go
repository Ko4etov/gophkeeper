// internal/client/service/storage/storage_test.go
package storage

import (
	"context"
	"testing"
	"time"

	"github.com/Ko4etov/gophkeeper/internal/client/config"
	"github.com/Ko4etov/gophkeeper/internal/client/crypto"
	"github.com/Ko4etov/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestStorage создает временное хранилище для тестов
func createTestStorage(t *testing.T) *Storage {
	session := crypto.NewSession(context.Background(), 15 * time.Minute)
	
	// Разблокируем сессию
	masterPassword := "test-master-password-123!"
	salt := []byte("test-salt-for-unlock-12345678")
	err := session.Unlock(masterPassword, salt)
	require.NoError(t, err)
	
	cfg := &config.ClientConfig{
		DataDir: t.TempDir(),
	}
	
	st, err := NewStorage(cfg, session)
	require.NoError(t, err)
	
	return st
}

// setupTestUser создает тестового пользователя
func setupTestUser(t *testing.T, st *Storage) string {
	email := "test@example.com"
	
	err := st.EnsureUserBuckets(email)
	require.NoError(t, err)
	
	return email
}

func TestNewStorage(t *testing.T) {
	st := createTestStorage(t)
	assert.NotNil(t, st)
	defer st.Close()
}

func TestEnsureUserBuckets(t *testing.T) {
	st := createTestStorage(t)
	defer st.Close()
	
	email := "test@example.com"
	err := st.EnsureUserBuckets(email)
	assert.NoError(t, err)
}

func TestSaveAndGetSalt(t *testing.T) {
	st := createTestStorage(t)
	defer st.Close()
	
	email := setupTestUser(t, st)
	
	salt := []byte("test-salt-1234567890123456")
	err := st.SaveSalt(email, salt)
	require.NoError(t, err)
	
	retrieved, err := st.GetSalt(email)
	require.NoError(t, err)
	assert.Equal(t, salt, retrieved)
}

func TestSaveAndGetEntry(t *testing.T) {
	st := createTestStorage(t)
	defer st.Close()
	
	email := setupTestUser(t, st)
	
	entry := &models.DataEntry{
		Meta: models.DataMeta{
			ID:        "test-id-1",
			Name:      "Test Entry",
			DataType:  models.TypeText,
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Data: models.TextData{
			Content: "test content",
			Format:  "text/plain",
		},
	}
	
	err := st.SaveEntry(email, entry)
	require.NoError(t, err)
	
	retrieved, err := st.GetEntry(email, "test-id-1")
	require.NoError(t, err)
	assert.Equal(t, entry.Meta.Name, retrieved.Meta.Name)
	assert.Equal(t, entry.Meta.DataType, retrieved.Meta.DataType)
}

func TestListEntriesMeta(t *testing.T) {
	st := createTestStorage(t)
	defer st.Close()
	
	email := setupTestUser(t, st)
	
	// Добавляем несколько записей
	for i := 0; i < 3; i++ {
		entry := &models.DataEntry{
			Meta: models.DataMeta{
				ID:        "test-id-" + string(rune('a'+i)),
				Name:      "Test Entry",
				DataType:  models.TypeText,
				Version:   1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Data: models.TextData{Content: "test", Format: "text"},
		}
		err := st.SaveEntry(email, entry)
		require.NoError(t, err)
	}
	
	metas, err := st.ListEntriesMeta(email)
	require.NoError(t, err)
	assert.Len(t, metas, 3)
}

func TestDeleteEntry(t *testing.T) {
	st := createTestStorage(t)
	defer st.Close()
	
	email := setupTestUser(t, st)
	
	entry := &models.DataEntry{
		Meta: models.DataMeta{
			ID:        "delete-id",
			Name:      "To Delete",
			DataType:  models.TypeText,
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Data: models.TextData{Content: "test", Format: "text"},
	}
	
	err := st.SaveEntry(email, entry)
	require.NoError(t, err)
	
	err = st.DeleteEntry(email, "delete-id")
	require.NoError(t, err)
	
	_, err = st.GetEntry(email, "delete-id")
	assert.Error(t, err)
}

func TestBatchDelete(t *testing.T) {
	st := createTestStorage(t)
	defer st.Close()
	
	email := setupTestUser(t, st)
	
	// Добавляем записи
	ids := []string{"id-1", "id-2", "id-3"}
	for _, id := range ids {
		entry := &models.DataEntry{
			Meta: models.DataMeta{
				ID:        id,
				Name:      "Test",
				DataType:  models.TypeText,
				Version:   1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Data: models.TextData{Content: "test", Format: "text"},
		}
		err := st.SaveEntry(email, entry)
		require.NoError(t, err)
	}
	
	// Удаляем две записи
	err := st.BatchDelete(email, []string{"id-1", "id-2"})
	require.NoError(t, err)
	
	// Проверяем
	_, err = st.GetEntry(email, "id-1")
	assert.Error(t, err)
	_, err = st.GetEntry(email, "id-2")
	assert.Error(t, err)
	
	// Третья запись должна остаться
	_, err = st.GetEntry(email, "id-3")
	assert.NoError(t, err)
}

func TestGetEntriesByIDs(t *testing.T) {
	st := createTestStorage(t)
	defer st.Close()
	
	email := setupTestUser(t, st)
	
	// Добавляем записи
	ids := []string{"id-1", "id-2", "id-3"}
	for i, id := range ids {
		entry := &models.DataEntry{
			Meta: models.DataMeta{
				ID:        id,
				Name:      "Test",
				DataType:  models.TypeText,
				Version:   1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Data: models.TextData{
				Content: "test content " + id,
				Format:  "text/plain",
			},
		}
		err := st.SaveEntry(email, entry)
		require.NoError(t, err)
		_ = i
	}
	
	// Используем GetEntriesByIDs - он уже должен правильно расшифровывать
	entries, err := st.GetEntriesByIDs(email, []string{"id-1", "id-3"})
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	
	// Проверяем содержимое
	for _, entry := range entries {
		assert.Contains(t, []string{"id-1", "id-3"}, entry.Meta.ID)
		textData, ok := entry.Data.(models.TextData)
		assert.True(t, ok)
		assert.Contains(t, textData.Content, "test content")
	}
}