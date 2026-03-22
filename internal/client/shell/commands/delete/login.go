// internal/shell/commands/delete/login.go
package delete

import (
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type DeleteLoginCommand struct {
    types.BaseCommand
}

func NewDeleteLoginCommand(ctx *types.CommandContext) *DeleteLoginCommand {
    return &DeleteLoginCommand{
        BaseCommand: types.BaseCommand{
            NameText:    "login",
            AliasesText: []string{"pass", "password", "account"},
            HelpText:    "login <id> - Delete login/password entry",
            Context:     ctx,
        },
    }
}

func (c *DeleteLoginCommand) Execute(args []string) error {
    if len(args) == 0 {
        c.Printf("❌ Please specify login ID\n")
        c.Printf("Usage: delete login <id>\n")
        return nil
    }
    
    loginID := args[0]
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    
    // Получаем информацию о записи для подтверждения
    entry, err := c.BaseCommand.Context.DataService.GetLogin(loginID, currentUser.Email)
    if err != nil {
        c.Printf("❌ Failed to get login: %v\n", err)
        return err
    }
    
    c.Printf("⚠️  You are about to DELETE this login:\n")
    c.Println(strings.Repeat("─", 40))
    c.Printf("ID:    %s\n", loginID)
    c.Printf("Name:  %s\n", entry.Meta.Name)
    c.Printf("Login: %s\n", entry.Data.(map[string]interface{})["login"])
    c.Printf("URL:   %s\n", entry.Data.(map[string]interface{})["url"])
    c.Println(strings.Repeat("─", 40))
    
    confirm, err := c.Confirm("Are you sure? This cannot be undone")
    if err != nil || !confirm {
        c.Printf("❌ Delete cancelled\n")
        return nil
    }
    
    c.Printf("\n🗑️  Deleting login... ")
    
    if err := c.Context.DataService.Delete(loginID, currentUser.Email); err != nil {
        c.Printf("❌ Failed: %v\n", err)
        return err
    }
    
    c.Printf("✅ Login deleted successfully!\n")
    return nil
}