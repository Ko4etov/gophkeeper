// Package crypto предоставляет функции для шифрования данных клиента.
package crypto

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Session управляет сессией разблокированного хранилища.
type Session struct {
	mu           sync.RWMutex
	encryptor    *Encryptor
	lastActivity time.Time
	timeout      time.Duration
	stopAutoLock chan struct{}
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewSession создает новую сессию с указанным таймаутом бездействия.
func NewSession(ctx context.Context, timeout time.Duration) *Session {
	sessionCtx, cancel := context.WithCancel(ctx)
	
	s := &Session{
		timeout:      timeout,
		stopAutoLock: make(chan struct{}),
		ctx:          sessionCtx,
		cancel:       cancel,
	}
	s.startAutoLock()
	return s
}

// startAutoLock запускает фоновую горутину для автоматической блокировки.
func (s *Session) startAutoLock() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.timeout / 2)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkAndLock()
			case <-s.stopAutoLock:
				return
			case <-s.ctx.Done():
				s.Lock()
				return
			}
		}
	}()
}

// checkAndLock проверяет таймаут и блокирует сессию при необходимости.
func (s *Session) checkAndLock() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.encryptor == nil {
		return
	}

	if time.Since(s.lastActivity) > s.timeout {
		s.encryptor = nil
	}
}

// Unlock разблокирует сессию, проверяя мастер-пароль.
func (s *Session) Unlock(masterPassword string, salt []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	encryptor := NewEncryptor(masterPassword, salt)

	if err := encryptor.Verify(); err != nil {
		return err
	}

	s.encryptor = encryptor
	s.lastActivity = time.Now()
	return nil
}

// IsUnlocked проверяет, разблокирована ли сессия и не истек ли таймаут.
func (s *Session) IsUnlocked() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.encryptor == nil {
		return false
	}

	if time.Since(s.lastActivity) > s.timeout {
		s.encryptor = nil
		return false
	}

	return true
}

// Lock блокирует сессию, удаляя шифровальщик из памяти.
func (s *Session) Lock() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.encryptor = nil
}

// Touch обновляет время последней активности, продлевая сессию.
func (s *Session) Touch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActivity = time.Now()
}

// GetEncryptor возвращает шифровальщик, если сессия разблокирована.
func (s *Session) GetEncryptor() (*Encryptor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.encryptor == nil {
		return nil, ErrLocked
	}

	if time.Since(s.lastActivity) > s.timeout {
		s.encryptor = nil
		return nil, ErrLocked
	}

	return s.encryptor, nil
}

// Stop останавливает фоновую горутину автоматической блокировки.
func (s *Session) Stop() {
	close(s.stopAutoLock)
	s.cancel()
	s.wg.Wait()
}

// ErrLocked возникает при попытке доступа к заблокированному хранилищу.
var ErrLocked = errors.New("storage is locked")