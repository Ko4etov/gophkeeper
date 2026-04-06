// internal/client/service/data/data_test.go
package data

import (
	"context"
	"testing"
	"time"

	"github.com/Ko4etov/gophkeeper/internal/client/config"
	"github.com/Ko4etov/gophkeeper/internal/client/crypto"
	"github.com/Ko4etov/gophkeeper/internal/client/service/storage"
	"github.com/Ko4etov/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestStorage создает временное хранилище для тестов и разблокирует сессию
func createTestStorage(t *testing.T) *storage.Storage {
	session := crypto.NewSession(context.Background(), 15 * time.Minute)

	masterPassword := "test-master-password-123!"
	salt := []byte("test-salt-for-unlock-12345678")

	err := session.Unlock(masterPassword, salt)
	require.NoError(t, err)

	cfg := &config.ClientConfig{
		DataDir: t.TempDir(),
	}

	st, err := storage.NewStorage(cfg, session)
	require.NoError(t, err)

	return st
}

// setupTestUser создает тестового пользователя в хранилище
func setupTestUser(t *testing.T, st *storage.Storage) string {
	email := "test@example.com"

	err := st.EnsureUserBuckets(email)
	require.NoError(t, err)

	user := &models.User{Email: email}
	err = st.SaveUser(user)
	require.NoError(t, err)

	salt := []byte("test-user-salt-1234567890123456")
	err = st.SaveSalt(email, salt)
	require.NoError(t, err)

	_, err = st.GetUser(email)
	require.NoError(t, err)

	_, err = st.GetSalt(email)
	require.NoError(t, err)

	return email
}

func TestNewDataService(t *testing.T) {
	ctx := context.Background()
	testStorage := createTestStorage(t)

	service, err := NewDataService(ctx, nil, testStorage)

	require.NoError(t, err)
	assert.NotNil(t, service)
}

func TestDataService_AddAndGetLogin(t *testing.T) {
	testStorage := createTestStorage(t)
	email := setupTestUser(t, testStorage)

	ctx := context.Background()
	service, _ := NewDataService(ctx, nil, testStorage)

	name := "My Login"
	login := "user@example.com"
	password := "secret123"
	url := "https://example.com"
	notes := "Test notes"
	tags := []string{"work", "important"}

	err := service.AddLogin(name, login, password, url, notes, tags, email)
	require.NoError(t, err)

	metas, err := service.ListMeta(email)
	require.NoError(t, err)
	assert.Len(t, metas, 1)
	assert.Equal(t, name, metas[0].Name)
	assert.Equal(t, models.TypeLoginPassword, metas[0].DataType)

	id := metas[0].ID
	entry, err := service.GetLogin(id, email)
	require.NoError(t, err)
	assert.Equal(t, name, entry.Meta.Name)

	loginData, ok := entry.Data.(models.LoginPasswordData)
	assert.True(t, ok)
	assert.Equal(t, login, loginData.Login)
	assert.Equal(t, password, loginData.Password)
	assert.Equal(t, url, loginData.URL)
	assert.Equal(t, notes, loginData.Notes)
}

func TestDataService_AddAndGetCard(t *testing.T) {
	testStorage := createTestStorage(t)
	email := setupTestUser(t, testStorage)

	ctx := context.Background()
	service, _ := NewDataService(ctx, nil, testStorage)

	err := service.AddCard(
		"My Card",
		"1234567812345678",
		"John Doe",
		12, 2025,
		"123",
		"Visa",
		"Test Bank",
		[]string{"finance"},
		email,
	)
	require.NoError(t, err)

	metas, err := service.ListMeta(email)
	require.NoError(t, err)
	assert.Len(t, metas, 1)

	entry, err := service.GetCard(metas[0].ID, email)
	require.NoError(t, err)

	cardData, ok := entry.Data.(models.BankCardData)
	assert.True(t, ok)
	assert.Equal(t, "1234567812345678", cardData.CardNumber)
	assert.Equal(t, "John Doe", cardData.CardHolder)
	assert.Equal(t, 12, cardData.ExpiryMonth)
	assert.Equal(t, 2025, cardData.ExpiryYear)
	assert.Equal(t, "123", cardData.CVV)
}

func TestDataService_AddAndGetText(t *testing.T) {
	testStorage := createTestStorage(t)
	email := setupTestUser(t, testStorage)

	ctx := context.Background()
	service, _ := NewDataService(ctx, nil, testStorage)

	err := service.AddText(
		"My Note",
		"This is a secret note",
		"text/plain",
		[]string{"personal"},
		email,
	)
	require.NoError(t, err)

	metas, err := service.ListMeta(email)
	require.NoError(t, err)
	assert.Len(t, metas, 1)

	entry, err := service.GetText(metas[0].ID, email)
	require.NoError(t, err)

	textData, ok := entry.Data.(models.TextData)
	assert.True(t, ok)
	assert.Equal(t, "This is a secret note", textData.Content)
	assert.Equal(t, "text/plain", textData.Format)
}

func TestDataService_UpdateLogin(t *testing.T) {
	testStorage := createTestStorage(t)
	email := setupTestUser(t, testStorage)

	ctx := context.Background()
	service, _ := NewDataService(ctx, nil, testStorage)

	err := service.AddLogin("Old Name", "old@example.com", "oldpass", "", "", nil, email)
	require.NoError(t, err)

	metas, _ := service.ListMeta(email)
	id := metas[0].ID

	err = service.UpdateLogin(id, "New Name", "new@example.com", "newpass", "https://new.com", "Updated", []string{"updated"}, email)
	require.NoError(t, err)

	entry, err := service.GetLogin(id, email)
	require.NoError(t, err)
	assert.Equal(t, "New Name", entry.Meta.Name)
	assert.EqualValues(t, 2, entry.Meta.Version)

	loginData, _ := entry.Data.(models.LoginPasswordData)
	assert.Equal(t, "new@example.com", loginData.Login)
	assert.Equal(t, "newpass", loginData.Password)
}

func TestDataService_UpdateText(t *testing.T) {
	testStorage := createTestStorage(t)
	email := setupTestUser(t, testStorage)

	ctx := context.Background()
	service, _ := NewDataService(ctx, nil, testStorage)

	err := service.AddText("Old Title", "Old content", "text/plain", nil, email)
	require.NoError(t, err)

	metas, _ := service.ListMeta(email)
	id := metas[0].ID

	err = service.UpdateText(id, "New Title", "New content", "text/markdown", []string{"new"}, email)
	require.NoError(t, err)

	entry, err := service.GetText(id, email)
	require.NoError(t, err)
	assert.Equal(t, "New Title", entry.Meta.Name)

	textData, _ := entry.Data.(models.TextData)
	assert.Equal(t, "New content", textData.Content)
	assert.Equal(t, "text/markdown", textData.Format)
}

func TestDataService_Delete(t *testing.T) {
	testStorage := createTestStorage(t)
	email := setupTestUser(t, testStorage)

	ctx := context.Background()
	service, _ := NewDataService(ctx, nil, testStorage)

	err := service.AddText("To Delete", "content", "text/plain", nil, email)
	require.NoError(t, err)

	metas, err := service.ListMeta(email)
	require.NoError(t, err)
	assert.Len(t, metas, 1)
	id := metas[0].ID

	err = service.Delete(id, email)
	require.NoError(t, err)

	_, err = service.Get(id, email)
	assert.Error(t, err)
}

func TestDataService_WrongDataType(t *testing.T) {
	testStorage := createTestStorage(t)
	email := setupTestUser(t, testStorage)

	ctx := context.Background()
	service, _ := NewDataService(ctx, nil, testStorage)

	err := service.AddText("Text Entry", "content", "text/plain", nil, email)
	require.NoError(t, err)

	metas, _ := service.ListMeta(email)
	id := metas[0].ID

	_, err = service.GetLogin(id, email)
	assert.ErrorIs(t, err, ErrWrongDataType)

	_, err = service.GetCard(id, email)
	assert.ErrorIs(t, err, ErrWrongDataType)

	_, err = service.GetBinary(id, email)
	assert.ErrorIs(t, err, ErrWrongDataType)
}

// Дополнительный тест для проверки сохранения соли
func TestDataService_SaltPersistence(t *testing.T) {
	testStorage := createTestStorage(t)
	email := "test-salt@example.com"

	err := testStorage.EnsureUserBuckets(email)
	require.NoError(t, err)

	user := &models.User{Email: email}
	err = testStorage.SaveUser(user)
	require.NoError(t, err)

	originalSalt := []byte("my-test-salt-1234567890123456")
	err = testStorage.SaveSalt(email, originalSalt)
	require.NoError(t, err)

	retrievedSalt, err := testStorage.GetSalt(email)
	require.NoError(t, err)
	assert.Equal(t, originalSalt, retrievedSalt)
}
