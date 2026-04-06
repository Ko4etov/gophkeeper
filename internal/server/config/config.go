// Package config предоставляет конфигурацию для сервера.
package config

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ko4etov/gophkeeper/internal/server/config/db"
	"github.com/Ko4etov/gophkeeper/internal/server/service/logger"
)

// ServerConfig содержит все параметры конфигурации сервера.
type ServerConfig struct {
	ServerAddress   string        // адрес сервера
	ConnectionPool  *pgxpool.Pool // пул подключений к базе данных
	HashKey         string        // ключ для хеширования
	CryptoKey       string        // директория для сохранения профилей
	GRPCAddress     string
	JWTSecret       string
	AccessTokenTTL  int
	RefreshTokenTTL int
}

// New создает новую конфигурацию сервера.
func New() (*ServerConfig, error) {
	var pool *pgxpool.Pool

	if err := logger.Initialize("info"); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLogerInitialization, err)
	}

	serverParameters := parseServerParameters()

	if serverParameters.HashKey == "" {
		return nil, HashKeyMissed
	}

	if serverParameters.JWTSecret == "" {
		return nil, JWTSecretMissed
	}

	if serverParameters.GRPCAddress == "" {
		return nil, ErrGRPCAddressMissed
	}

	if serverParameters.DBAddress == "" {
		return nil, DBAddressMissed
	}

	if _, err := pgxpool.ParseConfig(serverParameters.DBAddress); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseDBConfig, err)
	}

	if err := db.RunMigrations(serverParameters.DBAddress); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMigration, err)
	}

	pool, err := db.NewDBConnection(serverParameters.DBAddress)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDBConnection, err)
	}

	return &ServerConfig{
		ServerAddress:   serverParameters.Address,
		ConnectionPool:  pool,
		HashKey:         serverParameters.HashKey,
		CryptoKey:       serverParameters.CryptoKey,
		GRPCAddress:     serverParameters.GRPCAddress,
		AccessTokenTTL:  serverParameters.AccessTokenTTL,
		RefreshTokenTTL: serverParameters.RefreshTokenTTL,
	}, nil
}

var (
	ErrMigration           = errors.New("migration error")
	ErrParseDBConfig       = errors.New("parse db config error")
	ErrDBConnection        = errors.New("db connection error")
	ErrLogerInitialization = errors.New("logger initialization error")
	ErrGRPCAddressMissed   = errors.New("grpc address missed error")
	JWTSecretMissed        = errors.New("jwt secret missed error")
	DBAddressMissed        = errors.New("db address missed error")
	HashKeyMissed          = errors.New("hash key missed error")
)
