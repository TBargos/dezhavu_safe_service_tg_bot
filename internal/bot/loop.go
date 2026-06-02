package bot

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func InfiniteLoop(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, groupID int64) {
	for update := range updates {
		msg := update.Message
		switch {
		// Обработка нажатий на inline-кнопки
		case update.CallbackQuery != nil:
			callback := update.CallbackQuery
			log.Printf("callback received: from=%d data=%s", callback.From.ID, callback.Data)
			processCallback(bot, callback)

		// Если пришло сообщение
		case msg != nil:
			processMessage(bot, msg, groupID)

		// Если не коллбэк, не сообщение - не обрабатываем, пропускаем дальше
		default:
			log.Print("skip update (no message, no callback)")
		}
	}
}

// Обработка callback (приходят при нажатии на inline-кнопки)
func processCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	data := callback.Data
	user_data := getUser(callback.From)

	keepKeyboard := false // Флаг, который показывает, нужно ли после обработки коллбэка убрать клавиатуру.

	// Сначала проверяем на кнопки вне стэйтов
	isNonState := true // Флаг, который показывает, что коллбек обработан в этом блоке и не нужно смотреть на стэйт
	switch data {
	case BtnProfile:
		handleProfile(bot, userID, user_data)
	case BtnSupportDialog:
		handleSupportDialog(bot, userID, user_data)
	case BtnHelp:
		handleHelp(bot, userID, user_data)
	case BtnShowTextManual:
		handleShowTextManual(bot, userID, user_data)
		keepKeyboard = true // Оставляем клавиатуру, так как не страшно, если тут пользователь будет прожимать кнопки
	case BtnSendVideoManual:
		handleSendVideoManual(bot, userID, user_data)
		keepKeyboard = true // Оставляем клавиатуру, так как не страшно, если тут пользователь будет прожимать кнопки
	case BtnPrivacy:
		handlePrivacy(bot, userID, user_data)
	default:
		isNonState = false
	}

	// Если это не внеочередное событие, то смотрим на стэйт и выбираем сценарий обработки коллбэка
	if !isNonState {
		switch user_data.State {
		case StateIdle:
			if data == BtnUpdateProfile {
				handleProfileUpdate(bot, userID, callback)
				keepKeyboard = true // Оставляем клавиатуру, так как будет обновление сообщения, а не отправка нового
			} else {
				handleMainMenu(bot, callback, user_data)
			}

		case StateChoosePlan:
			handlePlan(bot, callback, user_data)
		}
	}

	bot.Request(tgbotapi.NewCallback(callback.ID, "")) // Убирает "часики" на кнопке
	if !keepKeyboard {
		removeKeyboard(bot, callback) // Убирает клавиатуру
	}
}

// Обработка команд
func processCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := msg.Text
	chatID := msg.Chat.ID
	username := msg.From.UserName

	log.Printf("command received: chat_id=%d command=%s", chatID, text)

	switch text {
	case "/start":
		sendMainMenu(bot, chatID, username)

	case "/help":
		sendHelp(bot, chatID)

	case "/privacy":
		sendPrivacy(bot, chatID)

	default:
		log.Printf("using unfamiliar command=%q", text)
	}
}

func processMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, groupID int64) {
	log.Printf("message received: chat_id=%d", msg.Chat.ID)
	user := getUser(msg.From)

	// Сначала проверяем на сценарии вне стэйтов
	switch {

	// Пришёл reply на сообщение бота
	case msg.ReplyToMessage != nil:
		log.Printf(
			"reply received: msg.Chat.ID=%d msg.From.ID=%d msg.ReplyToMessage.MessageID=%d msg.ReplyToMessage.From.ID=%d msg.ReplyToMessage.Chat.ID=%d",
			msg.Chat.ID,
			msg.From.ID,
			msg.ReplyToMessage.MessageID,
			msg.ReplyToMessage.From.ID,
			msg.ReplyToMessage.Chat.ID,
		)
		// Если это reply на сообщение бота в групповом чате, то обрабатываем как ответ админа
		if msg.ReplyToMessage.Chat.ID == groupID {
			log.Printf(
				"admin reply detected: chat_id=%d reply_to_msg_id=%d",
				msg.Chat.ID,
				msg.ReplyToMessage.MessageID,
			)
			handleAdminReply(bot, msg, groupID)
			return
		}
	// Команда от пользователя (это сообщение, начинающееся с "/")
	case msg.IsCommand():
		user.State = StateIdle
		processCommand(bot, msg)
		return
	}

	// Если это не внеочередное событие, то смотрим на стэйт и выбираем сценарий обработки сообщения
	switch user.State {
	case StateWaitEmail:
		handleEmail(bot, msg, user)
		return
	case StateSupportChat:
		log.Printf(
			"support chat message: chat_id=%d text=%q",
			msg.Chat.ID,
			msg.Text,
		)
		handleUserAppeal(bot, msg, user, groupID)
		return
	}

	// Если не подходит ни один из сценариев, просто отправляем юзеру напоминание использовать кнопки меню
	_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Используйте кнопки меню 👇"))
	if err != nil {
		log.Printf("telegram send error: %v", err)
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
