// internal/shell/commands/delete/text.go
package delete

import (
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type DeleteTextCommand struct {
    types.BaseCommand
}

func NewDeleteTextCommand(ctx *types.CommandContext) *DeleteTextCommand {
    return &DeleteTextCommand{
        BaseCommand: types.BaseCommand{
            NameText:    "text",
            AliasesText: []string{"txt", "note", "document"},
            HelpText:    "text <id> - Delete text data",
            Context:     ctx,
        },
    }
}

func (c *DeleteTextCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    if len(args) == 0 {
        c.BaseCommand.Printf("❌ Please specify text ID\n")
        c.BaseCommand.Printf("Usage: delete text <id>\n")
        c.BaseCommand.Printf("Example: delete text 123e4567-e89b-12d3-a456-426614174000\n")
        return nil
    }
    
    textID := args[0]
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    
    // Получаем информацию о тексте для подтверждения
    entry, err := c.BaseCommand.Context.DataService.GetText(textID, currentUser.Email)
    if err != nil {
        c.BaseCommand.Printf("❌ Failed to get text: %v\n", err)
        return err
    }
    
    // Извлекаем данные
    textData := entry.Data.(map[string]interface{})
    content := textData["content"].(string)
    format := textData["format"].(string)
    
    // Показываем предпросмотр
    c.BaseCommand.Printf("⚠️  You are about to DELETE this text:\n")
    c.BaseCommand.Println(strings.Repeat("─", 50))
    c.BaseCommand.Printf("ID:       %s\n", textID)
    c.BaseCommand.Printf("Name:     %s\n", entry.Meta.Name)
    c.BaseCommand.Printf("Format:   %s\n", format)
    c.BaseCommand.Printf("Created:  %s\n", entry.Meta.CreatedAt.Format("2006-01-02 15:04"))
    c.BaseCommand.Printf("Updated:  %s\n", entry.Meta.UpdatedAt.Format("2006-01-02 15:04"))
    
    // Показываем превью контента (первые 200 символов)
    preview := content
    if len(preview) > 200 {
        preview = preview[:200] + "..."
    }
    c.BaseCommand.Printf("\n📄 Content preview:\n")
    c.BaseCommand.Println(strings.Repeat("─", 50))
    c.BaseCommand.Printf("%s\n", preview)
    c.BaseCommand.Println(strings.Repeat("─", 50))
    
    // Если есть теги, показываем их
    if len(entry.Meta.Tags) > 0 {
        c.BaseCommand.Printf("Tags:     %s\n", strings.Join(entry.Meta.Tags, ", "))
        c.BaseCommand.Println(strings.Repeat("─", 50))
    }
    
    // Запрашиваем подтверждение
    confirm, err := c.BaseCommand.Confirm("Are you sure you want to delete this text? This cannot be undone")
    if err != nil || !confirm {
        c.BaseCommand.Printf("❌ Delete cancelled\n")
        return nil
    }
    
    // Второе подтверждение для важных данных
    if len(content) > 500 {
        c.BaseCommand.Printf("\n⚠️  This is a large text (%d characters). ", len(content))
        doubleConfirm, err := c.BaseCommand.Confirm("Are you REALLY sure?")
        if err != nil || !doubleConfirm {
            c.BaseCommand.Printf("❌ Delete cancelled\n")
            return nil
        }
    }
    
    c.BaseCommand.Printf("\n🗑️  Deleting text... ")
    
    if err := c.BaseCommand.Context.DataService.Delete(textID, currentUser.Email); err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
        return err
    }
    
    c.BaseCommand.Printf("✅ Text deleted successfully!\n")
    return nil
}

// GetTextPreview возвращает краткое содержимое для отображения в списке
func (c *DeleteTextCommand) GetTextPreview(content string, maxLen int) string {
    if len(content) <= maxLen {
        return content
    }
    
    // Обрезаем по словам, чтобы не разрывать текст
    truncated := content[:maxLen]
    lastSpace := strings.LastIndex(truncated, " ")
    if lastSpace > maxLen/2 {
        truncated = truncated[:lastSpace]
    }
    
    return truncated + "..."
}

// Completer возвращает возможные варианты для автодополнения
func (c *DeleteTextCommand) Completer() func(string) []string {
    return func(line string) []string {
        // Здесь можно добавить автодополнение ID текстов
        // Например, получить список недавних текстов
        return []string{}
    }
}