// internal/server/auth/service.go
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	authproto "github.com/Ko4etov/gophkeeper/internal/proto/auth"
	"github.com/Ko4etov/gophkeeper/internal/server/auth"
	"github.com/Ko4etov/gophkeeper/internal/server/storage"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
    authproto.UnimplementedAuthServiceServer
    storage     *storage.Storage
    jwtManager  *auth.JWTManager
}

func NewService(storage *storage.Storage, jwtManager *auth.JWTManager) *Service {
    return &Service{
        storage:    storage,
        jwtManager: jwtManager,
    }
}

func (s *Service) Register(ctx context.Context, req *authproto.RegisterRequest) (*authproto.RegisterResponse, error) {

    // Проверяем, существует ли пользователь
    existing, err := s.storage.GetUserByEmail(ctx, req.Email)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "database error: %v", err)
    }
    if existing != nil {
        return nil, status.Error(codes.AlreadyExists, "user already exists")
    }

    // Создаем пользователя
    user, err := s.storage.CreateUser(ctx, req.Email, req.Password)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
    }

    return &authproto.RegisterResponse{
        UserId:  user.ID,
        Message: "User registered successfully",
    }, nil
}

func (s *Service) Login(ctx context.Context, req *authproto.LoginRequest) (*authproto.LoginResponse, error) {
    // Получаем пользователя
    user, err := s.storage.GetUserByEmail(ctx, req.Email)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "database error: %v", err)
    }
    if user == nil {
        return nil, status.Error(codes.NotFound, "user not found")
    }

    // Проверяем пароль
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        return nil, status.Error(codes.Unauthenticated, "invalid credentials")
    }

    // Генерируем токены
    accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
    }

    refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
    }

    // Сохраняем refresh token в БД
    err = s.storage.CreateRefreshToken(ctx, user.ID, refreshToken, time.Now().Add(s.jwtManager.RefreshTTL))
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to save refresh token: %v", err)
    }

    return &authproto.LoginResponse{
        UserId:       user.ID,
        Email:        user.Email,
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    int64(s.jwtManager.AccessTTL.Seconds()),
    }, nil
}

func (s *Service) RefreshToken(ctx context.Context, req *authproto.RefreshTokenRequest) (*authproto.RefreshTokenResponse, error) {
    // Валидируем refresh token
    claims, err := s.jwtManager.ValidateToken(req.RefreshToken)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
    }

    // Получаем пользователя
    user, err := s.storage.GetUserByID(ctx, claims.Subject)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "database error: %v", err)
    }
    if user == nil {
        return nil, status.Error(codes.NotFound, "user not found")
    }

    // Генерируем новые токены
    accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
    }

    refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
    }

    // Сохраняем новый refresh token
    err = s.storage.CreateRefreshToken(ctx, user.ID, refreshToken, time.Now().Add(s.jwtManager.RefreshTTL))
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to save refresh token: %v", err)
    }

    return &authproto.RefreshTokenResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    int64(s.jwtManager.AccessTTL.Seconds()),
    }, nil
}

func (s *Service) Logout(ctx context.Context, req *authproto.LogoutRequest) (*authproto.LogoutResponse, error) {
    // Валидируем токен
    claims, err := s.jwtManager.ValidateToken(req.AccessToken)
    if err != nil {
        return &authproto.LogoutResponse{Success: false}, nil
    }

    // Отзываем все refresh токены пользователя
    err = s.storage.RevokeAllUserTokens(ctx, claims.Subject)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to revoke tokens: %v", err)
    }

    return &authproto.LogoutResponse{Success: true}, nil
}

func (s *Service) ValidateToken(ctx context.Context, req *authproto.ValidateTokenRequest) (*authproto.ValidateTokenResponse, error) {
    claims, err := s.jwtManager.ValidateToken(req.AccessToken)
    if err != nil {
        return &authproto.ValidateTokenResponse{Valid: false}, nil
    }

    return &authproto.ValidateTokenResponse{
        Valid:  true,
        UserId: claims.UserID,
        Email:  claims.Email,
    }, nil
}

func generateRandomToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return hex.EncodeToString(b)
}