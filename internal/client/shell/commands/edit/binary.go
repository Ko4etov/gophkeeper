// internal/shell/commands/edit/binary.go
package edit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/helpers"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
	"github.com/Ko4etov/gophkeeper/internal/models"
)

type EditBinaryCommand struct {
    types.BaseCommand
}

func NewEditBinaryCommand(ctx *types.CommandContext) *EditBinaryCommand {
    return &EditBinaryCommand{
        BaseCommand: types.BaseCommand{
            NameText:    "binary",
            AliasesText: []string{"file", "bin"},
            HelpText:    "binary <id> - Edit binary file metadata or replace content",
            Context:     ctx,
        },
    }
}

func (c *EditBinaryCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    if len(args) == 0 {
        c.BaseCommand.Printf("❌ Please specify binary ID\n")
        c.BaseCommand.Printf("Usage: edit binary <id>\n")
        c.BaseCommand.Printf("Example: edit binary 123e4567-e89b-12d3-a456-426614174000\n")
        return nil
    }
    
    binaryID := args[0]
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    
    // Получаем текущие данные
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
    
    c.BaseCommand.Printf("📎 Editing binary file: %s\n", entry.Meta.Name)
    c.BaseCommand.Println(strings.Repeat("─", 50))
    c.BaseCommand.Printf("Current file: %s\n", binaryData.Filename)
    c.BaseCommand.Printf("File size: %s\n", helpers.FormatFileSize(binaryData.Size))
    c.BaseCommand.Printf("MIME type: %s\n", binaryData.MimeType)
    c.BaseCommand.Println(strings.Repeat("─", 50))
    
    // Спрашиваем, хочет ли пользователь заменить файл
    replaceFile, err := c.BaseCommand.Confirm("Do you want to replace the file content?")
    if err != nil {
        return err
    }
    
    var newContent []byte
    var newFilename string
    var newMimeType string
    var newSize int64
    
    if replaceFile {
        // Запрашиваем путь к новому файлу
        filePath, err := c.BaseCommand.Prompt("New file path: ")
        if err != nil {
            return err
        }
        
        filePath = strings.TrimSpace(filePath)
        if filePath == "" {
            c.BaseCommand.Printf("❌ File path is required\n")
            return nil
        }
        
        // Проверяем существование файла
        info, err := os.Stat(filePath)
        if err != nil {
            if os.IsNotExist(err) {
                c.BaseCommand.Printf("❌ File not found: %s\n", filePath)
            } else {
                c.BaseCommand.Printf("❌ Error accessing file: %v\n", err)
            }
            return nil
        }
        
        if info.IsDir() {
            c.BaseCommand.Printf("❌ Path is a directory, expected a file\n")
            return nil
        }
        
        // Проверяем размер
        newSize = info.Size()
        if newSize > helpers.MaxFileSize {
            c.BaseCommand.Printf("❌ File size (%.2f MB) exceeds maximum allowed size (5 MB)\n", 
                float64(newSize)/(1024*1024))
            return nil
        }
        
        // Читаем новый файл
        c.BaseCommand.Printf("📖 Reading new file... ")
        newContent, err = os.ReadFile(filePath)
        if err != nil {
            c.BaseCommand.Printf("❌ Failed to read file: %v\n", err)
            return err
        }
        c.BaseCommand.Printf("✅ (%s)\n", helpers.FormatFileSize(int64(len(newContent))))
        
        newFilename = filepath.Base(filePath)
        newMimeType = helpers.DetectMimeType(filePath)
        
        c.BaseCommand.Printf("New file: %s (%s)\n", newFilename, helpers.FormatFileSize(newSize))
        if newMimeType != binaryData.MimeType {
            c.BaseCommand.Printf("📝 MIME type will change: %s → %s\n", binaryData.MimeType, newMimeType)
        }
        c.BaseCommand.Println(strings.Repeat("─", 50))
    }
    
    // Показываем текущие значения и запрашиваем новые
    c.BaseCommand.Printf("Press Enter to keep current value, or enter new value:\n\n")
    
    // Имя записи
    namePrompt := fmt.Sprintf("Name [%s]: ", entry.Meta.Name)
    name, err := c.BaseCommand.Prompt(namePrompt)
    if err != nil {
        return err
    }
    if name == "" {
        name = entry.Meta.Name
    }
    
    // Имя файла (отображаемое)
    var filename string
    if replaceFile {
        filenamePrompt := fmt.Sprintf("Display filename [%s]: ", newFilename)
        filename, err = c.BaseCommand.Prompt(filenamePrompt)
        if err != nil {
            return err
        }
        if filename == "" {
            filename = newFilename
        }
    } else {
        filenamePrompt := fmt.Sprintf("Display filename [%s]: ", binaryData.Filename)
        filename, err = c.BaseCommand.Prompt(filenamePrompt)
        if err != nil {
            return err
        }
        if filename == "" {
            filename = binaryData.Filename
        }
    }
    
    // MIME тип
    var mimeType string
    if replaceFile && newMimeType != "" {
        mimeTypePrompt := fmt.Sprintf("MIME type [%s]: ", newMimeType)
        mimeType, err = c.BaseCommand.Prompt(mimeTypePrompt)
        if err != nil {
            return err
        }
        if mimeType == "" {
            mimeType = newMimeType
        }
    } else {
        mimeTypePrompt := fmt.Sprintf("MIME type [%s]: ", binaryData.MimeType)
        mimeType, err = c.BaseCommand.Prompt(mimeTypePrompt)
        if err != nil {
            return err
        }
        if mimeType == "" {
            mimeType = binaryData.MimeType
        }
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
    
    // Показываем сводку изменений
    c.BaseCommand.Println(strings.Repeat("─", 50))
    c.BaseCommand.Printf("Summary of changes:\n")
    c.BaseCommand.Printf("  Name:     %s → %s\n", entry.Meta.Name, name)
    c.BaseCommand.Printf("  Filename: %s → %s\n", binaryData.Filename, filename)
    if mimeType != binaryData.MimeType {
        c.BaseCommand.Printf("  MIME:     %s → %s\n", binaryData.MimeType, mimeType)
    }
    if replaceFile {
        c.BaseCommand.Printf("  Content:  WILL BE REPLACED (new size: %s)\n", helpers.FormatFileSize(newSize))
    }
    c.BaseCommand.Println(strings.Repeat("─", 50))
    
    // Подтверждение
    confirm, err := c.BaseCommand.Confirm("Save changes?")
    if err != nil || !confirm {
        c.BaseCommand.Printf("❌ Edit cancelled\n")
        return nil
    }
    
    c.BaseCommand.Printf("\n💾 Saving changes... ")
    
    // Обновляем данные
    if replaceFile {
        if err := c.BaseCommand.Context.DataService.UpdateBinaryFull(
            binaryID, name, filename, newContent, mimeType, tags, currentUser.Email,
        ); err != nil {
            c.BaseCommand.Printf("❌ Failed: %v\n", err)
            return err
        }
    } else {
        if err := c.BaseCommand.Context.DataService.UpdateBinaryMetadata(
            binaryID, name, filename, mimeType, tags, currentUser.Email,
        ); err != nil {
            c.BaseCommand.Printf("❌ Failed: %v\n", err)
            return err
        }
    }
    
    c.BaseCommand.Printf("✅ Binary updated successfully!\n")
    if replaceFile {
        c.BaseCommand.Printf("📎 %s (%s)\n", filename, helpers.FormatFileSize(newSize))
    }
    
    return nil
}