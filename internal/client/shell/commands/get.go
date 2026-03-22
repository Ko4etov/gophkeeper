// internal/shell/commands/get.go
package commands

import (
	"encoding/json"
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/helpers"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type GetCommand struct {
    types.BaseCommand
}

func NewGetCommand(ctx *types.CommandContext) *GetCommand {
    return &GetCommand{
        BaseCommand: types.NewSecureCommand(
            "get",
            []string{"show", "view", "cat"},
            "get <id> - Show detailed information about an item",
            ctx,
        ),
    }
}

func (c *GetCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    if len(args) == 0 {
        c.BaseCommand.Printf("❌ Usage: get <id>\n")
		return nil
    }
    
    id := args[0]
    
    c.BaseCommand.Printf("🔍 Fetching item %s... ", id)

    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    item, err := c.BaseCommand.Context.DataService.Get(currentUser.Email, id)
    if err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
		return err
    }
    
    c.BaseCommand.Printf("✅ Found\n\n")
    
    // Метаданные
    c.BaseCommand.Printf("📋 Metadata:\n")
    c.BaseCommand.Printf("  📝 Name: %s\n", item.Meta.Name)
    c.BaseCommand.Printf("  🏷️ Type: %s\n", item.Meta.DataType)
    c.BaseCommand.Printf("  🆔 ID: %s\n", item.Meta.ID)
    c.BaseCommand.Printf("  📅 Created: %s\n", item.Meta.CreatedAt.Format("2006-01-02 15:04:05"))
    c.BaseCommand.Printf("  📅 Updated: %s\n", item.Meta.UpdatedAt.Format("2006-01-02 15:04:05"))
    if len(item.Meta.Tags) > 0 {
        c.BaseCommand.Printf("  🏷️ Tags: %s\n", strings.Join(item.Meta.Tags, ", "))
    }
    if item.Meta.Description != "" {
        c.BaseCommand.Printf("  📄 Description: %s\n", item.Meta.Description)
    }
    
    c.BaseCommand.Printf("\n📦 Data:\n")
    
    // Отображаем данные в зависимости от типа
    switch item.Meta.DataType {
    case "login":
        if data, ok := item.Data.(map[string]interface{}); ok {
            c.BaseCommand.Printf("  🔑 Login: %s\n", data["login"])
            c.BaseCommand.Printf("  🔐 Password: %s\n", helpers.MaskPassword(data["password"].(string)))
            if url, ok := data["url"]; ok && url != "" {
                c.BaseCommand.Printf("  🌐 URL: %s\n", url)
            }
            if notes, ok := data["notes"]; ok && notes != "" {
                c.BaseCommand.Printf("  📝 Notes: %s\n", notes)
            }
        }
        
    case "text":
        if data, ok := item.Data.(map[string]interface{}); ok {
            c.BaseCommand.Printf("  📄 Content:\n%s\n", data["content"])
        }
        
    case "card":
        if data, ok := item.Data.(map[string]interface{}); ok {
            c.BaseCommand.Printf("  💳 Card: %s\n", helpers.MaskCardNumber(data["number"].(string)))
            c.BaseCommand.Printf("  👤 Holder: %s\n", data["holder"])
            c.BaseCommand.Printf("  📅 Expires: %02d/%d\n", 
                int(data["expiry_month"].(float64)),
                int(data["expiry_year"].(float64)))
            if cvv, ok := data["cvv"]; ok && cvv != "" {
                c.BaseCommand.Printf("  🔐 CVV: ***\n")
            }
            if bank, ok := data["bank"]; ok && bank != "" {
                c.BaseCommand.Printf("  🏦 Bank: %s\n", bank)
            }
        }
        
    default:
        jsonData, _ := json.MarshalIndent(item.Data, "", "  ")
        c.BaseCommand.Printf("%s\n", string(jsonData))
    }
    
    return nil
}