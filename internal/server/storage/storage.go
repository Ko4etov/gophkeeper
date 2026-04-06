package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type StorageInterface interface {
	// User methods
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, email, password string) (*User, error)
	
	// Refresh token methods
	CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
	
	// Record methods
	GetAllRecordsMeta(ctx context.Context, userID string) ([]*RecordMeta, error)
	GetFullRecordsByIDs(ctx context.Context, userID string, ids []string) ([]*Record, error)
	CreateRecord(ctx context.Context, userID string, record *Record) error
	UpdateRecord(ctx context.Context, userID string, record *Record) error
	DeleteRecord(ctx context.Context, userID string, clientID string) error
}

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

type Record struct {
	ID          int64           `db:"id" json:"id"`
	ClientID    string          `db:"client_id" json:"client_id"`
	UserID      string          `db:"user_id" json:"user_id"`
	DataType    string          `db:"data_type" json:"data_type"`
	Name        string          `db:"name" json:"name"`
	Tags        []string        `db:"tags" json:"tags,omitempty"`
	Data        json.RawMessage `db:"data" json:"data"`
	Version     int32           `db:"version" json:"version"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

type RecordMeta struct {
	ClientID  string       `db:"client_id" json:"client_id"`
	Version   int32        `db:"version" json:"version"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
}

type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool, ctx context.Context) (*Storage, error) {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Storage{db: pool}, nil
}

func (s *Storage) Close() error {
	s.db.Close()
	return nil
}

// GetAllRecordsMeta получает только метаданные всех записей пользователя
func (s *Storage) GetAllRecordsMeta(ctx context.Context, userID string) ([]*RecordMeta, error) {
	query := `
		SELECT client_id, version, updated_at
		FROM records
		WHERE user_id = $1
	`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metas []*RecordMeta
	for rows.Next() {
		var meta RecordMeta
		err := rows.Scan(&meta.ClientID, &meta.Version, &meta.UpdatedAt)
		if err != nil {
			return nil, err
		}
		metas = append(metas, &meta)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return metas, nil
}

// GetFullRecordsByIDs получает полные записи только для указанных ID
func (s *Storage) GetFullRecordsByIDs(ctx context.Context, userID string, ids []string) ([]*Record, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT client_id, data_type, name, tags, data, version, created_at, updated_at
		FROM records
		WHERE user_id = $1 AND client_id = ANY($2)
	`

	rows, err := s.db.Query(ctx, query, userID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*Record
	for rows.Next() {
		var record Record
		var tags interface{}

		err := rows.Scan(
			&record.ClientID,
			&record.DataType,
			&record.Name,
			&tags,
			&record.Data,
			&record.Version,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if tags != nil {
			switch t := tags.(type) {
			case []string:
				record.Tags = t
			case []byte:
				json.Unmarshal(t, &record.Tags)
			}
		}

		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// CreateRecord создает новую запись
func (s *Storage) CreateRecord(ctx context.Context, userID string, record *Record) error {
	query := `
		INSERT INTO records (client_id, user_id, data_type, name, tags, data, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := s.db.Exec(ctx, query,
		record.ClientID,
		userID,
		record.DataType,
		record.Name,
		record.Tags,
		record.Data,
		record.Version,
		record.CreatedAt,
		record.UpdatedAt,
	)

	return err
}

// UpdateRecord обновляет существующую запись
func (s *Storage) UpdateRecord(ctx context.Context, userID string, record *Record) error {
	query := `
		UPDATE records
		SET name = $1, tags = $2, data = $3, 
		    version = $4, updated_at = $5
		WHERE user_id = $6 AND client_id = $7
	`

	_, err := s.db.Exec(ctx, query,
		record.Name,
		record.Tags,
		record.Data,
		record.Version,
		record.UpdatedAt,
		userID,
		record.ClientID,
	)

	return err
}

// DeleteRecord удаляет запись
func (s *Storage) DeleteRecord(ctx context.Context, userID, clientID string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM records WHERE user_id = $1 AND client_id = $2", userID, clientID)
	return err
}

func (s *Storage) CreateUser(ctx context.Context, email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var user User
	query := `
        INSERT INTO users (email, password_hash)
        VALUES ($1, $2)
        RETURNING id, email, password_hash, created_at, updated_at
    `

	err = s.db.QueryRow(ctx, query, email, string(hashedPassword)).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `
        SELECT id, email, password_hash, created_at, updated_at
        FROM users
        WHERE email = $1
    `

	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *Storage) GetUserByID(ctx context.Context, id string) (*User, error) {
	var user User
	query := `
        SELECT id, email, password_hash, created_at, updated_at
        FROM users
        WHERE id = $1
    `

	err := s.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *Storage) CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	query := `
        INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
        VALUES ($1, $2, $3)
    `

	_, err := s.db.Exec(ctx, query, userID, string(hashedToken), expiresAt)
	return err
}

func (s *Storage) ValidateRefreshToken(ctx context.Context, token string) (string, error) {
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	var userID string
	var expiresAt time.Time
	var revokedAt sql.NullTime

	query := `
        SELECT user_id, expires_at, revoked_at
        FROM refresh_tokens
        WHERE token_hash = $1
    `

	err := s.db.QueryRow(ctx, query, tokenHash).Scan(
		&userID, &expiresAt, &revokedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("token not found: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("database error: %w", err)
	}

	if time.Now().After(expiresAt) {
		return "", fmt.Errorf("token expired")
	}

	if revokedAt.Valid {
		return "", fmt.Errorf("token revoked")
	}

	return userID, nil
}

func (s *Storage) RevokeRefreshToken(ctx context.Context, token string) error {
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1`
	_, err := s.db.Exec(ctx, query, tokenHash)
	return err
}

func (s *Storage) RevokeAllUserTokens(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`
	_, err := s.db.Exec(ctx, query, userID)
	return err
}

func (s *Storage) GetUserSalt(ctx context.Context, userID string) (string, error) {
	var saltBase64 string
	query := `SELECT salt FROM user_salt WHERE user_id = $1`
	
	err := s.db.QueryRow(ctx, query, userID).Scan(&saltBase64)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get salt: %w", err)
	}
	
	return saltBase64, nil
}

// SaveUserSalt сохраняет соль пользователя.
func (s *Storage) SaveUserSalt(ctx context.Context, userID, saltBase64 string) error {
	query := `
		INSERT INTO user_salt (user_id, salt, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET salt = EXCLUDED.salt, updated_at = NOW()
	`
	
	_, err := s.db.Exec(ctx, query, userID, saltBase64)
	if err != nil {
		return fmt.Errorf("failed to save salt: %w", err)
	}
	
	return nil
}