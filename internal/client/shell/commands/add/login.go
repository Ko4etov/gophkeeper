// internal/shell/commands/add/login.go
package add

import (
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type AddLoginCommand struct {
    types.BaseCommand
}

func NewAddLoginCommand(ctx *types.CommandContext) *AddLoginCommand {
    return &AddLoginCommand{
        BaseCommand: types.BaseCommand{
            NameText:    "login",
            AliasesText: []string{"lp", "password"},
            HelpText:    "login <name> - Add a login/password pair",
            Context:     ctx,
        },
    }
}

func (c *AddLoginCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    var name string
    if len(args) > 0 {
        name = args[0]
    } else {
        var err error
        name, err = c.BaseCommand.Prompt("Name for this login: ")
        if err != nil {
            return err
        }
    }
    
    if name == "" {
        c.BaseCommand.Printf("❌ Name is required\n")
		return nil
    }
    
    c.BaseCommand.Printf("📝 Adding login/password: %s\n", name)
    c.BaseCommand.Println(strings.Repeat("─", 40))
    
    login, err := c.BaseCommand.Prompt("Login: ")
    if err != nil {
        return err
    }
    
    password, err := c.BaseCommand.PromptPassword("Password: ")
    if err != nil {
        return err
    }
    
    url, _ := c.BaseCommand.Prompt("URL (optional): ")

    notes, _ := c.BaseCommand.Prompt("Notes (optional): ")
    
    tagsStr, _ := c.BaseCommand.Prompt("Tags (comma-separated, optional): ")
    var tags []string
    if tagsStr != "" {
        tags = strings.Split(tagsStr, ",")
        for i, tag := range tags {
            tags[i] = strings.TrimSpace(tag)
        }
    }
    
    c.BaseCommand.Printf("\n💾 Saving... ")
    
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    if err := c.BaseCommand.Context.DataService.AddLogin(name, login, password, url, notes, tags, currentUser.Email); err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
		return err
    }
    
    c.BaseCommand.Printf("✅ Added successfully!\n")
    return nil
}