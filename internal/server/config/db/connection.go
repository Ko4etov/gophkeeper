// Package db предоставляет функции для работы с базой данных.
package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDBConnection создает новое подключение к базе данных.
func NewDBConnection(address string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(
		context.Background(),
		address)

	if err != nil {
		return nil, err
	}

	return pool, nil
}