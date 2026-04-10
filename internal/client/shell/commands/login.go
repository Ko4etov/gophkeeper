// internal/shell/commands/login.go
package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Ko4etov/gophkeeper/internal/client/crypto"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
	"github.com/Ko4etov/gophkeeper/internal/common/auth"
	"github.com/Ko4etov/gophkeeper/internal/models"
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
    
    if err := c.handleSalt(user); err != nil {
        return err
    }
    
    c.BaseCommand.Context.SyncManager.Start(5 * time.Minute)
    c.BaseCommand.Printf("🔄 Sync started (every 5 minutes)\n")
    
    return nil
}

// handleSalt обрабатывает получение/сохранение соли (гибридный подход)
func (c *LoginCommand) handleSalt(user *models.User) error {
    ctx := context.Background()
    
    // 1. Пытаемся получить соль из локального хранилища
    localSalt, err := c.BaseCommand.Context.DataService.GetLocalSalt(user.Email)
    
    // 2. Если соль есть локально — используем её (офлайн-режим)
    if err == nil && localSalt != nil {
        c.BaseCommand.Printf("🔓 Using local salt (offline mode available)\n")
        return c.unlockStorage(user.Email, localSalt)
    }
    
    // 3. Соли нет локально — запрашиваем с сервера
    c.BaseCommand.Printf("🌐 No local salt found, fetching from server...\n")
    
    remoteSalt, err := c.BaseCommand.Context.DataService.GetRemoteSalt(ctx, user.Token)
    
    // 4. Если на сервере тоже нет соли — это первый вход, создаем новую
    if err != nil {
        c.BaseCommand.Printf("📝 No salt found on server, setting up master password...\n")
        return c.setupMasterPassword(user)
    }
    
    // 5. Сохраняем соль локально для будущих офлайн-сессий
    if err := c.BaseCommand.Context.DataService.SaveLocalSalt(remoteSalt, user.Email); err != nil {
        c.BaseCommand.Printf("⚠️ Warning: failed to save salt locally: %v\n", err)
    } else {
        c.BaseCommand.Printf("💾 Salt saved locally for offline access\n")
    }
    
    // 6. Разблокируем хранилище
    return c.unlockStorage(user.Email, remoteSalt)
}

// setupMasterPassword настраивает мастер-пароль при первом входе (новый пользователь)
func (c *LoginCommand) setupMasterPassword(user *models.User) error {
    fmt.Println("\n🔐 First time login - setting up local encryption...")
    fmt.Println("⚠️  This master password will encrypt ALL your data.")
    fmt.Println("⚠️  It CANNOT be recovered if lost. Store it safely!")
    
    ctx := context.Background()
    
    masterPassword, err := c.BaseCommand.PromptPassword("Create master password: ")
    if err != nil {
        return err
    }
    
    if err := auth.ValidatePasswordStrength(masterPassword); err != nil {
        c.BaseCommand.Printf("❌ %v\n", err)
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
    
    if err := c.BaseCommand.Context.DataService.SaveRemoteSalt(ctx, user.Token, salt); err != nil {
        c.BaseCommand.Printf("⚠️ Warning: failed to save salt on server: %v\n", err)
        c.BaseCommand.Printf("   Other devices won't be able to sync.\n")
    } else {
        c.BaseCommand.Printf("☁️ Salt saved on server for other devices\n")
    }
    
    if err := c.BaseCommand.Context.DataService.SaveLocalSalt(salt, user.Email); err != nil {
        c.BaseCommand.Printf("❌ Failed to save salt locally: %v\n", err)
        return nil
    }
    
    if err := c.BaseCommand.Context.Session.Unlock(masterPassword, salt); err != nil {
        c.BaseCommand.Printf("❌ Failed to unlock storage: %v\n", err)
        return nil
    }
    
    fmt.Println("\n✅ Local storage initialized and unlocked!")
    fmt.Println("💡 You can now work offline - salt is stored locally.")
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