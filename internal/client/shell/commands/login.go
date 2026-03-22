// internal/shell/commands/login.go
package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/Ko4etov/gophkeeper/internal/client/crypto"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type LoginCommand struct {
    types.BaseCommand
}

func NewLoginCommand(ctx *types.CommandContext) *LoginCommand {
    return &LoginCommand{
        BaseCommand: types.NewPublicCommand(
            "login",
            []string{"signin"},
            "login <email> <password> - Login to your account",
            ctx,
        ),
    }
}

func (c *LoginCommand) Execute(args []string) error {
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    if currentUser != nil {
        c.BaseCommand.Printf("✅ Already logged in as %s\n", currentUser.Email)
        return nil
    }
    
    var email, password string
    var err error
    
    if len(args) >= 2 {
        email = args[0]
        password = args[1]
    } else {
        email, err = c.BaseCommand.Prompt("Email: ")
        if err != nil {
            return err
        }
        
        password, err = c.BaseCommand.PromptPassword("Password: ")
        if err != nil {
            return err
        }
    }
    
    email = strings.TrimSpace(email)
    password = strings.TrimSpace(password)
    
    if email == "" || password == "" {
        c.BaseCommand.Printf("❌ Email and password are required\n")
        return nil
    }
    
    c.BaseCommand.Printf("🔑 Logging in as %s... ", email)
    
    user, err := c.BaseCommand.Context.AuthService.Login(email, password)
    if err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
        return err
    }
    
    c.BaseCommand.Printf("✅ Welcome, %s!\n", user.Email)
    
    salt, err := c.BaseCommand.Context.DataService.GetSalt(email)
    if err != nil {
        c.BaseCommand.Printf("❌ Failed to check local storage: %v\n", err)
        return nil
    }
    
    if salt == nil {
        if err := c.setupMasterPassword(email); err != nil {
            return err
        }
    } else {
        if err := c.unlockStorage(email, salt); err != nil {
            return err
        }
    }
    
    c.BaseCommand.Context.SyncManager.Start(5 * time.Minute)
    
    return nil
}

// setupMasterPassword настраивает мастер-пароль при первом входе
func (c *LoginCommand) setupMasterPassword(email string) error {
    fmt.Println("\n🔐 First time login - setting up local encryption...")
    fmt.Println("⚠️  This master password will encrypt ALL your data.")
    fmt.Println("⚠️  It CANNOT be recovered if lost. Store it safely!\n")
    
    // Запрашиваем мастер-пароль
    masterPassword, err := c.BaseCommand.PromptPassword("Create master password: ")
    if err != nil {
        return err
    }
    
    confirmMaster, err := c.BaseCommand.PromptPassword("Confirm master password: ")
    if err != nil {
        return err
    }
    
    if masterPassword != confirmMaster {
        c.BaseCommand.Printf("❌ Master passwords do not match\n")
        return nil
    }
    
    // Генерируем соль
    salt, err := crypto.GenerateSalt()
    if err != nil {
        c.BaseCommand.Printf("❌ Failed to generate salt: %v\n", err)
        return nil
    }
    
    // Сохраняем соль
    if err := c.BaseCommand.Context.DataService.SaveSalt(salt, email); err != nil {
        c.BaseCommand.Printf("❌ Failed to save encryption salt: %v\n", err)
        return nil
    }
    
    if err := c.BaseCommand.Context.Session.Unlock(masterPassword, salt); err != nil {
        c.BaseCommand.Printf("❌ Failed to unlock storage: %v\n", err)
        return nil
    }
    
    fmt.Println("\n✅ Local storage initialized and unlocked!")
    fmt.Println("\n💡 Next steps:")
    fmt.Println("   • Use 'add login' to store your first password")
    fmt.Println("   • Use 'sync' to synchronize with server")
    fmt.Println("   • Use 'lock' to lock storage when done")
    
    return nil
}

// unlockStorage разблокирует существующее хранилище
func (c *LoginCommand) unlockStorage(email string, salt []byte) error {
    masterPassword, err := c.BaseCommand.PromptPassword("Master password: ")
    if err != nil {
        return err
    }
    
    if err := c.BaseCommand.Context.Session.Unlock(masterPassword, salt); err != nil {
        c.BaseCommand.Printf("❌ Invalid master password\n")
        return nil
    }
    
    fmt.Println("✅ Storage unlocked!")
    return nil
}