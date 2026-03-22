// internal/shell/commands/add/binary.go
package add

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/helpers"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type AddBinaryCommand struct {
    types.BaseCommand
}

func NewAddBinaryCommand(ctx *types.CommandContext) *AddBinaryCommand {
    return &AddBinaryCommand{
        BaseCommand: types.BaseCommand{
            NameText:    "binary",
            AliasesText: []string{"file", "bin", "upload"},
            HelpText:    "binary <filepath> - Add binary file data (max 5MB)",
            Context:     ctx,
        },
    }
}

func (c *AddBinaryCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    var filePath string
    if len(args) > 0 {
        filePath = args[0]
    } else {
        var err error
        filePath, err = c.BaseCommand.Prompt("File path: ")
        if err != nil {
            return err
        }
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
    
    // Проверяем, что это файл, а не директория
    if info.IsDir() {
        c.BaseCommand.Printf("❌ Path is a directory, expected a file\n")
        return nil
    }
    
    // ✅ Проверяем размер файла
    fileSize := info.Size()
    if fileSize > helpers.MaxFileSize {
        c.BaseCommand.Printf("❌ File size (%.2f MB) exceeds maximum allowed size (5 MB)\n", 
            float64(fileSize)/(1024*1024))
        return nil
    }
    
    // Получаем имя файла (без пути)
    fileName := filepath.Base(filePath)
    
    c.BaseCommand.Printf("📎 Adding binary file: %s\n", fileName)
    c.BaseCommand.Printf("📦 Size: %s (max 5 MB)\n", helpers.FormatFileSize(fileSize))
    c.BaseCommand.Println(strings.Repeat("─", 40))
    
    // Имя записи (можно использовать имя файла или запросить свое)
    namePrompt := fmt.Sprintf("Name for this file (default: %s): ", fileName)
    name, err := c.BaseCommand.Prompt(namePrompt)
    if err != nil {
        return err
    }
    if name == "" {
        name = fileName
    }
    
    // MIME тип (опционально)
    mimeType, _ := c.BaseCommand.Prompt("MIME type (optional, auto-detect if empty): ")
    if mimeType == "" {
        mimeType = detectMimeType(filePath)
        if mimeType != "" {
            c.BaseCommand.Printf("🔍 Detected MIME type: %s\n", mimeType)
        }
    }
    
    // Теги
    tagsStr, _ := c.BaseCommand.Prompt("Tags (comma-separated, optional): ")
    var tags []string
    if tagsStr != "" {
        tags = strings.Split(tagsStr, ",")
        for i, tag := range tags {
            tags[i] = strings.TrimSpace(tag)
        }
    }
    
    // Читаем файл
    c.BaseCommand.Printf("\n📖 Reading file... ")
    
    fileContent, err := os.ReadFile(filePath)
    if err != nil {
        c.BaseCommand.Printf("❌ Failed to read file: %v\n", err)
        return err
    }
    
    c.BaseCommand.Printf("✅ (%s)\n", helpers.FormatFileSize(int64(len(fileContent))))
    
    // Подтверждение
    c.BaseCommand.Printf("\n💾 Saving file... ")
    
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    if err := c.BaseCommand.Context.DataService.AddBinary(
        name, fileName, fileContent, mimeType, tags, currentUser.Email,
    ); err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
        return err
    }
    
    c.BaseCommand.Printf("✅ Added successfully!\n")
    c.BaseCommand.Printf("📎 %s (%s)\n", fileName, helpers.FormatFileSize(fileSize))
    
    return nil
}

// detectMimeType определяет MIME тип файла по расширению
func detectMimeType(filePath string) string {
    ext := strings.ToLower(filepath.Ext(filePath))
    
    mimeTypes := map[string]string{
        ".txt":  "text/plain",
        ".md":   "text/markdown",
        ".json": "application/json",
        ".yaml": "application/yaml",
        ".yml":  "application/yaml",
        ".xml":  "application/xml",
        ".html": "text/html",
        ".htm":  "text/html",
        ".css":  "text/css",
        ".js":   "application/javascript",
        ".pdf":  "application/pdf",
        ".jpg":  "image/jpeg",
        ".jpeg": "image/jpeg",
        ".png":  "image/png",
        ".gif":  "image/gif",
        ".webp": "image/webp",
        ".svg":  "image/svg+xml",
        ".zip":  "application/zip",
        ".tar":  "application/x-tar",
        ".gz":   "application/gzip",
        ".7z":   "application/x-7z-compressed",
        ".mp3":  "audio/mpeg",
        ".mp4":  "video/mp4",
        ".mov":  "video/quicktime",
        ".avi":  "video/x-msvideo",
    }
    
    if mime, ok := mimeTypes[ext]; ok {
        return mime
    }
    
    return "application/octet-stream"
}