// internal/shell/commands/edit/edit.go
package edit

import (
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type EditCommand struct {
    types.BaseCommand
    subCommands map[string]types.Command
}

func NewEditCommand(ctx *types.CommandContext) *EditCommand {
    cmd := &EditCommand{
        BaseCommand: types.NewSecureCommand(
            "edit",
            []string{"update", "modify"},
            "edit <command> - Edit existing data",
            ctx,
        ),
        subCommands: make(map[string]types.Command),
    }
    
    // Регистрируем подкоманды
    cmd.registerSubCommand(NewEditCardCommand(ctx))
    cmd.registerSubCommand(NewEditLoginCommand(ctx))
    cmd.registerSubCommand(NewEditTextCommand(ctx))
    cmd.registerSubCommand(NewEditBinaryCommand(ctx))
    
    return cmd
}

func (c *EditCommand) registerSubCommand(cmd types.Command) {
    c.subCommands[cmd.Name()] = cmd
    for _, alias := range cmd.Aliases() {
        c.subCommands[alias] = cmd
    }
}

func (c *EditCommand) Execute(args []string) error {
    if len(args) == 0 {
        c.Printf("Available edit commands:\n")
        for name := range c.subCommands {
            c.Printf("  edit %s\n", name)
        }
        return nil
    }
    
    subCmd := c.subCommands[args[0]]
    if subCmd == nil {
        c.Printf("❌ Unknown edit command: %s\n", args[0])
        return nil
    }
    
    return subCmd.Execute(args[1:])
}

func (c *EditCommand) SubCommands() []types.Command {
    var cmds []types.Command
    seen := make(map[string]bool)
    for _, cmd := range c.subCommands {
        if !seen[cmd.Name()] {
            seen[cmd.Name()] = true
            cmds = append(cmds, cmd)
        }
    }
    return cmds
}

func (c *EditCommand) GetSubCommand(name string) types.Command {
    return c.subCommands[name]
}