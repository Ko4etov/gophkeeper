// internal/shell/commands/add/add.go
package add

import "github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"




type AddCommand struct {
    types.BaseCommand
    subCommands map[string]types.Command
}

func NewAddCommand(ctx *types.CommandContext) *AddCommand {
    cmd := &AddCommand{
        BaseCommand: types.NewSecureCommand(
            "add",
            []string{"create", "new"},
            "add <type> <name> - Add new data (types: login, text, card)",
            ctx,
        ),
        subCommands: make(map[string]types.Command),
    }
    
    // Регистрируем дочерние команды
    cmd.registerSubCommand(NewAddLoginCommand(ctx))
    cmd.registerSubCommand(NewAddTextCommand(ctx))
    cmd.registerSubCommand(NewAddCardCommand(ctx))
    cmd.registerSubCommand(NewAddBinaryCommand(ctx))
    
    return cmd
}

func (c *AddCommand) registerSubCommand(cmd types.Command) {
    c.subCommands[cmd.Name()] = cmd
    for _, alias := range cmd.Aliases() {
        c.subCommands[alias] = cmd
    }
}

func (c *AddCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    if len(args) == 0 {
        c.BaseCommand.Printf("❌ Usage: %s\n", c.BaseCommand.Help())
		return nil
    }
    
    subCmd := c.subCommands[args[0]]
    if subCmd == nil {
        c.BaseCommand.Printf("❌ Unknown type: %s\n", args[0])
		return nil
    }
    
    return subCmd.Execute(args[1:])
}

func (c *AddCommand) SubCommands() []types.Command {
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

func (c *AddCommand) GetSubCommand(name string) types.Command {
    return c.subCommands[name]
}