package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/Ko4etov/gophkeeper/internal/client/config"
	"github.com/Ko4etov/gophkeeper/internal/client/crypto"
	"github.com/Ko4etov/gophkeeper/internal/models"
	"go.etcd.io/bbolt"
)

var (
	bucketUserData = []byte("data")
	bucketUserMeta = []byte("meta")
)

type Storage struct {
	db      *bbolt.DB
	session *crypto.Session
}

type StoredEntry struct {
	Meta models.DataMeta `json:"meta"`
	Data json.RawMessage `json:"data"`
}

// NewStorage открывает или создает БД
func NewStorage(config *config.ClientConfig, session *crypto.Session) (*Storage, error) {
	dbPath := filepath.Join(config.DataDir, "gophkeeper.db")

	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
		session: session,
	}, nil
}

// getUserBucketName возвращает имя бакета для пользователя
func getUserBucketName(email string) []byte {
	return []byte("user:" + email)
}

// ensureUserBuckets создает бакеты для пользователя
func (s *Storage) EnsureUserBuckets(email string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		userBucket, err := tx.CreateBucketIfNotExists(getUserBucketName(email))
		if err != nil {
			return err
		}

		// Создаем подбакеты внутри бакета пользователя
		if _, err := userBucket.CreateBucketIfNotExists(bucketUserMeta); err != nil {
			return err
		}
		if _, err := userBucket.CreateBucketIfNotExists(bucketUserData); err != nil {
			return err
		}

		return nil
	})
}

// Close закрывает БД
func (s *Storage) Close() error {
	return s.db.Close()
}

// GetSalt возвращает соль из хранилища
func (s *Storage) GetSalt(email string) ([]byte, error) {
	var salt []byte

	err := s.db.View(func(tx *bbolt.Tx) error {
		userBucket := tx.Bucket(getUserBucketName(email))
		if userBucket == nil {
			return nil
		}

		metaBucket := userBucket.Bucket([]byte(bucketUserMeta))
		if metaBucket == nil {
			return nil
		}

		data := metaBucket.Get([]byte("salt"))
		if data == nil {
			return nil
		}

		salt = make([]byte, len(data))
		copy(salt, data)
		return nil
	})

	return salt, err
}

func (s *Storage) SaveSalt(email string, salt []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		userBucket, err := tx.CreateBucketIfNotExists(getUserBucketName(email))
		if err != nil {
			return err
		}

		metaBucket, err := userBucket.CreateBucketIfNotExists([]byte(bucketUserMeta))
		if err != nil {
			return err
		}

		return metaBucket.Put([]byte("salt"), salt)
	})
}

// SaveEntry сохраняет запись
func (s *Storage) SaveEntry(email string, entry *models.DataEntry) error {
	encryptor, err := s.session.GetEncryptor()
	if err != nil {
		return err
	}

	dataJSON, err := json.Marshal(entry.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	encryptedData, err := encryptor.Encrypt(dataJSON)
	if err != nil {
		return err
	}

	stored := StoredEntry{
		Meta: entry.Meta,
		Data: json.RawMessage(fmt.Sprintf(`"%s"`, encryptedData)),
	}

	storedJSON, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("failed to marshal stored entry: %w", err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		userBucket := tx.Bucket(getUserBucketName(email))
		if userBucket == nil {
			return errors.New("user not found")
		}

		dataBucket := userBucket.Bucket(bucketUserData)
		if dataBucket == nil {
			return errors.New("data bucket not found")
		}

		return dataBucket.Put([]byte(entry.Meta.ID), storedJSON)
	})
}

// GetEntry возвращает запись по ID
func (s *Storage) GetEntry(email string, id string) (*models.DataEntry, error) {
	encryptor, err := s.session.GetEncryptor()
	if err != nil {
		return nil, err
	}

	var stored StoredEntry

	err = s.db.View(func(tx *bbolt.Tx) error {
		userBucket := tx.Bucket(getUserBucketName(email))
		if userBucket == nil {
			return errors.New("user not found")
		}

		dataBucket := userBucket.Bucket(bucketUserData)
		if dataBucket == nil {
			return errors.New("data bucket not found")
		}

		data := dataBucket.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("entry %s not found", id)
		}
		return json.Unmarshal(data, &stored)
	})
	if err != nil {
		return nil, err
	}

	var encryptedStr string
	if err := json.Unmarshal(stored.Data, &encryptedStr); err != nil {
		return nil, err
	}

	decryptedData, err := encryptor.Decrypt(encryptedStr)
	if err != nil {
		return nil, err
	}

	var data interface{}
	switch stored.Meta.DataType {
	case models.TypeLoginPassword:
		var loginData models.LoginPasswordData
		if err := json.Unmarshal(decryptedData, &loginData); err != nil {
			return nil, err
		}
		data = loginData
	case models.TypeText:
		var textData models.TextData
		if err := json.Unmarshal(decryptedData, &textData); err != nil {
			return nil, err
		}
		data = textData
	case models.TypeBankCard:
		var cardData models.BankCardData
		if err := json.Unmarshal(decryptedData, &cardData); err != nil {
			return nil, err
		}
		data = cardData
	default:
		var raw map[string]interface{}
		if err := json.Unmarshal(decryptedData, &raw); err != nil {
			return nil, err
		}
		data = raw
	}

	return &models.DataEntry{
		Meta: stored.Meta,
		Data: data,
	}, nil
}

