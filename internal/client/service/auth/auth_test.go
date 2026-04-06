// internal/client/service/auth/auth_test.go
package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Ko4etov/gophkeeper/internal/client/config"
	"github.com/Ko4etov/gophkeeper/internal/client/crypto"
	"github.com/Ko4etov/gophkeeper/internal/client/service/storage"
	"github.com/Ko4etov/gophkeeper/internal/models"
	"github.com/Ko4etov/gophkeeper/internal/proto/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockGrpcClient - мок для grpc клиента
type MockGrpcClient struct {
	mock.Mock
}

func (m *MockGrpcClient) Login(email, password string) (*auth.LoginResponse, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.LoginResponse), args.Error(1)
}

func (m *MockGrpcClient) Register(email, password string) error {
	args := m.Called(email, password)
	return args.Error(0)
}

func (m *MockGrpcClient) RefreshToken(accessToken, refreshToken string) (*auth.RefreshTokenResponse, error) {
	args := m.Called(accessToken, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.RefreshTokenResponse), args.Error(1)
}

// createTestStorage создает временное хранилище для тестов
func createTestStorage(t *testing.T) *storage.Storage {
	session := crypto.NewSession(context.Background(), 15 * time.Minute)
	
	// Создаем тестовую директорию
	config := &config.ClientConfig{
		DataDir: t.TempDir(),
	}
	
	st, err := storage.NewStorage(config, session)
	require.NoError(t, err)
	
	return st
}

func TestNewAuthService(t *testing.T) {
	ctx := context.Background()
	mockGrpc := new(MockGrpcClient)
	testStorage := createTestStorage(t)

	service, err := NewAuthService(ctx, mockGrpc, testStorage)

	require.NoError(t, err)
	assert.NotNil(t, service)
}

func TestAuthService_Register(t *testing.T) {
	mockGrpc := new(MockGrpcClient)
	testStorage := createTestStorage(t)
	ctx := context.Background()

	service, _ := NewAuthService(ctx, mockGrpc, testStorage)

	tests := []struct {
		name     string
		email    string
		password string
		setup    func()
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "success",
			email:    "test@example.com",
			password: "Password123!",
			setup: func() {
				mockGrpc.On("Register", "test@example.com", "Password123!").Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:     "empty email",
			email:    "",
			password: "Password123!",
			setup:    func() {},
			wantErr:  true,
			errMsg:   "email cannot be empty",
		},
		{
			name:     "empty password",
			email:    "test@example.com",
			password: "",
			setup:    func() {},
			wantErr:  true,
			errMsg:   "password cannot be empty",
		},
		{
			name:     "weak password",
			email:    "test@example.com",
			password: "weak",
			setup:    func() {},
			wantErr:  true,
			errMsg:   "weak password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.Register(tt.email, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
			mockGrpc.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	mockGrpc := new(MockGrpcClient)
	testStorage := createTestStorage(t)
	ctx := context.Background()

	service, _ := NewAuthService(ctx, mockGrpc, testStorage)

	loginResp := &auth.LoginResponse{
		UserId:       "user-123",
		Email:        "test@example.com",
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    900,
	}

	tests := []struct {
		name     string
		email    string
		password string
		setup    func()
		wantErr  bool
	}{
		{
			name:     "success",
			email:    "test@example.com",
			password: "Password123!",
			setup: func() {
				mockGrpc.On("Login", "test@example.com", "Password123!").Return(loginResp, nil).Once()
			},
			wantErr: false,
		},
		{
			name:     "login failed",
			email:    "test@example.com",
			password: "wrong",
			setup: func() {
				mockGrpc.On("Login", "test@example.com", "wrong").Return(nil, errors.New("invalid credentials")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			user, err := service.Login(tt.email, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, "user-123", user.ID)
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, "access-token", user.Token)
			}
			mockGrpc.AssertExpectations(t)
		})
	}
}

func TestAuthService_GetUser_SetUser(t *testing.T) {
	mockGrpc := new(MockGrpcClient)
	testStorage := createTestStorage(t)
	ctx := context.Background()

	service, _ := NewAuthService(ctx, mockGrpc, testStorage)

	// Изначально nil
	assert.Nil(t, service.GetUser())

	// Устанавливаем пользователя
	user := &models.User{
		ID:    "user-123",
		Email: "test@example.com",
	}
	service.SetUser(user)

	// Проверяем
	assert.Equal(t, user, service.GetUser())
}