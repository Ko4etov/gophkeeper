// internal/shell/commands/register.go
package commands

import (
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type RegisterCommand struct {
    types.BaseCommand
}

func NewRegisterCommand(ctx *types.CommandContext) *RegisterCommand {
    return &RegisterCommand{
        BaseCommand: types.NewPublicCommand(
            "register",
            []string{"reg", "signup"},
            "register <email> <password> - Create a new account",
            ctx,
        ),
    }
}

func (c *RegisterCommand) Execute(args []string) error {
    var email, password string
    var err error
    
    if len(args) >= 2 {
        email = args[0]
        password = args[1]
    } else {
        // Интерактивный ввод
        email, err = c.BaseCommand.Prompt("Email: ")
        if err != nil {
            return err
        }
        
        password, err = c.BaseCommand.PromptPassword("Password: ")
        if err != nil {
            return err
        }
        
        // Подтверждение пароля
        confirm, err := c.BaseCommand.PromptPassword("Confirm password: ")
        if err != nil {
            return err
        }
        
        if password != confirm {
            c.BaseCommand.Printf("❌ Passwords do not match\n")
			return nil
        }
    }
    
    // Валидация
    email = strings.TrimSpace(email)
    password = strings.TrimSpace(password)
    
    if email == "" {
        c.BaseCommand.Printf("❌ Email cannot be empty\n")
		return nil
    }
    if !strings.Contains(email, "@") {
        c.BaseCommand.Printf("❌ Invalid email format\n")
		return nil
    }
    if len(password) < 8 {
        c.BaseCommand.Printf("❌ Password must be at least 8 characters\n")
		return nil
    }
    
    c.BaseCommand.Printf("📝 Registering %s... ", email)
    
    if err := c.BaseCommand.Context.AuthService.Register(email, password); err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
		return nil 
    }
    
    c.BaseCommand.Printf("✅ Success! You can now login with 'login %s'\n", email)
    return nil
}

func (c *RegisterCommand) Completer() func(string) []string {
    return func(line string) []string {
        // Можно добавить автодополнение email из истории
        return []string{}
    }
}