// Package auth предоставляет сервис аутентификации клиента GophKeeper.
package auth

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Ko4etov/gophkeeper/internal/client/service/storage"
	"github.com/Ko4etov/gophkeeper/internal/common/auth"
	"github.com/Ko4etov/gophkeeper/internal/models"
	protoauth "github.com/Ko4etov/gophkeeper/internal/proto/auth"
)

// AuthService управляет аутентификацией пользователя и обновлением токенов.
type AuthService struct {
	user         *models.User          // текущий пользователь
	grpcclient   GrpcClientInterface
	storage      *storage.Storage
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	refreshTimer *time.Timer           // таймер для автоматического обновления токена
	refreshDone  chan struct{}          // канал для остановки таймера
}

type GrpcClientInterface interface {
	Login(email, password string) (*protoauth.LoginResponse, error)
	Register(email, password string) error
	RefreshToken(accessToken, refreshToken string) (*protoauth.RefreshTokenResponse, error)
}

// NewAuthService создает новый сервис аутентификации.
func NewAuthService(ctx context.Context, grpcclient GrpcClientInterface, storage *storage.Storage) (*AuthService, error) {
	serviceCtx, cancel := context.WithCancel(ctx)

	return &AuthService{
		grpcclient:  grpcclient,
		storage:     storage,
		ctx:         serviceCtx,
		refreshDone: make(chan struct{}),
		cancel:      cancel,
	}, nil
}

// GetUser возвращает текущего авторизованного пользователя.
func (a *AuthService) GetUser() *models.User {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.user
}

// SetUser устанавливает текущего пользователя.
func (a *AuthService) SetUser(user *models.User) *models.User {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.user = user
	return a.user
}

// Login выполняет вход пользователя на сервере.
func (a *AuthService) Login(email string, password string) (*models.User, error) {
	resp, err := a.grpcclient.Login(email, password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           resp.UserId,
		Email:        resp.Email,
		Token:        resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}

	a.SetUser(user)

	if err := a.storage.EnsureUserBuckets(user.Email); err != nil {
		return nil, err
	}

	if err := a.storage.SaveUser(user); err != nil {
		return nil, err
	}

	a.StartAccesTokenRefreshing()

	return user, nil
}

// Register создает нового пользователя на сервере.
func (a *AuthService) Register(email, password string) error {
	if user := a.GetUser(); user != nil {
		return fmt.Errorf("already logged in as %s", user.Email)
	}

	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if err := auth.ValidatePasswordStrength(password); err != nil {
		return fmt.Errorf("weak password: %v", err)
	}

	return a.grpcclient.Register(email, password)
}

// StartAccesTokenRefreshing запускает автоматическое обновление токена.
func (a *AuthService) StartAccesTokenRefreshing() {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.scheduleRefreshLocked()
}

// refreshToken выполняет обновление токена.
func (a *AuthService) refreshToken() {
	a.mu.Lock()
	if a.user == nil {
		a.mu.Unlock()
		return
	}
	refreshToken := a.user.RefreshToken
	accessToken := a.user.Token
	a.mu.Unlock()

	resp, err := a.grpcclient.RefreshToken(accessToken, refreshToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to refresh token: %v\n", err)
		a.mu.Lock()
		a.user = nil
		a.mu.Unlock()
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.user != nil {
		a.user.Token = resp.AccessToken
		a.user.RefreshToken = resp.RefreshToken
		a.user.ExpiresAt = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	}
	
	a.scheduleRefreshLocked()
}

// scheduleRefreshLocked планирует следующее обновление токена.
func (a *AuthService) scheduleRefreshLocked() {
	if a.user == nil {
		return
	}

	a.stopRefreshTimerLocked()

	timeUntilExpiry := time.Until(a.user.ExpiresAt)
	if timeUntilExpiry <= 0 {
		a.user = nil
		return
	}

	// Обновляем за 1 минуту до истечения
	refreshIn := timeUntilExpiry - 1*time.Minute
	if refreshIn <= 0 {
		go a.refreshToken()
		return
	}

	a.refreshDone = make(chan struct{})
	a.refreshTimer = time.AfterFunc(refreshIn, func() {
		select {
		case <-a.refreshDone:
			return
		default:
			a.refreshToken()
		}
	})
}

// stopRefreshTimerLocked останавливает таймер обновления.
func (a *AuthService) stopRefreshTimerLocked() {
	if a.refreshTimer != nil {
		a.refreshTimer.Stop()
		a.refreshTimer = nil
	}
	if a.refreshDone != nil {
		close(a.refreshDone)
		a.refreshDone = nil
	}
}