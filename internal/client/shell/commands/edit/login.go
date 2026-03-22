// internal/shell/commands/edit/login.go
package edit

import (
	"fmt"
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/client/shell/commands/types"
	"github.com/Ko4etov/gophkeeper/internal/models"
)

type EditLoginCommand struct {
    types.BaseCommand
}

func NewEditLoginCommand(ctx *types.CommandContext) *EditLoginCommand {
    return &EditLoginCommand{
        BaseCommand: types.BaseCommand{
            NameText:    "login",
            AliasesText: []string{"lp", "password", "account"},
            HelpText:    "login <id> - Edit login/password entry",
            Context:     ctx,
        },
    }
}

func (c *EditLoginCommand) Execute(args []string) error {
    if err := c.BaseCommand.CheckLoggedIn(); err != nil {
        return err
    }
    
    if len(args) == 0 {
        c.BaseCommand.Printf("❌ Please specify login ID\n")
        c.BaseCommand.Printf("Usage: edit login <id>\n")
        c.BaseCommand.Printf("Example: edit login 123e4567-e89b-12d3-a456-426614174000\n")
        return nil
    }
    
    loginID := args[0]
    currentUser := c.BaseCommand.Context.AuthService.GetUser()
    
    // Получаем текущие данные логина
    entry, err := c.BaseCommand.Context.DataService.GetLogin(loginID, currentUser.Email)
    if err != nil {
        c.BaseCommand.Printf("❌ Failed to get login: %v\n", err)
        return err
    }
    
    // Извлекаем данные
    loginData, ok := entry.Data.(models.LoginPasswordData)
    if !ok {
        c.BaseCommand.Printf("❌ Invalid data type for login entry\n")
        return nil
    }

    currentLogin := loginData.Login
    currentPassword := loginData.Password
    currentURL := loginData.URL
    currentNotes := loginData.Notes
    
    c.BaseCommand.Printf("📝 Editing login: %s\n", entry.Meta.Name)
    c.BaseCommand.Println(strings.Repeat("─", 50))
    
    // Показываем текущие значения и запрашиваем новые
    c.BaseCommand.Printf("Press Enter to keep current value, or enter new value:\n\n")
    
    // Имя записи
    namePrompt := fmt.Sprintf("Name [%s]: ", entry.Meta.Name)
    name, err := c.BaseCommand.Prompt(namePrompt)
    if err != nil {
        return err
    }
    if name == "" {
        name = entry.Meta.Name
    }
    
    // Логин
    loginPrompt := fmt.Sprintf("Login [%s]: ", currentLogin)
    login, err := c.BaseCommand.Prompt(loginPrompt)
    if err != nil {
        return err
    }
    if login == "" {
        login = currentLogin
    }
    
    // Пароль (всегда запрашиваем новый, но можно пропустить)
    c.BaseCommand.Printf("Password (leave empty to keep current):\n")
    password, err := c.BaseCommand.PromptPassword("New password: ")
    if err != nil {
        return err
    }
    if password == "" {
        password = currentPassword
    } else {
        // Запрашиваем подтверждение пароля
        confirm, err := c.BaseCommand.PromptPassword("Confirm password: ")
        if err != nil {
            return err
        }
        if password != confirm {
            c.BaseCommand.Printf("❌ Passwords do not match\n")
            return nil
        }
    }
    
    // URL
    urlPrompt := fmt.Sprintf("URL [%s]: ", currentURL)
    url, err := c.BaseCommand.Prompt(urlPrompt)
    if err != nil {
        return err
    }
    if url == "" {
        url = currentURL
    }
    
    // Заметки
    notesPrompt := fmt.Sprintf("Notes [%s]: ", currentNotes)
    notes, err := c.BaseCommand.Prompt(notesPrompt)
    if err != nil {
        return err
    }
    if notes == "" {
        notes = currentNotes
    }
    
    // Теги
    currentTags := strings.Join(entry.Meta.Tags, ", ")
    tagsPrompt := fmt.Sprintf("Tags [%s]: ", currentTags)
    tagsStr, err := c.BaseCommand.Prompt(tagsPrompt)
    if err != nil {
        return err
    }
    
    var tags []string
    if tagsStr != "" {
        tags = strings.Split(tagsStr, ",")
        for i, tag := range tags {
            tags[i] = strings.TrimSpace(tag)
        }
    } else {
        tags = entry.Meta.Tags
    }
    
    // Показываем сводку изменений
    c.BaseCommand.Println(strings.Repeat("─", 50))
    c.BaseCommand.Printf("Summary of changes:\n")
    c.BaseCommand.Printf("  Name:  %s → %s\n", entry.Meta.Name, name)
    c.BaseCommand.Printf("  Login: %s → %s\n", currentLogin, login)
    if password != currentPassword {
        c.BaseCommand.Printf("  Password: [will be updated]\n")
    }
    if url != currentURL {
        c.BaseCommand.Printf("  URL:   %s → %s\n", currentURL, url)
    }
    c.BaseCommand.Println(strings.Repeat("─", 50))
    
    // Подтверждение
    confirm, err := c.BaseCommand.Confirm("Save changes?")
    if err != nil || !confirm {
        c.BaseCommand.Printf("❌ Edit cancelled\n")
        return nil
    }
    
    c.BaseCommand.Printf("\n💾 Saving changes... ")
    
    // Обновляем запись
    if err := c.BaseCommand.Context.DataService.UpdateLogin(
        loginID, name, login, password, url, notes, tags, currentUser.Email,
    ); err != nil {
        c.BaseCommand.Printf("❌ Failed: %v\n", err)
        return err
    }
    
    c.BaseCommand.Printf("✅ Login updated successfully!\n")
    return nil
}

// Completer возвращает возможные варианты для автодополнения
func (c *EditLoginCommand) Completer() func(string) []string {
    return func(line string) []string {
        // Здесь можно добавить автодополнение ID логинов
        // Например, получить список недавних логинов
        return []string{}
    }
}