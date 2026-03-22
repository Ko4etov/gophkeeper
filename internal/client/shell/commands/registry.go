// internal/shell/commands/registry.go
package commands

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/config"
	"github.com/Ko4etov/gophkeeper/internal/client/crypto"
	grpcclient "github.com/Ko4etov/gophkeeper/internal/client/grpc_client"
	"github.com/Ko4etov/gophkeeper/internal/client/service/auth"
	"github.com/Ko4etov/gophkeeper/internal/client/service/data"
	"github.com/Ko4etov/gophkeeper/internal/client/service/sync"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/add"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/delete"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/edit"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/wrappers"
	"github.com/chzyer/readline"
)

// Registry хранит все зарегистрированные команды
type Registry struct {
	commands map[string]types.Command
	ctx      *types.CommandContext
	session  *crypto.Session
}

// NewRegistry создает новый реестр команд
func NewRegistry(
	dataService *data.DataService,
	authService *auth.AuthService,
	syncManager *sync.SyncManager,
	cfg *config.ClientConfig,
	rl *readline.Instance,
	buildInfo *grpcclient.BuildInfo,
	session *crypto.Session,
) *Registry {

	ctx := &types.CommandContext{
		DataService: dataService,
		AuthService: authService,
		SyncManager: syncManager,
		Config:      cfg,
		Reader:      rl,
		BuildInfo:   buildInfo,
		Session:     session,
	}

	r := &Registry{
		commands: make(map[string]types.Command),
		ctx:      ctx,
		session:  session,
	}

	unlockFunc := func() error {
		return r.promptUnlock()
	}

	// Регистрируем все команды
	r.Register(NewHelpCommand(ctx, r))
	r.Register(NewExitCommand(ctx))
	r.Register(NewVersionCommand(ctx))
	r.Register(NewRegisterCommand(ctx))
	r.Register(NewLoginCommand(ctx))

	r.Register(wrappers.NewSecureCommandWrapper(NewListCommand(ctx), session, unlockFunc))
	r.Register(wrappers.NewSecureCommandWrapper(NewGetCommand(ctx), session, unlockFunc))
	r.Register(wrappers.NewSecureCommandWrapper(NewSyncCommand(ctx), session, unlockFunc))

	// Регистрируем составные команды
	r.Register(wrappers.NewSecureCommandWrapper(add.NewAddCommand(ctx), session, unlockFunc))
	r.Register(wrappers.NewSecureCommandWrapper(edit.NewEditCommand(ctx), session, unlockFunc))
	r.Register(wrappers.NewSecureCommandWrapper(delete.NewDeleteCommand(ctx), session, unlockFunc))

	return r
}

func (r *Registry) promptUnlock() error {
	user := r.ctx.AuthService.GetUser()
	fmt.Print("🔒 Master password: ")
	passwordBytes, err := readline.Password("")
	if err != nil {
		return err
	}
	fmt.Println()

	password := string(passwordBytes)

	// Получаем соль из хранилища
	salt, err := r.ctx.DataService.GetSalt(user.Email)
	if err != nil {
		return fmt.Errorf("failed to get salt: %w", err)
	}

	return r.session.Unlock(password, salt)
}

// Register регистрирует команду
func (r *Registry) Register(cmd types.Command) {
	r.commands[cmd.Name()] = cmd
	for _, alias := range cmd.Aliases() {
		r.commands[alias] = cmd
	}
}

// Get возвращает команду по имени
func (r *Registry) Get(name string) types.Command {
	return r.commands[strings.ToLower(name)]
}

// GetAll возвращает все команды
func (r *Registry) GetAll() []types.Command {
	log.Printf("%v", r)
	var cmds []types.Command
	seen := make(map[string]bool)

	for _, cmd := range r.commands {
		if !seen[cmd.Name()] {
			seen[cmd.Name()] = true
			cmds = append(cmds, cmd)
		}
	}

	// Сортируем по имени
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name() < cmds[j].Name()
	})

	return cmds
}

// GetCompleter возвращает функцию автодополнения для readline
func (r *Registry) GetCompleter() func(string) []string {
	return func(line string) []string {
		var completions []string

		// Получаем текущее слово
		words := strings.Fields(line)
		current := ""
		if len(words) > 0 {
			current = words[len(words)-1]
		}

		// Если это первое слово - дополняем команды
		if len(words) <= 1 {
			for _, cmd := range r.GetAll() {
				if strings.HasPrefix(cmd.Name(), current) {
					completions = append(completions, cmd.Name())
				}
			}
			return completions
		}

		// Если это не первое слово - ищем подкоманды
		rootCmd := r.Get(words[0])
		if parent, ok := rootCmd.(types.ParentCommand); ok {
			for _, sub := range parent.SubCommands() {
				if strings.HasPrefix(sub.Name(), current) {
					completions = append(completions, sub.Name())
				}
			}
		}

		// Если есть специфичный комплитер для команды
		if rootCmd != nil && rootCmd.Completer() != nil {
			return rootCmd.Completer()(line)
		}

		return completions
	}
}
