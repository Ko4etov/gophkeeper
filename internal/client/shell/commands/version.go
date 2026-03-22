// internal/shell/commands/version.go
package commands

import (
	"runtime"
	"strings"

	grpcclient "github.com/Ko4etov/gophkeeper/internal/client/grpc_client"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

type VersionCommand struct {
    types.BaseCommand
}

func NewVersionCommand(ctx *types.CommandContext) *VersionCommand {
    return &VersionCommand{
        BaseCommand: types.NewPublicCommand(
            "version",
            []string{"ver", "v", "--version"},
            "version - Show version information",
            ctx,
        ),
    }
}

func (c *VersionCommand) Execute(args []string) error {
    buildInfo := c.BaseCommand.Context.BuildInfo
    if buildInfo == nil {
        buildInfo = &grpcclient.BuildInfo{
            Version:   "dev",
            BuildDate: "unknown",
        }
    }
    
    c.BaseCommand.Println("🔐 GophKeeper")
    c.BaseCommand.Println(strings.Repeat("─", 40))
    
    c.BaseCommand.Printf("📦 Version: %s\n", buildInfo.Version)
    c.BaseCommand.Printf("📅 Build date: %s\n", buildInfo.BuildDate)
    
    c.BaseCommand.Printf("🖥️  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
    c.BaseCommand.Printf("☕ Go version: %s\n", runtime.Version())
    
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    if currentUser != nil {
        c.BaseCommand.Printf("\n👤 Logged in as: %s\n", currentUser.Email)
    } else {
        c.BaseCommand.Printf("\n👤 Not logged in\n")
    }
    
    if c.BaseCommand.Context.Config != nil {
        c.BaseCommand.Printf("🌐 Server: %s\n", c.BaseCommand.Context.Config.ServerAddress)
    }
    
    if len(args) > 0 && args[0] == "--verbose" {
        c.BaseCommand.Println("\n📋 Detailed information:")
        c.BaseCommand.Printf("   Data directory: %s\n", c.BaseCommand.Context.Config.DataDir)
        c.BaseCommand.Printf("   History file: %s\n", c.BaseCommand.Context.Config.HistoryFile)
    }
    
    return nil
}