// internal/shell/commands/sync.go
package commands

import (
	"fmt"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

// SyncCommand выполняет синхронизацию данных с сервером.
type SyncCommand struct {
    types.BaseCommand
}

// NewSyncCommand создает новую команду синхронизации.
func NewSyncCommand(ctx *types.CommandContext) *SyncCommand {
    return &SyncCommand{
        BaseCommand: types.NewSecureCommand(
            "sync",
            []string{"synchronize", "push", "pull"},
            "sync - Synchronize data with server",
            ctx,
        ),
    }
}

// Execute запускает процесс синхронизации.
func (c *SyncCommand) Execute(args []string) error {
    // Проверка авторизации
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }

    user := c.Context.AuthService.GetUser()
    if user == nil {
        return fmt.Errorf("not logged in")
    }

    c.Context.SyncManager.ForceSync()

    return nil
}