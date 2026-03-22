package models

import "time"

// User представляет пользователя системы
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Token        string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}