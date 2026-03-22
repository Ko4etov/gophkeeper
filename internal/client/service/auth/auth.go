// Package auth предоставляет сервис аутентификации клиента GophKeeper.
package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	grpcclient "github.com/Ko4etov/gophkeeper/internal/client/grpc_client"
	"github.com/Ko4etov/gophkeeper/internal/client/service/storage"
	"github.com/Ko4etov/gophkeeper/internal/models"
)

// AuthService управляет аутентификацией пользователя и обновлением токенов.
type AuthService struct {
	user         *models.User          // текущий пользователь
	grpcclient   *grpcclient.GrpcClient
	storage      *storage.Storage
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	refreshTimer *time.Timer           // таймер для автоматического обновления токена
	refreshDone  chan struct{}          // канал для остановки таймера
}

// NewAuthService создает новый сервис аутентификации.
func NewAuthService(ctx context.Context, grpcclient *grpcclient.GrpcClient, storage *storage.Storage) (*AuthService, error) {
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

	a.SetUser(&models.User{
		ID:           resp.UserId,
		Email:        resp.Email,
		Token:        resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	})

	user := a.GetUser()

	if err := a.storage.EnsureUserBuckets(user.Email); err != nil {
		return nil, err
	}

	if err := a.storage.SaveUser(user); err != nil {
		return nil, err
	}

	a.StartAccesTokenRefreshing()

	return a.user, nil
}

// Register создает нового пользователя на сервере.
func (a *AuthService) Register(email, password string) error {
	if a.user != nil {
		return fmt.Errorf("already logged in as %s", a.user.Email)
	}

	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	return a.grpcclient.Register(email, password)
}

// StartAccesTokenRefreshing запускает автоматическое обновление токена.
func (a *AuthService) StartAccesTokenRefreshing() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.user == nil {
		return
	}

	a.stopRefreshTimer()

	timeUntilExpiry := time.Until(a.user.ExpiresAt)

	if timeUntilExpiry <= 0 {
		a.user = nil
		return
	}

	// Обновляем за 1 минуту до истечения
	refreshIn := timeUntilExpiry - 1*time.Minute
	if refreshIn < 0 {
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
		fmt.Printf("Failed to refresh token: %v\n", err)
		a.mu.Lock()
		a.user = nil
		a.mu.Unlock()
		return
	}

	a.mu.Lock()
	if a.user != nil {
		a.user.Token = resp.AccessToken
		a.user.RefreshToken = resp.RefreshToken
		a.user.ExpiresAt = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	}
	a.mu.Unlock()

	a.scheduleNextRefresh()
}

// scheduleNextRefresh планирует следующее обновление токена.
func (a *AuthService) scheduleNextRefresh() {
	if a.user == nil {
		return
	}

	timeUntilExpiry := time.Until(a.user.ExpiresAt)
	refreshIn := timeUntilExpiry - 1*time.Minute

	if refreshIn <= 0 {
		go a.refreshToken()
		return
	}

	if a.refreshTimer != nil {
		a.refreshTimer.Stop()
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

// stopRefreshTimer останавливает таймер обновления.
func (a *AuthService) stopRefreshTimer() {
	if a.refreshTimer != nil {
		a.refreshTimer.Stop()
		a.refreshTimer = nil
	}
	if a.refreshDone != nil {
		close(a.refreshDone)
		a.refreshDone = nil
	}
}