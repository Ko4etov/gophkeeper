package grpcclient

import (
	"context"

	"github.com/Ko4etov/gophkeeper/internal/proto/sync"
)

// Sync создает stream для синхронизации данных с сервером.
// Возвращает клиент stream, через который можно отправлять и получать сообщения.
func (c *GrpcClient) Sync(ctx context.Context, token string) (sync.SyncService_SyncClient, error) {
	authCtx := c.GetAuthContext(ctx, token)

	stream, err := c.syncClient.Sync(authCtx)
	if err != nil {
		return nil, err
	}

	return stream, nil
}