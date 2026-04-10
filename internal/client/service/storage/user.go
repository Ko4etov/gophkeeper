// internal/client/storage/token.go
package storage

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Ko4etov/gophkeeper/internal/models"
	"go.etcd.io/bbolt"
)

// SaveUser сохраняет пользователя
func (s *Storage) SaveUser(user *models.User) error {
    return s.db.Update(func(tx *bbolt.Tx) error {
        userBucket := tx.Bucket(getUserBucketName(user.Email))
        if userBucket == nil {
            return errors.New("user not found")
        }

        metaBucket, err := userBucket.CreateBucketIfNotExists(bucketUserMeta)
        if err != nil {
            return fmt.Errorf("failed to create meta bucket: %w", err)
        }

        data, err := json.Marshal(user)
        if err != nil {
            return err
        }

        return metaBucket.Put(bucketUserData, data)
    })
}

// GetUser возвращает сохраненного пользователя
func (s *Storage) GetUser(email string) (*models.User, error) {
    var user models.User

    err := s.db.View(func(tx *bbolt.Tx) error {
        userBucket := tx.Bucket(getUserBucketName(email))
        if userBucket == nil {
            return errors.New("user not found")
        }

        metaBucket := userBucket.Bucket(bucketUserMeta)
        if metaBucket == nil {
            return errors.New("meta bucket not found")
        }

        data := metaBucket.Get(bucketUserData)
        if data == nil {
            return errors.New("user data not found")
        }

        return json.Unmarshal(data, &user)
    })

    if err != nil {
        return nil, err
    }

    return &user, nil
}