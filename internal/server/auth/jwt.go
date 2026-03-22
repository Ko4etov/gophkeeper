package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
    SecretKey     string
    AccessTTL     int
    RefreshTTL    int
}

type JWTManager struct {
    SecretKey     []byte
    AccessTTL     time.Duration
    RefreshTTL    time.Duration
}

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func NewJWTManager(config *JWTConfig) *JWTManager {
    return &JWTManager{
        SecretKey:  []byte(config.SecretKey),
        AccessTTL:  time.Duration(config.AccessTTL) * time.Minute,
        RefreshTTL: time.Duration(config.RefreshTTL) * time.Hour,
    }
}

func (m *JWTManager) GenerateAccessToken(userID, email string) (string, error) {
    claims := Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.AccessTTL)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   userID,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.SecretKey)
}

func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
    claims := jwt.RegisteredClaims{
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.RefreshTTL)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
        Subject:   userID,
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.SecretKey)
}

func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return m.SecretKey, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("invalid token")
}