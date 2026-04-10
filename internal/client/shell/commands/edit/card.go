package edit

import (
	"fmt"
	"strings"
	"time"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/helpers"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
	"github.com/Ko4etov/gophkeeper/internal/models"
)

type EditCardCommand struct {
    types.BaseCommand
}

func NewEditCardCommand(ctx *types.CommandContext) *EditCardCommand {
    return &EditCardCommand{
        BaseCommand: types.BaseCommand{
            NameText:    "card",
            AliasesText: []string{"bankcard", "creditcard"},
            HelpText:    "card <id> - Edit bank card data",
            Context:     ctx,
        },
    }
}

func (c *EditCardCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    if len(args) == 0 {
        c.BaseCommand.Printf("❌ Please specify card ID\n")
        c.BaseCommand.Printf("Usage: edit card <id>\n")
        return nil
    }
    
    cardID := args[0]
    
    // Получаем текущие данные карты
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    entry, err := c.BaseCommand.Context.DataService.GetCard(cardID, currentUser.Email)
    if err != nil {
        c.BaseCommand.Printf("❌ Failed to get card: %v\n", err)
        return err
    }
    
    // ✅ Правильное приведение типа
    cardData, ok := entry.Data.(models.BankCardData)
    if !ok {
        c.BaseCommand.Printf("❌ Invalid data type for card entry\n")
        return nil
    }
    
    c.BaseCommand.Printf("💳 Editing card: %s\n", entry.Meta.Name)
    c.BaseCommand.Println(strings.Repeat("─", 40))
    
    // Показываем текущие значения и запрашиваем новые
    c.BaseCommand.Printf("Current values (press Enter to keep current):\n")
    
    // Имя
    namePrompt := fmt.Sprintf("Name [%s]: ", entry.Meta.Name)
    name, err := c.BaseCommand.Prompt(namePrompt)
    if err != nil {
        return err
    }
    if name == "" {
        name = entry.Meta.Name
    }
    
    // Номер карты
    numberPrompt := fmt.Sprintf("Card number [%s]: ", helpers.MaskCardNumber(cardData.CardNumber))
    number, err := c.BaseCommand.Prompt(numberPrompt)
    if err != nil {
        return err
    }
    number = strings.ReplaceAll(number, " ", "")
    number = strings.ReplaceAll(number, "-", "")
    if number == "" {
        number = cardData.CardNumber
    } else if !helpers.IsValidCardNumber(number) {
        c.BaseCommand.Printf("❌ Invalid card number\n")
        return nil
    }
    
    // Владелец
    holderPrompt := fmt.Sprintf("Card holder [%s]: ", cardData.CardHolder)
    holder, err := c.BaseCommand.Prompt(holderPrompt)
    if err != nil {
        return err
    }
    if holder == "" {
        holder = cardData.CardHolder
    }
    
    // Месяц
    monthPrompt := fmt.Sprintf("Expiry month [%d]: ", cardData.ExpiryMonth)
    monthStr, err := c.BaseCommand.Prompt(monthPrompt)
    if err != nil {
        return err
    }
    month := cardData.ExpiryMonth
    if monthStr != "" {
        if _, err := fmt.Sscanf(monthStr, "%d", &month); err != nil || month < 1 || month > 12 {
            c.BaseCommand.Printf("❌ Invalid month\n")
            return nil
        }
    }
    
    // Год
    yearPrompt := fmt.Sprintf("Expiry year [%d]: ", cardData.ExpiryYear)
    yearStr, err := c.BaseCommand.Prompt(yearPrompt)
    if err != nil {
        return err
    }
    year := cardData.ExpiryYear
    if yearStr != "" {
        if _, err := fmt.Sscanf(yearStr, "%d", &year); err != nil || year < time.Now().Year() {
            c.BaseCommand.Printf("❌ Invalid year\n")
            return nil
        }
    }
    
    // CVV (опционально)
    cvvPrompt := fmt.Sprintf("CVV [%s]: ", cardData.CVV)
    cvv, err := c.BaseCommand.Prompt(cvvPrompt)
    if err != nil {
        return err
    }
    if cvv == "" {
        cvv = cardData.CVV
    } else if cvv != "" && !helpers.IsValidCVV(cvv) {
        c.BaseCommand.Printf("❌ CVV must be 3 or 4 digits\n")
        return nil
    }
    
    // Тип карты
    typePrompt := fmt.Sprintf("Card type [%s]: ", cardData.CardType)
    cardType, err := c.BaseCommand.Prompt(typePrompt)
    if err != nil {
        return err
    }
    if cardType == "" {
        cardType = cardData.CardType
    }
    
    // Банк
    bankPrompt := fmt.Sprintf("Bank name [%s]: ", cardData.BankName)
    bank, err := c.BaseCommand.Prompt(bankPrompt)
    if err != nil {
        return err
    }
    if bank == "" {
        bank = cardData.BankName
    }
    
    // Теги
    currentTags := strings.Join(entry.Meta.Tags, ", ")
    tagsPrompt := fmt.Sprintf("Tags [%s]: ", currentTags)
    tagsStr, err := c.BaseCommand.Prompt(tagsPrompt)
    if err != nil {
        return err
    }
    var tags []string
    if tagsStr != "" {
        tags = strings.Split(tagsStr, ",")
        for i, tag := range tags {
            tags[i] = strings.TrimSpace(tag)
        }
    } else {
        tags = entry.Meta.Tags
    }
    
    // Подтверждение
    c.BaseCommand.Printf("\n📝 Changes to save:\n")
    c.BaseCommand.Printf("  Name: %s → %s\n", entry.Meta.Name, name)
    c.BaseCommand.Printf("  Card: %s → %s\n", helpers.MaskCardNumber(cardData.CardNumber), helpers.MaskCardNumber(number))
    
    confirm, err := c.BaseCommand.Confirm("Save changes?")
    if err != nil || !confirm {
        c.BaseCommand.Printf("❌ Edit cancelled\n")
        return nil
    }
    
    c.BaseCommand.Printf("\n💾 Saving changes... ")
    
    if err := c.BaseCommand.Context.DataService.UpdateCard(cardID, name, number, holder, month, year, cvv, cardType, bank, tags, currentUser.Email); err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
        return err
    }
    
    c.BaseCommand.Printf("✅ Card updated successfully!\n")
    return nil
}