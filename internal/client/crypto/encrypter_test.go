// internal/client/crypto/crypto_test.go
package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	require.NoError(t, err)
	assert.Len(t, salt1, 32)

	salt2, err := GenerateSalt()
	require.NoError(t, err)
	assert.NotEqual(t, salt1, salt2)
}

func TestNewEncryptor(t *testing.T) {
	salt, _ := GenerateSalt()
	encryptor := NewEncryptor("password", salt)
	assert.NotNil(t, encryptor)
	assert.Len(t, encryptor.key, 32)
}

func TestEncryptor_EncryptDecrypt(t *testing.T) {
	salt, _ := GenerateSalt()
	encryptor := NewEncryptor("my-secret-password", salt)

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"simple text", []byte("hello world")},
		{"empty data", []byte{}},
		{"long text", []byte("Lorem ipsum dolor sit amet")},
		{"binary data", []byte{0x00, 0x01, 0x02, 0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := encryptor.Encrypt(tt.plaintext)
			require.NoError(t, err)

			decrypted, err := encryptor.Decrypt(encrypted)
			require.NoError(t, err)
			
			if !bytes.Equal(tt.plaintext, decrypted) {
				t.Errorf("expected %v, got %v", tt.plaintext, decrypted)
			}
		})
	}
}

func TestEncryptor_NotInitialized(t *testing.T) {
	encryptor := &Encryptor{key: nil}
	
	_, err := encryptor.Encrypt([]byte("test"))
	assert.EqualError(t, err, "encryptor not initialized")
	
	_, err = encryptor.Decrypt("test")
	assert.EqualError(t, err, "encryptor not initialized")
	
	err = encryptor.Verify()
	assert.EqualError(t, err, "encryptor not initialized")
}

func TestEncryptor_Decrypt_InvalidBase64(t *testing.T) {
	salt, _ := GenerateSalt()
	encryptor := NewEncryptor("password", salt)
	
	_, err := encryptor.Decrypt("not-valid-base64!!!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode base64")
}

func TestEncryptor_Decrypt_InvalidCiphertext(t *testing.T) {
	salt, _ := GenerateSalt()
	encryptor := NewEncryptor("password", salt)
	
	_, err := encryptor.Decrypt("dGVzdA==")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext too short")
}

func TestEncryptor_Verify(t *testing.T) {
	salt, _ := GenerateSalt()
	encryptor := NewEncryptor("correct-password", salt)
	
	err := encryptor.Verify()
	assert.NoError(t, err)
}

func TestEncryptor_Verify_WrongPassword(t *testing.T) {
	salt, _ := GenerateSalt()
	encryptor1 := NewEncryptor("correct-password", salt)
	encryptor2 := NewEncryptor("wrong-password", salt)
	
	encrypted, err := encryptor1.Encrypt([]byte("test"))
	require.NoError(t, err)
	
	_, err = encryptor2.Decrypt(encrypted)
	assert.Error(t, err)
}

func TestEncryptor_DifferentNonce(t *testing.T) {
	salt, _ := GenerateSalt()
	encryptor := NewEncryptor("password", salt)
	
	plaintext := []byte("same text")
	
	encrypted1, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)
	
	encrypted2, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)
	
	assert.NotEqual(t, encrypted1, encrypted2)
	
	decrypted1, _ := encryptor.Decrypt(encrypted1)
	decrypted2, _ := encryptor.Decrypt(encrypted2)
	assert.Equal(t, plaintext, decrypted1)
	assert.Equal(t, plaintext, decrypted2)
}

func TestEncryptor_KeyDerivation(t *testing.T) {
	salt, _ := GenerateSalt()
	
	encryptor1 := NewEncryptor("password", salt)
	encryptor2 := NewEncryptor("password", salt)
	assert.Equal(t, encryptor1.key, encryptor2.key)
	
	encryptor3 := NewEncryptor("different", salt)
	assert.NotEqual(t, encryptor1.key, encryptor3.key)
}