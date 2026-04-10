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
    if !w.Command.RequiresUnlock() {
        return w.Command.Execute(args)
    }
    
    if !w.session.IsUnlocked() {
        fmt.Println("\n🔒 Storage is locked. Master password required.")
        
        if err := w.unlockFunc(); err != nil {
            return fmt.Errorf("unlock failed: %w", err)
        }
        
        fmt.Println("✅ Storage unlocked successfully!")
    }
    
    w.session.Touch()
    
    return w.Command.Execute(args)
}