// internal/shell/commands/delete/delete.go
package delete

import (
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
)

// DeleteCommand родительская команда для всех операций удаления
type DeleteCommand struct {
    types.BaseCommand
    subCommands map[string]types.Command
}

// NewDeleteCommand создает новую команду delete
func NewDeleteCommand(ctx *types.CommandContext) *DeleteCommand {
    cmd := &DeleteCommand{
        BaseCommand: types.NewSecureCommand(
            "delete",
            []string{"del", "rm", "remove"},
            "delete <command> - Delete existing data (card, login, text, binary)",
            ctx,
        ),
        subCommands: make(map[string]types.Command),
    }
    
    // Регистрируем все подкоманды удаления
    cmd.registerSubCommand(NewDeleteCardCommand(ctx))
    cmd.registerSubCommand(NewDeleteLoginCommand(ctx))
    cmd.registerSubCommand(NewDeleteTextCommand(ctx))
    cmd.registerSubCommand(NewDeleteBinaryCommand(ctx))
    
    return cmd
}

// registerSubCommand регистрирует подкоманду и ее алиасы
func (c *DeleteCommand) registerSubCommand(cmd types.Command) {
    c.subCommands[cmd.Name()] = cmd
    for _, alias := range cmd.Aliases() {
        c.subCommands[alias] = cmd
    }
}

// Execute выполняет команду delete
func (c *DeleteCommand) Execute(args []string) error {
    // Проверка авторизации для всех операций удаления
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    // Если нет аргументов - показываем справку
    if len(args) == 0 {
        c.printHelp()
        return nil
    }
    
    // Ищем подкоманду
    subCmd := c.subCommands[strings.ToLower(args[0])]
    if subCmd == nil {
        c.Printf("❌ Unknown delete command: %s\n", args[0])
        c.Printf("Run 'delete' without arguments to see available commands\n")
        return nil
    }
    
    // Выполняем подкоманду
    return subCmd.Execute(args[1:])
}

// printHelp выводит список доступных подкоманд
func (c *DeleteCommand) printHelp() {
    c.Printf("📝 Available delete commands:\n\n")
    
    // Группируем по типу для лучшей читаемости
    c.Printf("  🔐 Logins & Passwords:\n")
    c.Printf("    delete login <id>     - Delete login/password entry\n")
    c.Printf("    delete pass <id>      - Alias for login\n\n")
    
    c.Printf("  💳 Bank Cards:\n")
    c.Printf("    delete card <id>      - Delete bank card\n")
    c.Printf("    delete creditcard <id> - Alias for card\n\n")
    
    c.Printf("  📝 Text Data:\n")
    c.Printf("    delete text <id>      - Delete text note\n")
    c.Printf("    delete note <id>      - Alias for text\n\n")
    
    c.Printf("  📎 Binary Files:\n")
    c.Printf("    delete binary <id>    - Delete binary file\n")
    c.Printf("    delete file <id>      - Alias for binary\n\n")
    
    c.Printf("Examples:\n")
    c.Printf("  delete card 123e4567-e89b-12d3-a456-426614174000\n")
    c.Printf("  delete login 987fcdeb-51a2-43f7-9a8b-123456789abc\n")
    c.Printf("  delete text b1d4e5f6-7890-4a5b-8c6d-7e8f9a0b1c2d\n")
}

// SubCommands возвращает список всех подкоманд
func (c *DeleteCommand) SubCommands() []types.Command {
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

// GetSubCommand возвращает подкоманду по имени
func (c *DeleteCommand) GetSubCommand(name string) types.Command {
    return c.subCommands[strings.ToLower(name)]
}

// Completer возвращает функцию автодополнения
func (c *DeleteCommand) Completer() func(string) []string {
    return func(line string) []string {
        var completions []string
        
        // Получаем текущее слово
        words := strings.Fields(line)
        if len(words) <= 1 {
            // Дополняем имена подкоманд
            for name := range c.subCommands {
                completions = append(completions, name)
            }
        } else {
            // Для подкоманд используем их комплитеры
            subCmd := c.subCommands[strings.ToLower(words[1])]
            if subCmd != nil && subCmd.Completer() != nil {
                return subCmd.Completer()(line)
            }
        }
        
        return completions
    }
}