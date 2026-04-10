package auth

import (
	"context"
)

// Определяем пользовательский тип для ключей (предотвращает коллизии)
type contextKey string

func (c contextKey) String() string {
    return "auth context key " + string(c)
}

// Константы ключей
const (
    userIDKey    contextKey = "user_id"
    emailKey     contextKey = "email"
    claimsKey    contextKey = "claims"
)

// WithUserContext добавляет данные пользователя в контекст
func WithUserContext(ctx context.Context, claims *Claims) context.Context {
    ctx = context.WithValue(ctx, userIDKey, claims.UserID)
    ctx = context.WithValue(ctx, emailKey, claims.Email)
    ctx = context.WithValue(ctx, claimsKey, claims)
    return ctx
}

// GetUserID извлекает user_id из контекста
func GetUserID(ctx context.Context) (string, bool) {
    val := ctx.Value(userIDKey)
    if val == nil {
        return "", false
    }
    userID, ok := val.(string)
    return userID, ok
}

// GetEmail извлекает email из контекста
func GetEmail(ctx context.Context) (string, bool) {
    val := ctx.Value(emailKey)
    if val == nil {
        return "", false
    }
    email, ok := val.(string)
    return email, ok
}

// GetClaims извлекает все claims
func GetClaims(ctx context.Context) (*Claims, bool) {
    val := ctx.Value(claimsKey)
    if val == nil {
        return nil, false
    }
    claims, ok := val.(*Claims)
    return claims, ok
}
