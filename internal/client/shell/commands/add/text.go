// internal/shell/commands/add/text.go
package add

import (
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type AddTextCommand struct {
    types.BaseCommand
}

func NewAddTextCommand(ctx *types.CommandContext) *AddTextCommand {
    return &AddTextCommand{
        BaseCommand: types.BaseCommand{
            NameText:    "text",
            AliasesText: []string{"txt", "note"},
            HelpText:    "text <name> - Add text data",
            Context:     ctx,
        },
    }
}

func (c *AddTextCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    var name string
    if len(args) > 0 {
        name = args[0]
    } else {
        var err error
        name, err = c.BaseCommand.Prompt("Name for this text: ")
        if err != nil {
            return err
        }
    }
    
    if name == "" {
        c.BaseCommand.Printf("❌ Name is required\n")
		return nil
    }
    
    c.BaseCommand.Printf("📝 Adding text: %s\n", name)
    c.BaseCommand.Println(strings.Repeat("─", 40))
    
    c.BaseCommand.Printf("Available formats: plain, json, markdown, yaml\n")
    format, err := c.BaseCommand.Prompt("Format (default: plain): ")
    if err != nil {
        return err
    }
    if format == "" {
        format = "plain"
    }
    
    validFormats := map[string]bool{
        "plain": true, "json": true, "markdown": true, "yaml": true,
    }
    if !validFormats[format] {
        c.BaseCommand.Printf("❌ Invalid format. Use: plain, json, markdown, yaml\n")
		return nil
    }
    
    c.BaseCommand.Printf("Enter content (end with Ctrl+D on new line):\n")
    c.BaseCommand.Println(strings.Repeat("─", 40))
    
    var contentLines []string
    for {
        line, err := c.BaseCommand.Context.Reader.Readline()
        if err != nil {
            break
        }
        contentLines = append(contentLines, line)
    }
    content := strings.Join(contentLines, "\n")
    
    if content == "" {
        c.BaseCommand.Printf("❌ Content cannot be empty\n")
		return nil
    }
    
    c.BaseCommand.Println(strings.Repeat("─", 40))
    
    tagsStr, _ := c.BaseCommand.Prompt("Tags (comma-separated, optional): ")
    var tags []string
    if tagsStr != "" {
        tags = strings.Split(tagsStr, ",")
        for i, tag := range tags {
            tags[i] = strings.TrimSpace(tag)
        }
    }
    
    c.BaseCommand.Printf("\n💾 Saving text... ")
    
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    if err := c.BaseCommand.Context.DataService.AddText(name, content, format, tags, currentUser.Email); err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
		return err
    }
    
    c.BaseCommand.Printf("✅ Added successfully!\n")
    return nil
}

func (c *AddTextCommand) Completer() func(string) []string {
    return func(line string) []string {
        return []string{"plain", "json", "markdown", "yaml"}
    }
}