// ListEntries возвращает все записи (только метаданные)
func (s *Storage) ListEntriesMeta(email string) ([]*models.DataMeta, error) {
	var entries []*models.DataMeta

	err := s.db.View(func(tx *bbolt.Tx) error {
		userBucket := tx.Bucket(getUserBucketName(email))
		if userBucket == nil {
			return errors.New("user not found")
		}

		dataBucket := userBucket.Bucket(bucketUserData)
		if dataBucket == nil {
			return errors.New("data bucket not found")
		}

		return dataBucket.ForEach(func(k, v []byte) error {
			var stored StoredEntry
			if err := json.Unmarshal(v, &stored); err != nil {
				return err
			}
			entries = append(entries, &stored.Meta)
			return nil
		})
	})

	return entries, err
}

// DeleteEntry удаляет запись
func (s *Storage) DeleteEntry(email string, id string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		userBucket := tx.Bucket(getUserBucketName(email))
		if userBucket == nil {
			return errors.New("user not found")
		}

		dataBucket := userBucket.Bucket(bucketUserData)
		if dataBucket == nil {
			return errors.New("data bucket not found")
		}
		return dataBucket.Delete([]byte(id))
	})
}

func (s *Storage) GetEntriesByIDs(email string, ids []string) ([]*models.DataEntry, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	idMap := make(map[string]bool, len(ids))
	for _, id := range ids {
		idMap[id] = true
	}

	var entries []*models.DataEntry

	err := s.db.View(func(tx *bbolt.Tx) error {
		userBucket := tx.Bucket(getUserBucketName(email))
		if userBucket == nil {
			return nil
		}

		dataBucket := userBucket.Bucket([]byte(bucketUserData))
		if dataBucket == nil {
			return nil
		}

		return dataBucket.ForEach(func(k, v []byte) error {
			if !idMap[string(k)] {
				return nil
			}

			entry, err := s.GetEntry(email, string(k))
			if err != nil {
				return err
			}
			entries = append(entries, entry)
			return nil
		})
	})

	return entries, err
}

func (s *Storage) GetEntriesMetaByIDs(email string, ids []string) ([]*models.DataMeta, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	idMap := make(map[string]bool, len(ids))
	for _, id := range ids {
		idMap[id] = true
	}

	var entriesMeta []*models.DataMeta

	err := s.db.View(func(tx *bbolt.Tx) error {
		userBucket := tx.Bucket(getUserBucketName(email))
		if userBucket == nil {
			return nil
		}

		dataBucket := userBucket.Bucket([]byte(bucketUserData))
		if dataBucket == nil {
			return nil
		}

		return dataBucket.ForEach(func(k, v []byte) error {
			if !idMap[string(k)] {
				return nil
			}

			var stored StoredEntry
			if err := json.Unmarshal(v, &stored); err != nil {
				return err
			}

			entriesMeta = append(entriesMeta, &stored.Meta)
			return nil
		})
	})

	return entriesMeta, err
}

func (s *Storage) BatchDelete(email string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		userBucket := tx.Bucket(getUserBucketName(email))
		if userBucket == nil {
			return errors.New("user not found")
		}

		dataBucket := userBucket.Bucket(bucketUserData)
		if dataBucket == nil {
			return errors.New("data bucket not found")
		}

		for _, id := range ids {
			if err := dataBucket.Delete([]byte(id)); err != nil {
				return fmt.Errorf("failed to delete entry %s: %w", id, err)
			}
		}
		return nil
	})
}

func (s *Storage) GetEncryptedEntry(email string, id string) (string, error) {
    var stored StoredEntry

    err := s.db.View(func(tx *bbolt.Tx) error {
        userBucket := tx.Bucket(getUserBucketName(email))
        if userBucket == nil {
            return errors.New("user not found")
        }

        dataBucket := userBucket.Bucket(bucketUserData)
        if dataBucket == nil {
            return errors.New("data bucket not found")
        }

        data := dataBucket.Get([]byte(id))
        if data == nil {
            return fmt.Errorf("entry %s not found", id)
        }
        return json.Unmarshal(data, &stored)
    })
    if err != nil {
        return "", err
    }

    var encryptedStr string
    if err := json.Unmarshal(stored.Data, &encryptedStr); err != nil {
        return "", err
    }

    return encryptedStr, nil
}

func (s *Storage) SaveEncryptedEntry(email string, entry *models.DataEntry, encryptedData string) error {
    stored := StoredEntry{
        Meta: entry.Meta,
        Data: json.RawMessage(fmt.Sprintf(`"%s"`, encryptedData)),
    }

    storedJSON, err := json.Marshal(stored)
    if err != nil {
        return fmt.Errorf("failed to marshal stored entry: %w", err)
    }

    return s.db.Update(func(tx *bbolt.Tx) error {
        userBucket, err := tx.CreateBucketIfNotExists(getUserBucketName(email))
        if err != nil {
            return err
        }

        dataBucket, err := userBucket.CreateBucketIfNotExists([]byte(bucketUserData))
        if err != nil {
            return err
        }

        return dataBucket.Put([]byte(entry.Meta.ID), storedJSON)
    })
}
