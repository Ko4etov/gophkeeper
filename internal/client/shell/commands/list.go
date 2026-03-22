// internal/shell/commands/list.go
package commands

import (
	"fmt"
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type ListCommand struct {
    types.BaseCommand
}

func NewListCommand(ctx *types.CommandContext) *ListCommand {
    return &ListCommand{
        BaseCommand: types.NewSecureCommand(
            "list",
            []string{"ls", "show"},
            "list - List all items",
            ctx,
        ),
    }
}

func (c *ListCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    c.BaseCommand.Printf("🔍 Fetching items... ")
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    items, err := c.BaseCommand.Context.DataService.ListMeta(currentUser.Email)
    for _, item := range items {
        fmt.Printf("%+v\n", item)
    }
    if err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
		return err 
    }
    
    if len(items) == 0 {
        c.BaseCommand.Printf("\n📭 No items found\n")
        return nil
    }
    
    c.BaseCommand.Printf("✅ Found %d item(s)\n\n", len(items))
    
    // Заголовок таблицы
    c.BaseCommand.Printf("%-30s %-25s %-12s %-15s %s\n",
        "ID", "NAME", "TYPE", "TAGS", "UPDATED")
    c.BaseCommand.Printf("%-30s %-25s %-12s %-15s %s\n",
        strings.Repeat("─", 30),
        strings.Repeat("─", 25),
        strings.Repeat("─", 15),
        strings.Repeat("─", 15),
        strings.Repeat("─", 10))
    
    for _, item := range items {
        // Обрабатываем теги
        tags := strings.Join(item.Tags, ",")
        if len(tags) > 15 {
            tags = tags[:12] + "..."
        }
        
        // Обрабатываем имя
        name := item.Name
        if len(name) > 25 {
            name = name[:22] + "..."
        }
        
        c.BaseCommand.Printf("%-30s %-25s %-12s %-15s %s\n",
            item.ID,
            name,
            item.DataType,
            tags,
            item.UpdatedAt.Format("2006-01-02"),
        )
    }
    
    c.BaseCommand.Printf("\n💡 Use 'get <id>' to see details\n")
    return nil
}

func (c *ListCommand) Completer() func(string) []string {
    return func(line string) []string {
        return []string{"login", "text", "card", "all"}
    }
}