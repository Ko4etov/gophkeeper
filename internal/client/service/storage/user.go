// internal/client/storage/token.go
package storage

import (
	"encoding/json"
	"errors"

	"github.com/Ko4etov/gophkeeper/internal/models"
	"go.etcd.io/bbolt"
)

// SaveToken сохраняет пользователя
func (s *Storage) SaveUser(user *models.User) error {
    return s.db.Update(func(tx *bbolt.Tx) error {
        userBucket := tx.Bucket(getUserBucketName(user.Email))
        if userBucket == nil {
            return errors.New("user not found")
        }

        metaBucket := userBucket.Bucket(bucketUserMeta)

        data, err := json.Marshal(user)
        if err != nil {
            return err
        }

        return metaBucket.Put(bucketUserMeta, data)
    })
}

// GetToken возвращает сохраненный токен
func (s *Storage) GetUser(email string) (*models.User, error) {
    var user models.User

    err := s.db.View(func(tx *bbolt.Tx) error {
        userBucket := tx.Bucket(getUserBucketName(email))
        if userBucket == nil {
            return errors.New("user not found")
        }

        metaBucket := userBucket.Bucket(bucketUserMeta)

        data := metaBucket.Get(bucketUserMeta)
        if metaBucket == nil {
            return errors.New("meta bucket not found")
        }

        return json.Unmarshal(data, &user)
    })

    if err != nil {
        return nil, err
    }

    return &user, nil
}