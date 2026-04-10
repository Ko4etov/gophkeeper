// internal/shell/commands/exit.go
package commands

import (
	"os"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type ExitCommand struct {
    types.BaseCommand
}

func NewExitCommand(ctx *types.CommandContext) *ExitCommand {
    return &ExitCommand{
        BaseCommand: types.NewPublicCommand(
            "exit",
            []string{"quit", "q", "bye"},
            "exit - Exit the program",
            ctx,
        ),
    }
}

func (c *ExitCommand) Execute(args []string) error {
    c.BaseCommand.Printf("👋 Goodbye!\n")
    os.Exit(0)
    return nil
}