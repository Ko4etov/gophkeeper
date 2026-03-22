// Package crypto предоставляет функции для шифрования данных клиента.
package crypto

import (
	"errors"
	"sync"
	"time"
)

// Session управляет сессией разблокированного хранилища.
// Хранит шифровальщик в памяти и автоматически блокируется после таймаута.
type Session struct {
	mu           sync.RWMutex
	encryptor    *Encryptor
	lastActivity time.Time
	timeout      time.Duration
}

// NewSession создает новую сессию с указанным таймаутом бездействия.
func NewSession(timeout time.Duration) *Session {
	return &Session{
		timeout: timeout,
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
		go s.Lock()
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
		return nil, ErrLocked
	}

	return s.encryptor, nil
}

// ErrLocked возникает при попытке доступа к заблокированному хранилищу.
var ErrLocked = errors.New("storage is locked")
