package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateAndValidate(t *testing.T) {
	m := NewJWTManager(&JWTConfig{
		SecretKey:  "secret",
		AccessTTL:  15,
		RefreshTTL: 24,
	})

	// Тест access токена
	accessToken, err := m.GenerateAccessToken("user123", "test@example.com")
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	claims, err := m.ValidateToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)

	// Тест refresh токена
	refreshToken, err := m.GenerateRefreshToken("user123")
	require.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	refreshClaims, err := m.ValidateToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, "user123", refreshClaims.Subject)
}

func TestJWTManager_InvalidToken(t *testing.T) {
	m := NewJWTManager(&JWTConfig{
		SecretKey:  "secret",
		AccessTTL:  15,
		RefreshTTL: 24,
	})

	claims, err := m.ValidateToken("invalid-token")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_WrongSecret(t *testing.T) {
	m1 := NewJWTManager(&JWTConfig{SecretKey: "secret-1", AccessTTL: 15, RefreshTTL: 24})
	m2 := NewJWTManager(&JWTConfig{SecretKey: "secret-2", AccessTTL: 15, RefreshTTL: 24})

	token, err := m1.GenerateAccessToken("user123", "test@example.com")
	require.NoError(t, err)

	claims, err := m2.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	m := NewJWTManager(&JWTConfig{
		SecretKey:  "secret",
		AccessTTL:  0, // истекает сразу
		RefreshTTL: 24,
	})

	token, err := m.GenerateAccessToken("user123", "test@example.com")
	require.NoError(t, err)

	time.Sleep(1 * time.Millisecond)

	claims, err := m.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}