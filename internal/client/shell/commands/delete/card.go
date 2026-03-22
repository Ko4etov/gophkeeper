package delete

import (
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/helpers"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type DeleteCardCommand struct {
    types.BaseCommand
}

func NewDeleteCardCommand(ctx *types.CommandContext) *DeleteCardCommand {
    return &DeleteCardCommand{
        BaseCommand: types.BaseCommand{
            NameText:    "card",
            AliasesText: []string{"bankcard", "creditcard"},
            HelpText:    "card <id> - Delete bank card data",
            Context:     ctx,
        },
    }
}

func (c *DeleteCardCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    if len(args) == 0 {
        c.BaseCommand.Printf("❌ Please specify card ID\n")
        c.BaseCommand.Printf("Usage: delete card <id>\n")
        return nil
    }
    
    cardID := args[0]
    
    // Получаем информацию о карте для подтверждения
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    entry, err := c.BaseCommand.Context.DataService.GetCard(cardID, currentUser.Email)
    if err != nil {
        c.BaseCommand.Printf("❌ Failed to get card: %v\n", err)
        return err
    }
    
    c.BaseCommand.Printf("⚠️  You are about to DELETE this card:\n")
    c.BaseCommand.Println(strings.Repeat("─", 40))
    c.BaseCommand.Printf("ID:   %s\n", cardID)
    c.BaseCommand.Printf("Name: %s\n", entry.Meta.Name)
    c.BaseCommand.Printf("Card: %s\n", helpers.MaskCardNumber(entry.Data.(map[string]interface{})["card_number"].(string)))
    c.BaseCommand.Println(strings.Repeat("─", 40))
    
    confirm, err := c.BaseCommand.Confirm("Are you sure? This cannot be undone")
    if err != nil || !confirm {
        c.BaseCommand.Printf("❌ Delete cancelled\n")
        return nil
    }
    
    c.BaseCommand.Printf("\n🗑️  Deleting card... ")
    
    if err := c.BaseCommand.Context.DataService.Delete(cardID, currentUser.Email); err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
        return err
    }
    
    c.BaseCommand.Printf("✅ Card deleted successfully!\n")
    return nil
}