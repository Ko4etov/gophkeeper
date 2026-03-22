// internal/shell/commands/delete/binary.go
package delete

import (
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/helpers"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
	"github.com/Ko4etov/gophkeeper/internal/models"
)

type DeleteBinaryCommand struct {
    types.BaseCommand
}

func NewDeleteBinaryCommand(ctx *types.CommandContext) *DeleteBinaryCommand {
    return &DeleteBinaryCommand{
        BaseCommand: types.BaseCommand{
            NameText:    "binary",
            AliasesText: []string{"file", "bin"},
            HelpText:    "binary <id> - Delete binary file",
            Context:     ctx,
        },
    }
}

func (c *DeleteBinaryCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    if len(args) == 0 {
        c.BaseCommand.Printf("❌ Please specify binary ID\n")
        c.BaseCommand.Printf("Usage: delete binary <id>\n")
        c.BaseCommand.Printf("Example: delete binary 123e4567-e89b-12d3-a456-426614174000\n")
        return nil
    }
    
    binaryID := args[0]
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    
    // Получаем информацию о файле
    entry, err := c.BaseCommand.Context.DataService.GetBinary(binaryID, currentUser.Email)
    if err != nil {
        c.BaseCommand.Printf("❌ Failed to get binary info: %v\n", err)
        return err
    }
    
    // ✅ Правильное приведение типа
    binaryData, ok := entry.Data.(models.BinaryData)
    if !ok {
        c.BaseCommand.Printf("❌ Invalid data type for binary entry\n")
        return nil
    }
    
    c.BaseCommand.Printf("⚠️  You are about to DELETE this file:\n")
    c.BaseCommand.Println(strings.Repeat("─", 50))
    c.BaseCommand.Printf("ID:       %s\n", binaryID)
    c.BaseCommand.Printf("Name:     %s\n", entry.Meta.Name)
    c.BaseCommand.Printf("File:     %s\n", binaryData.Filename)
    c.BaseCommand.Printf("Type:     %s\n", binaryData.MimeType)
    c.BaseCommand.Printf("Size:     %s\n", helpers.FormatFileSize(binaryData.Size))
    c.BaseCommand.Printf("Created:  %s\n", entry.Meta.CreatedAt.Format("2006-01-02 15:04:05"))
    c.BaseCommand.Printf("Updated:  %s\n", entry.Meta.UpdatedAt.Format("2006-01-02 15:04:05"))
    
    // Если есть теги, показываем их
    if len(entry.Meta.Tags) > 0 {
        c.BaseCommand.Printf("Tags:     %s\n", strings.Join(entry.Meta.Tags, ", "))
    }
    c.BaseCommand.Println(strings.Repeat("─", 50))
    
    // Запрашиваем подтверждение
    confirm, err := c.BaseCommand.Confirm("Are you sure you want to delete this file? This cannot be undone")
    if err != nil || !confirm {
        c.BaseCommand.Printf("❌ Delete cancelled\n")
        return nil
    }
    
    c.BaseCommand.Printf("\n🗑️  Deleting file... ")
    
    if err := c.BaseCommand.Context.DataService.Delete(binaryID, currentUser.Email); err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
        return err
    }
    
    c.BaseCommand.Printf("✅ File deleted successfully!\n")
    return nil
}