package wrappers

import (
	"fmt"

	"github.com/Ko4etov/gophkeeper/internal/client/crypto"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

// SecureCommandWrapper оборачивает команду и проверяет разблокировку
type SecureCommandWrapper struct {
    types.Command
    session *crypto.Session
    unlockFunc func() error
}

func NewSecureCommandWrapper(cmd types.Command, session *crypto.Session, unlockFunc func() error) types.Command {
    return &SecureCommandWrapper{
        Command:    cmd,
        session:    session,
        unlockFunc: unlockFunc,
    }
}

func (w *SecureCommandWrapper) Execute(args []string) error {
    // Если команда не требует разблокировки - выполняем сразу
    if !w.Command.RequiresUnlock() {
        return w.Command.Execute(args)
    }
    
    // Проверяем, разблокировано ли хранилище
    if !w.session.IsUnlocked() {
        fmt.Println("\n🔒 Storage is locked. Master password required.")
        
        // Запрашиваем мастер-пароль
        if err := w.unlockFunc(); err != nil {
            return fmt.Errorf("unlock failed: %w", err)
        }
        
        fmt.Println("✅ Storage unlocked successfully!")
    }
    
    // Обновляем время активности
    w.session.Touch()
    
    // Выполняем команду
    return w.Command.Execute(args)
}