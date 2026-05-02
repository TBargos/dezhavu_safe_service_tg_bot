package bot

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func InfiniteLoop(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, channelID int64) {
	for update := range updates {
		log.Printf("incoming update: %+v", update)

		// Обработка нажатий на inline-кнопки
		if update.CallbackQuery != nil {
			callback := update.CallbackQuery
			log.Printf("callback received: from=%s data=%s", callback.From.UserName, callback.Data)
			processCallback(bot, callback)
			continue
		}

		msg := update.Message
		// Если не коллбэк, но и не сообщение - не обрабатываем, пропускаем дальше
		if msg == nil {
			log.Printf("skip update (no message, no callback): %+v", update)
			continue
		}

		log.Printf("message received: chat_id=%d user=%s", msg.Chat.ID, msg.From.UserName)
		// На этом моменте
		user := getUser(msg.From)

		// Если админ в чате заявок ответил на пришедшую форму, то обрабатываем его ответ
		if msg.ReplyToMessage != nil {
			log.Printf(
				"admin reply detected: chat_id=%d reply_to_msg_id=%d",
				msg.Chat.ID,
				msg.ReplyToMessage.MessageID,
			)
			handleAdminReply(bot, msg, channelID)
			continue
		}

		// Если это сообщение в открытом диалоге поддержки
		if user.State == StateSupportChat && !strings.HasPrefix(msg.Text, "/") {
			log.Printf(
				"support chat message: user=%s text=%q",
				msg.From.UserName,
				msg.Text,
			)
			handleUserAppeal(bot, update.Message, user, channelID)
			continue
		}

		// Если же была не inline-кнопка, а кнопка из меню
		text := msg.Text
		chatID := msg.Chat.ID
		username := msg.From.UserName

		if strings.HasPrefix(text, "/") {
			user.State = StateIdle
			switch text {
			case "/start":
				sendMainMenu(bot, chatID, username)
				continue

			case "/help":
				sendHelp(bot, chatID)
				continue

			case "/privacy":
				sendPrivacy(bot, chatID)
				continue

			default:
				log.Printf("using unfamiliar command=%q", text)
				continue
			}
		}

		// Если событие (update) ничего из вышеперечисленного, то обрабатываем состояние пользователя
		switch user.State {
		case StateWaitEmail:
			handleEmail(bot, msg, user, channelID)

		default:
			if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Используйте кнопки меню 👇")); err != nil {
				log.Printf("telegram send error: %v", err)
			}
		}
	}
}

// Обработка callback (приходят при нажатии на inline-кнопки)
func processCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	bot.Request(tgbotapi.NewCallback(callback.ID, "")) // Убирает "часики" на кнопке
	removeKeyboard(bot, callback)                      // Убирает клавиатуру

	chatID := callback.From.ID
	data := callback.Data
	user_data := getUser(callback.From)

	// Обработка случаев, независящих от состояния (стэйта) пользователя
	switch data {
	case BtnSupportDialog:
		handleSupportDialog(bot, chatID, user_data)
		return
	}

	// Перенаправление контекста в зависимости от выставленного стэйта (состояния) пользователя
	switch user_data.State {
	case StateIdle:
		handleMainMenu(bot, callback, user_data)

	case StateChoosePlan:
		handlePlan(bot, callback, user_data)

	case StateWaitEmail:
		// ничего не делаем — ждём текст
	}
}

// В обработке callback убирает inline-клавиатуру
func removeKeyboard(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) {
	edit := tgbotapi.NewEditMessageReplyMarkup(
		cb.Message.Chat.ID,
		cb.Message.MessageID,
		tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
		},
	)

	_, err := bot.Send(edit)
	if err != nil && !strings.Contains(err.Error(), "message is not modified") {
		fmt.Printf("failed to remove inline keyboard: %v\n", err)
	}
}
