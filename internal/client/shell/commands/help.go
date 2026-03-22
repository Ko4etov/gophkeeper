// internal/shell/commands/help.go
package commands

import (
	"sort"
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type HelpCommand struct {
    types.BaseCommand
    registry *Registry
}

func NewHelpCommand(ctx *types.CommandContext, r *Registry) *HelpCommand {
    return &HelpCommand{
        BaseCommand: types.NewPublicCommand(
            "help",
            []string{"h", "?"},
            "help [command] - Show help for commands",
            ctx,
        ),
        registry: r,
    }
}

func (c *HelpCommand) Execute(args []string) error {
    if len(args) > 0 {
        return c.showCommandHelp(args[0])
    }
    
    c.BaseCommand.Println("\n📚 Available commands:")
    c.BaseCommand.Println(strings.Repeat("─", 50))
    
    var cmds []types.Command
    for _, cmd := range c.registry.GetAll() {
        if cmd.Name() != "help" {
            cmds = append(cmds, cmd)
        }
    }
    
    sort.Slice(cmds, func(i, j int) bool {
        return cmds[i].Name() < cmds[j].Name()
    })
    
    for _, cmd := range cmds {
        c.BaseCommand.Printf("  %-15s %s\n", cmd.Name(), cmd.Help())
    }
    
    c.BaseCommand.Println("\nUse 'help <command>' for detailed help on a specific command.")
    c.BaseCommand.Println("Use TAB for autocompletion, ↑/↓ for history.")
    
    return nil
}

func (c *HelpCommand) showCommandHelp(cmdName string) error {
    cmd := c.registry.Get(cmdName)
    if cmd == nil {
        c.BaseCommand.Printf("Unknown command: %s\n", cmdName)
		return nil
    }
    
    c.BaseCommand.Printf("\n📘 Help for '%s':\n", cmd.Name())
    c.BaseCommand.Println(strings.Repeat("─", 50))
    c.BaseCommand.Println(cmd.Help())
    
    if parent, ok := cmd.(types.ParentCommand); ok && len(parent.SubCommands()) > 0 {
        c.BaseCommand.Println("\nSubcommands:")
        for _, sub := range parent.SubCommands() {
            c.BaseCommand.Printf("  %-15s %s\n", sub.Name(), sub.Help())
        }
    }
    
    if len(cmd.Aliases()) > 0 {
        c.BaseCommand.Printf("\nAliases: %s\n", strings.Join(cmd.Aliases(), ", "))
    }
    
    return nil
}