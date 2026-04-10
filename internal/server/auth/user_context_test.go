package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithUserContext(t *testing.T) {
	ctx := context.Background()
	
	claims := &Claims{
		UserID: "user-123",
		Email:  "test@example.com",
	}
	
	newCtx := WithUserContext(ctx, claims)
	
	assert.NotEqual(t, ctx, newCtx, "Context should be different")
}

func TestGetUserID(t *testing.T) {
	t.Run("user ID exists", func(t *testing.T) {
		ctx := context.Background()
		claims := &Claims{UserID: "user-123", Email: "test@example.com"}
		ctx = WithUserContext(ctx, claims)
		
		userID, ok := GetUserID(ctx)
		
		assert.True(t, ok)
		assert.Equal(t, "user-123", userID)
	})
	
	t.Run("user ID does not exist", func(t *testing.T) {
		ctx := context.Background()
		
		userID, ok := GetUserID(ctx)
		
		assert.False(t, ok)
		assert.Empty(t, userID)
	})
}

func TestGetEmail(t *testing.T) {
	t.Run("email exists", func(t *testing.T) {
		ctx := context.Background()
		claims := &Claims{UserID: "user-123", Email: "test@example.com"}
		ctx = WithUserContext(ctx, claims)
		
		email, ok := GetEmail(ctx)
		
		assert.True(t, ok)
		assert.Equal(t, "test@example.com", email)
	})
	
	t.Run("email does not exist", func(t *testing.T) {
		ctx := context.Background()
		
		email, ok := GetEmail(ctx)
		
		assert.False(t, ok)
		assert.Empty(t, email)
	})
}

func TestGetClaims(t *testing.T) {
	t.Run("claims exist", func(t *testing.T) {
		ctx := context.Background()
		originalClaims := &Claims{
			UserID: "user-123",
			Email:  "test@example.com",
		}
		ctx = WithUserContext(ctx, originalClaims)
		
		claims, ok := GetClaims(ctx)
		
		assert.True(t, ok)
		require.NotNil(t, claims)
		assert.Equal(t, originalClaims.UserID, claims.UserID)
		assert.Equal(t, originalClaims.Email, claims.Email)
	})
	
	t.Run("claims do not exist", func(t *testing.T) {
		ctx := context.Background()
		
		claims, ok := GetClaims(ctx)
		
		assert.False(t, ok)
		assert.Nil(t, claims)
	})
}

func TestContextIntegration(t *testing.T) {
	ctx := context.Background()
	
	claims := &Claims{
		UserID: "integration-test",
		Email:  "integration@example.com",
	}
	
	// Сохраняем в контекст
	ctx = WithUserContext(ctx, claims)
	
	// Извлекаем все данные
	userID, ok1 := GetUserID(ctx)
	email, ok2 := GetEmail(ctx)
	retrievedClaims, ok3 := GetClaims(ctx)
	
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.True(t, ok3)
	
	assert.Equal(t, "integration-test", userID)
	assert.Equal(t, "integration@example.com", email)
	assert.Equal(t, claims, retrievedClaims)
}

func TestContextKeyString(t *testing.T) {
	testCases := []struct {
		key      contextKey
		expected string
	}{
		{userIDKey, "auth context key user_id"},
		{emailKey, "auth context key email"},
		{claimsKey, "auth context key claims"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.key.String())
		})
	}
}