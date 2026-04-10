// internal/shell/commands/types/types.go
package types

import (
	"fmt"
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/config"
	"github.com/Ko4etov/gophkeeper/internal/client/crypto"
	grpcclient "github.com/Ko4etov/gophkeeper/internal/client/grpc_client"
	"github.com/Ko4etov/gophkeeper/internal/client/service/auth"
	"github.com/Ko4etov/gophkeeper/internal/client/service/data"
	"github.com/Ko4etov/gophkeeper/internal/client/service/sync"
	"github.com/chzyer/readline"
)

// CommandContext содержит все зависимости, необходимые командам
type CommandContext struct {
	AuthService *auth.AuthService
	DataService *data.DataService
	SyncManager *sync.SyncManager
	Config      *config.ClientConfig
	Reader      *readline.Instance
	BuildInfo   *grpcclient.BuildInfo
	Session     *crypto.Session
}

// Command определяет интерфейс команды
type Command interface {
	Name() string
	Aliases() []string
	Execute(args []string) error
	Help() string
	Completer() func(string) []string
	RequiresUnlock() bool
}

// ParentCommand определяет интерфейс для команд с подкомандами
type ParentCommand interface {
	Command
	SubCommands() []Command
	GetSubCommand(name string) Command
}

// BaseCommand предоставляет базовую реализацию
type BaseCommand struct {
	NameText           string
	AliasesText        []string
	HelpText           string
	Context            *CommandContext
	RequiresUnlockFlag bool
}

func NewSecureCommand(name string, aliases []string, help string, ctx *CommandContext) BaseCommand {
	return BaseCommand{
		NameText:           name,
		HelpText:           help,
		AliasesText:        aliases,
		Context:            ctx,
		RequiresUnlockFlag: true,
	}
}

// Конструктор для публичных команд (не требуют разблокировки)
func NewPublicCommand(name string, aliases []string, help string, ctx *CommandContext) BaseCommand {
	return BaseCommand{
		NameText:           name,
		HelpText:           help,
		AliasesText:        aliases,
		Context:            ctx,
		RequiresUnlockFlag: false,
	}
}

func (c *BaseCommand) RequiresUnlock() bool {
	return c.RequiresUnlockFlag
}

func (c *BaseCommand) Name() string {
	return c.NameText
}

func (c *BaseCommand) Aliases() []string {
	return c.AliasesText
}

func (c *BaseCommand) Help() string {
	return c.HelpText
}

func (c *BaseCommand) Completer() func(string) []string {
	return nil
}

func (c *BaseCommand) Printf(format string, args ...interface{}) {
	if c.Context != nil && c.Context.Reader != nil {
		fmt.Fprintf(c.Context.Reader.Stderr(), format, args...)
	}
}

func (c *BaseCommand) Println(a ...interface{}) {
	if c.Context != nil && c.Context.Reader != nil {
		fmt.Fprintln(c.Context.Reader.Stderr(), a...)
	}
}

func (c *BaseCommand) CheckLoggedIn() error {
	if c.Context == nil || c.Context.AuthService.GetUser() == nil {
		return fmt.Errorf("you must be logged in first")
	}
	return nil
}

// Prompt запрашивает ввод с приглашением
func (c *BaseCommand) Prompt(prompt string) (string, error) {
	c.Context.Reader.SetPrompt(prompt)

	line, err := c.Context.Reader.Readline()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

// PromptPassword запрашивает ввод пароля (без эха)
func (c *BaseCommand) PromptPassword(prompt string) (string, error) {
	fmt.Fprint(c.Context.Reader.Stderr(), prompt)
	password, err := readline.Password(prompt)
	fmt.Fprintln(c.Context.Reader.Stderr())
	return string(password), err
}

// Confirm запрашивает подтверждение
func (c *BaseCommand) Confirm(prompt string) (bool, error) {
	answer, err := c.Prompt(prompt + " (y/n): ")
	if err != nil {
		return false, err
	}
	answer = strings.ToLower(answer)
	return answer == "y" || answer == "yes", nil
}

// GetVersion возвращает строку с версией
func (c *BaseCommand) GetVersion() string {
	if c.Context.BuildInfo != nil {
		return c.Context.BuildInfo.Version
	}
	return "dev"
}
