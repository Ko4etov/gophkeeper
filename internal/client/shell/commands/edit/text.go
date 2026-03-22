package edit

import (
	"fmt"
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
	"github.com/Ko4etov/gophkeeper/internal/models"
)

type EditTextCommand struct {
    types.BaseCommand
}

func NewEditTextCommand(ctx *types.CommandContext) *EditTextCommand {
    return &EditTextCommand{
        BaseCommand: types.BaseCommand{
            NameText:    "text",
            AliasesText: []string{"txt", "note", "document"},
            HelpText:    "text <id> - Edit text data",
            Context:     ctx,
        },
    }
}

func (c *EditTextCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    if len(args) == 0 {
        c.BaseCommand.Printf("❌ Please specify text ID\n")
        c.BaseCommand.Printf("Usage: edit text <id>\n")
        return nil
    }
    
    textID := args[0]
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    
    // Получаем текущие данные
    entry, err := c.BaseCommand.Context.DataService.GetText(textID, currentUser.Email)
    if err != nil {
        c.BaseCommand.Printf("❌ Failed to get text: %v\n", err)
        return err
    }
    
    textData, ok := entry.Data.(models.TextData)
    if !ok {
        c.BaseCommand.Printf("❌ Invalid data type for text entry\n")
        return nil
    }
    
    currentContent := textData.Content
    currentFormat := textData.Format
    
    c.BaseCommand.Printf("📝 Editing text: %s\n", entry.Meta.Name)
    c.BaseCommand.Println(strings.Repeat("─", 50))
    
    namePrompt := fmt.Sprintf("Name [%s]: ", entry.Meta.Name)
    name, err := c.BaseCommand.Prompt(namePrompt)
    if err != nil {
        return err
    }
    if name == "" {
        name = entry.Meta.Name
    }
    
    c.BaseCommand.Printf("Available formats: plain, json, markdown, yaml\n")
    formatPrompt := fmt.Sprintf("Format [%s]: ", currentFormat)
    format, err := c.BaseCommand.Prompt(formatPrompt)
    if err != nil {
        return err
    }
    if format == "" {
        format = currentFormat
    }
    
    validFormats := map[string]bool{
        "plain": true, "json": true, "markdown": true, "yaml": true,
    }
    if !validFormats[format] {
        c.BaseCommand.Printf("❌ Invalid format. Use: plain, json, markdown, yaml\n")
        return nil
    }
    
    c.BaseCommand.Printf("\nCurrent content (%d characters):\n", len(currentContent))
    c.BaseCommand.Println(strings.Repeat("─", 50))
    preview := currentContent
    if len(preview) > 500 {
        preview = preview[:500] + "..."
    }
    c.BaseCommand.Printf("%s\n", preview)
    c.BaseCommand.Println(strings.Repeat("─", 50))
    
    changeContent, err := c.BaseCommand.Confirm("Do you want to change the content?")
    if err != nil {
        return err
    }
    
    content := currentContent
    if changeContent {
        c.BaseCommand.Printf("Enter new content (end with Ctrl+D on new line):\n")
        c.BaseCommand.Println(strings.Repeat("─", 50))
        
        var contentLines []string
        for {
            line, err := c.BaseCommand.Context.Reader.Readline()
            if err != nil {
                break
            }
            contentLines = append(contentLines, line)
        }
        newContent := strings.Join(contentLines, "\n")
        if newContent != "" {
            content = newContent
        }
    }
    
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
    
    confirm, err := c.BaseCommand.Confirm("Save changes?")
    if err != nil || !confirm {
        c.BaseCommand.Printf("❌ Edit cancelled\n")
        return nil
    }
    
    c.BaseCommand.Printf("\n💾 Saving changes... ")
    
    if err := c.BaseCommand.Context.DataService.UpdateText(
        textID, name, content, format, tags, currentUser.Email,
    ); err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
        return err
    }
    
    c.BaseCommand.Printf("✅ Text updated successfully!\n")
    return nil
}