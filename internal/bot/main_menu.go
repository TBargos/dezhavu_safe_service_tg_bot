package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Настраивает команды в главном меню бота (которые появляются при вводе "/")
func SetMyCommands(bot *tgbotapi.BotAPI) (*tgbotapi.APIResponse, error) {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Главное меню"},
		{Command: "help", Description: "Помощь"},
		{Command: "privacy", Description: "Политика конфиденциальности"},
	}
	return bot.Request(tgbotapi.NewSetMyCommands(commands...))
}
