package main

import (
	"log"
	"os"
	"strconv"

	bot_stuff "dezhavu_tg_bot/internal/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Получаем токен бота из переменных окружения
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN environment variable is required")
	}

	// Получаем канал для заявок и поддержки
	channelIDStr := os.Getenv("CHANNEL_ID")
	if channelIDStr == "" {
		log.Fatal("CHANNEL_ID не задан")
	}
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		log.Fatal("невалидный CHANNEL_ID: %w", err)
	}

	// Инициализация бота с помощью токена
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	// Настройка команд в главном меню
	_, err = bot_stuff.SetMyCommands(bot)
	if err != nil {
		log.Printf("failed to set bot commands: %v", err)
	}

	// Настройка канала для получения обновлений от Telegram
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Получаем канал обновлений от Telegram
	updates := bot.GetUpdatesChan(u)

	// Запускаем бесконечный цикл обработки обновлений
	bot_stuff.InfiniteLoop(updates, bot, channelID)
}
