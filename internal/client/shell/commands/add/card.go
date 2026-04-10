// internal/shell/commands/add/card.go
package add

import (
	"fmt"
	"strings"
	"time"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/helpers"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type AddCardCommand struct {
    types.BaseCommand
}

func NewAddCardCommand(ctx *types.CommandContext) *AddCardCommand {
    return &AddCardCommand{
        BaseCommand: types.BaseCommand{
            NameText:    "card",
            AliasesText: []string{"bankcard", "creditcard", "cc"},
            HelpText:    "card <name> - Add bank card data",
            Context:     ctx,
        },
    }
}

func (c *AddCardCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    var name string
    if len(args) > 0 {
        name = args[0]
    } else {
        var err error
        name, err = c.BaseCommand.Prompt("Name for this card: ")
        if err != nil {
            return err
        }
    }
    
    if name == "" {
        c.BaseCommand.Printf("❌ Name is required\n")
		return nil
    }
    
    c.BaseCommand.Printf("💳 Adding bank card: %s\n", name)
    c.BaseCommand.Println(strings.Repeat("─", 40))
    
    number, err := c.BaseCommand.Prompt("Card number: ")
    if err != nil {
        return err
    }
    number = strings.ReplaceAll(number, " ", "")
    number = strings.ReplaceAll(number, "-", "")
    
    if !helpers.IsValidCardNumber(number) {
        c.BaseCommand.Printf("❌ Invalid card number\n")
		return nil
    }
    
    holder, err := c.BaseCommand.Prompt("Card holder: ")
    if err != nil {
        return err
    }
    
    monthStr, err := c.BaseCommand.Prompt("Expiry month (1-12): ")
    if err != nil {
        return err
    }
    var month int
    if _, err := fmt.Sscanf(monthStr, "%d", &month); err != nil || month < 1 || month > 12 {
        c.BaseCommand.Printf("❌ Invalid month\n")
		return nil
    }
    
    yearStr, err := c.BaseCommand.Prompt("Expiry year (YYYY): ")
    if err != nil {
        return err
    }
    var year int
    if _, err := fmt.Sscanf(yearStr, "%d", &year); err != nil || year < time.Now().Year() {
        c.BaseCommand.Printf("❌ Invalid year\n")
		return nil
    }
    
    cvv, _ := c.BaseCommand.Prompt("CVV (optional): ")
    if cvv != "" && !helpers.IsValidCVV(cvv) {
        c.BaseCommand.Printf("❌ CVV must be 3 or 4 digits\n")
		return nil
    }
    
    cardType, _ := c.BaseCommand.Prompt("Card type (visa/mastercard/mir, optional): ")
    
    bank, _ := c.BaseCommand.Prompt("Bank name (optional): ")
    
    tagsStr, _ := c.BaseCommand.Prompt("Tags (comma-separated, optional): ")
    var tags []string
    if tagsStr != "" {
        tags = strings.Split(tagsStr, ",")
        for i, tag := range tags {
            tags[i] = strings.TrimSpace(tag)
        }
    }
    
    c.BaseCommand.Printf("\n💾 Saving card... ")
    
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    if err := c.BaseCommand.Context.DataService.AddCard(name, number, holder, month, year, cvv, cardType, bank, tags, currentUser.Email); err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
		return err
    }
    
    c.BaseCommand.Printf("✅ Added successfully!\n")
    c.BaseCommand.Printf("💳 %s\n", helpers.MaskCardNumber(number))
    
    return nil
}

func (c *AddCardCommand) Completer() func(string) []string {
    return func(line string) []string {
        return []string{"visa", "mastercard", "mir", "amex"}
    }
}