// internal/shell/commands/get.go
package commands

import (
	"encoding/json"
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/helpers"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
	"github.com/Ko4etov/gophkeeper/internal/models"
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

	switch item.Meta.DataType {
	case models.TypeLoginPassword:
		if data, ok := item.Data.(models.LoginPasswordData); ok {
			c.BaseCommand.Printf("  🔑 Login: %s\n", data.Login)
			c.BaseCommand.Printf("  🔐 Password: %s\n", helpers.MaskPassword(data.Password))
			if data.URL != "" {
				c.BaseCommand.Printf("  🌐 URL: %s\n", data.URL)
			}
			if data.Notes != "" {
				c.BaseCommand.Printf("  📝 Notes: %s\n", data.Notes)
			}
		}

	case models.TypeText:
		if data, ok := item.Data.(models.TextData); ok {
			c.BaseCommand.Printf("  📄 Content:\n%s\n", data.Content)
		}

	case models.TypeBankCard:
		if data, ok := item.Data.(models.BankCardData); ok {
			c.BaseCommand.Printf("  💳 Card: %s\n", helpers.MaskCardNumber(data.CardNumber))
			c.BaseCommand.Printf("  👤 Holder: %s\n", data.CardHolder)
			c.BaseCommand.Printf("  📅 Expires: %02d/%d\n", data.ExpiryMonth, data.ExpiryYear)
			if data.CVV != "" {
				c.BaseCommand.Printf("  🔐 CVV: ***\n")
			}
			if data.BankName != "" {
				c.BaseCommand.Printf("  🏦 Bank: %s\n", data.BankName)
			}
			if data.CardType != "" {
				c.BaseCommand.Printf("  💳 Type: %s\n", data.CardType)
			}
		}

	case models.TypeBinary:
		if data, ok := item.Data.(models.BinaryData); ok {
			c.BaseCommand.Printf("  📁 Filename: %s\n", data.Filename)
			c.BaseCommand.Printf("  📦 Size: %d bytes\n", data.Size)
			if data.MimeType != "" {
				c.BaseCommand.Printf("  🏷️ MIME: %s\n", data.MimeType)
			}
		}

	default:
		jsonData, _ := json.MarshalIndent(item.Data, "", "  ")
		c.BaseCommand.Printf("%s\n", string(jsonData))
	}

	return nil
}
