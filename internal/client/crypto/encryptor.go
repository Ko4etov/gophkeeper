// Package crypto предоставляет функции для шифрования данных клиента.
// Использует AES-256-GCM для шифрования и PBKDF2 для получения ключа.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// Encryptor управляет шифрованием данных.
type Encryptor struct {
	key []byte
}

// NewEncryptor создает шифровальщик из мастер-пароля и соли через PBKDF2.
func NewEncryptor(masterPassword string, salt []byte) *Encryptor {
	key := pbkdf2.Key([]byte(masterPassword), salt, 100000, 32, sha256.New)
	return &Encryptor{key: key}
}

// Encrypt шифрует данные алгоритмом AES-256-GCM и возвращает base64-строку.
func (e *Encryptor) Encrypt(plaintext []byte) (string, error) {
	if e.key == nil {
		return "", errors.New("encryptor not initialized")
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt расшифровывает данные из base64-строки.
func (e *Encryptor) Decrypt(encryptedData string) ([]byte, error) {
	if e.key == nil {
		return nil, errors.New("encryptor not initialized")
	}

	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// Verify проверяет корректность мастер-пароля.
func (e *Encryptor) Verify() error {
	testData := []byte("test")
	encrypted, err := e.Encrypt(testData)
	if err != nil {
		return err
	}

	decrypted, err := e.Decrypt(encrypted)
	if err != nil {
		return err
	}

	if string(decrypted) != string(testData) {
		return errors.New("verification failed")
	}

	return nil
}

// GenerateSalt генерирует случайную соль длиной 32 байта.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}