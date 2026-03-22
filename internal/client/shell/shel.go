// internal/shell/shell.go
package shell

import (
	"fmt"
	"io"
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/config"
	"github.com/Ko4etov/gophkeeper/internal/client/crypto"
	grpcclient "github.com/Ko4etov/gophkeeper/internal/client/grpc_client"
	"github.com/Ko4etov/gophkeeper/internal/client/service/auth"
	"github.com/Ko4etov/gophkeeper/internal/client/service/data"
	"github.com/Ko4etov/gophkeeper/internal/client/service/sync"
	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands"
	"github.com/chzyer/readline"
)

type Shell struct {
	AuthService *auth.AuthService
	DataService *data.DataService
	config      *config.ClientConfig
	reader      *readline.Instance
	registry    *commands.Registry
	buildInfo   *grpcclient.BuildInfo
}

// New создает новую оболочку
func New(
	authService *auth.AuthService,
	cfg *config.ClientConfig,
	buildInfo *grpcclient.BuildInfo,
	dataService *data.DataService,
	syncManager *sync.SyncManager,
	session *crypto.Session,
) (*Shell, error) {

	s := &Shell{
		AuthService: authService,
		DataService: dataService,
		config:    cfg,
		buildInfo: buildInfo,
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            s.getPrompt(),
		HistoryFile:       cfg.HistoryFile,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		AutoComplete:      nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create readline: %w", err)
	}

	s.reader = rl

	s.registry = commands.NewRegistry(
		dataService,
		authService,
		syncManager,
		cfg,
		rl,
		buildInfo,
		session,
	)

	// Устанавливаем автодополнение
	rl.Config.AutoComplete = readline.NewPrefixCompleter(
		readline.PcItemDynamic(s.registry.GetCompleter()),
	)

	return s, nil
}

func (s *Shell) getPrompt() string {
	CurrentUser := s.AuthService.GetUser()

	if CurrentUser != nil {
		return fmt.Sprintf("%s@gophkeeper> ", CurrentUser.Email)
	}
	return "gophkeeper> "
}

func (s *Shell) Run() error {
	defer s.reader.Close()

	s.printWelcome()

	for {
		line, err := s.reader.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			}
			continue
		} else if err == io.EOF {
			break
		} else if err != nil {
			s.printf("Error reading input: %v\n", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := parseLine(line)
		if len(parts) == 0 {
			continue
		}

		cmdName := parts[0]
		args := parts[1:]

		cmd := s.registry.Get(cmdName)
		if cmd == nil {
			s.printf("Unknown command: %s (type 'help')\n", cmdName)
			continue
		}

		if err := cmd.Execute(args); err != nil {
			s.printf("Error: %v\n", err)
		}

		s.reader.SetPrompt(s.getPrompt())
	}

	s.printf("Goodbye!\n")
	return nil
}

func (s *Shell) printWelcome() {
	version := "dev"
	if s.buildInfo != nil {
		version = s.buildInfo.Version
	}

	s.printf("\n🔐 GophKeeper %s\n", version)
	s.printf("Type 'help' for commands, 'exit' or Ctrl+D to quit\n\n")
}

func (s *Shell) printf(format string, args ...interface{}) {
	fmt.Fprintf(s.reader.Stderr(), format, args...)
}

func parseLine(line string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false

	for _, r := range line {
		switch {
		case r == '"' || r == '\'':
			inQuotes = !inQuotes
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
