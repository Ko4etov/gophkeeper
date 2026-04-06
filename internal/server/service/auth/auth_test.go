// internal/server/auth/service_test.go
package auth

import (
	"context"
	"testing"
	"time"

	authproto "github.com/Ko4etov/gophkeeper/internal/proto/auth"
	"github.com/Ko4etov/gophkeeper/internal/server/auth"
	"github.com/Ko4etov/gophkeeper/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockUserStorage - мок для storage.Storage
type MockUserStorage struct {
	mock.Mock
}

func (m *MockUserStorage) GetUserByEmail(ctx context.Context, email string) (*storage.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.User), args.Error(1)
}

func (m *MockUserStorage) GetUserByID(ctx context.Context, id string) (*storage.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.User), args.Error(1)
}

func (m *MockUserStorage) CreateUser(ctx context.Context, email, password string) (*storage.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.User), args.Error(1)
}

func (m *MockUserStorage) CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, token, expiresAt)
	return args.Error(0)
}

func (m *MockUserStorage) RevokeAllUserTokens(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserStorage) GetUserSalt(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(0)
}

func (m *MockUserStorage) SaveUserSalt(ctx context.Context, userID string, salt string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// createRealJWTManager создает реальный JWTManager для тестов
func createRealJWTManager() *auth.JWTManager {
	config := &auth.JWTConfig{
		SecretKey:  "test-secret-key",
		AccessTTL:  15, // 15 минут
		RefreshTTL: 24, // 24 часа
	}
	return auth.NewJWTManager(config)
}

// Тесты
func TestService_Register(t *testing.T) {
	mockStorage := new(MockUserStorage)
	realJWT := createRealJWTManager()
	service := NewService(mockStorage, realJWT)

	tests := []struct {
		name    string
		req     *authproto.RegisterRequest
		setup   func()
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &authproto.RegisterRequest{
				Email:    "test@example.com",
				Password: "Password123!",
			},
			setup: func() {
				mockStorage.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(nil, nil).Once()
				mockStorage.On("CreateUser", mock.Anything, "test@example.com", "Password123!").
					Return(&storage.User{ID: "user-123", Email: "test@example.com"}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "empty email",
			req: &authproto.RegisterRequest{
				Email:    "",
				Password: "Password123!",
			},
			setup:   func() {},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "empty password",
			req: &authproto.RegisterRequest{
				Email:    "test@example.com",
				Password: "",
			},
			setup:   func() {},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "weak password",
			req: &authproto.RegisterRequest{
				Email:    "test@example.com",
				Password: "weak",
			},
			setup:   func() {},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "user already exists",
			req: &authproto.RegisterRequest{
				Email:    "test@example.com",
				Password: "Password123!",
			},
			setup: func() {
				mockStorage.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(&storage.User{ID: "user-123", Email: "test@example.com"}, nil).Once()
			},
			wantErr: true,
			errCode: codes.AlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := service.Register(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "User registered successfully", resp.Message)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestService_Login(t *testing.T) {
	mockStorage := new(MockUserStorage)
	realJWT := createRealJWTManager()
	service := NewService(mockStorage, realJWT)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)

	tests := []struct {
		name    string
		req     *authproto.LoginRequest
		setup   func()
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &authproto.LoginRequest{
				Email:    "test@example.com",
				Password: "Password123!",
			},
			setup: func() {
				mockStorage.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(&storage.User{
						ID:           "user-123",
						Email:        "test@example.com",
						PasswordHash: string(hashedPassword),
					}, nil).Once()
				mockStorage.On("CreateRefreshToken", mock.Anything, "user-123", mock.Anything, mock.Anything).
					Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "user not found",
			req: &authproto.LoginRequest{
				Email:    "test@example.com",
				Password: "Password123!",
			},
			setup: func() {
				mockStorage.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(nil, nil).Once()
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "invalid password",
			req: &authproto.LoginRequest{
				Email:    "test@example.com",
				Password: "WrongPassword!",
			},
			setup: func() {
				mockStorage.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(&storage.User{
						ID:           "user-123",
						Email:        "test@example.com",
						PasswordHash: string(hashedPassword),
					}, nil).Once()
			},
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := service.Login(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
			} else {
				assert.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "user-123", resp.UserId)
				assert.Equal(t, "test@example.com", resp.Email)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
				assert.Equal(t, int64(realJWT.AccessTTL.Seconds()), resp.ExpiresIn)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestService_RefreshToken(t *testing.T) {
	mockStorage := new(MockUserStorage)
	realJWT := createRealJWTManager()
	service := NewService(mockStorage, realJWT)

	// Создаем валидный refresh токен
	validRefreshToken, err := realJWT.GenerateRefreshToken("user-123")
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     *authproto.RefreshTokenRequest
		setup   func()
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &authproto.RefreshTokenRequest{
				RefreshToken: validRefreshToken,
			},
			setup: func() {
				mockStorage.On("GetUserByID", mock.Anything, "user-123").
					Return(&storage.User{ID: "user-123", Email: "test@example.com"}, nil).Once()
				mockStorage.On("CreateRefreshToken", mock.Anything, "user-123", mock.Anything, mock.Anything).
					Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "invalid refresh token",
			req: &authproto.RefreshTokenRequest{
				RefreshToken: "invalid-token",
			},
			setup:   func() {},
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
		{
			name: "user not found",
			req: &authproto.RefreshTokenRequest{
				RefreshToken: validRefreshToken,
			},
			setup: func() {
				mockStorage.On("GetUserByID", mock.Anything, "user-123").
					Return(nil, nil).Once()
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := service.RefreshToken(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
			} else {
				assert.NoError(t, err)
				require.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
				assert.Equal(t, int64(realJWT.AccessTTL.Seconds()), resp.ExpiresIn)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestService_Logout(t *testing.T) {
	mockStorage := new(MockUserStorage)
	realJWT := createRealJWTManager()
	service := NewService(mockStorage, realJWT)

	// Создаем валидный access токен
	validAccessToken, err := realJWT.GenerateAccessToken("user-123", "test@example.com")
	require.NoError(t, err)

	tests := []struct {
		name        string
		req         *authproto.LogoutRequest
		setup       func()
		wantErr     bool
		wantSuccess bool
	}{
		{
			name: "success",
			req: &authproto.LogoutRequest{
				AccessToken: validAccessToken,
			},
			setup: func() {
				mockStorage.On("RevokeAllUserTokens", mock.Anything, "user-123").
					Return(nil).Once()
			},
			wantErr:     false,
			wantSuccess: true,
		},
		{
			name: "invalid token - returns success false",
			req: &authproto.LogoutRequest{
				AccessToken: "invalid-token",
			},
			setup:       func() {},
			wantErr:     false,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := service.Logout(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.wantSuccess, resp.Success)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestService_ValidateToken(t *testing.T) {
	mockStorage := new(MockUserStorage)
	realJWT := createRealJWTManager()
	service := NewService(mockStorage, realJWT)

	// Создаем валидный access токен
	validAccessToken, err := realJWT.GenerateAccessToken("user-123", "test@example.com")
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     *authproto.ValidateTokenRequest
		valid   bool
		userID  string
		email   string
	}{
		{
			name: "valid token",
			req: &authproto.ValidateTokenRequest{
				AccessToken: validAccessToken,
			},
			valid:  true,
			userID: "user-123",
			email:  "test@example.com",
		},
		{
			name: "invalid token",
			req: &authproto.ValidateTokenRequest{
				AccessToken: "invalid-token",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.ValidateToken(context.Background(), tt.req)

			assert.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.valid, resp.Valid)
			if tt.valid {
				assert.Equal(t, tt.userID, resp.UserId)
				assert.Equal(t, tt.email, resp.Email)
			}
		})
	}
}

// Тест для проверки истекшего токена
func TestService_ValidateToken_Expired(t *testing.T) {
	mockStorage := new(MockUserStorage)
	
	// Создаем JWTManager с истекшим временем
	config := &auth.JWTConfig{
		SecretKey:  "test-secret-key",
		AccessTTL:  0, // 0 минут - токен истекает сразу
		RefreshTTL: 24,
	}
	expiredJWT := auth.NewJWTManager(config)
	service := NewService(mockStorage, expiredJWT)

	// Генерируем истекший токен
	expiredToken, err := expiredJWT.GenerateAccessToken("user-123", "test@example.com")
	require.NoError(t, err)

	// Небольшая задержка для гарантии истечения
	time.Sleep(1 * time.Millisecond)

	resp, err := service.ValidateToken(context.Background(), &authproto.ValidateTokenRequest{
		AccessToken: expiredToken,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Valid)
}