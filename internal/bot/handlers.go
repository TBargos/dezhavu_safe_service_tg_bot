package bot

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleMainMenu(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, user *UserData) {
	chatID := callback.From.ID
	username := callback.From.UserName
	data := callback.Data
	switch data {

	case BtnGetAccess:
		user.State = StateChoosePlan
		sendPlanSelectionMenu(bot, chatID)

	case BtnNotWork:
		sendNotWorking(bot, chatID)

	case BtnHelp:
		sendHelp(bot, chatID)

	case BtnPrivacy:
		sendPrivacy(bot, chatID)

	case BtnBackMain:
		user.State = StateIdle
		sendMainMenu(bot, chatID, username)
	}
}

func handlePlan(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, user *UserData) {
	chatID := callback.From.ID
	username := callback.From.UserName

	data := callback.Data
	switch data {

	case BtnBackMain:
		user.State = StateIdle
		sendMainMenu(bot, chatID, username)
		return

	case BtnPlan1:
		user.Plan = "1 месяц"

	case BtnPlan3:
		user.Plan = "3 месяца"

	case BtnPlan12:
		user.Plan = "1 год"

	default:
		return
	}

	user.State = StateWaitEmail
	if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Введите ваш email:")); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

func handleEmail(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, user *UserData, channelID int64) {
	chatID := msg.Chat.ID
	text := msg.Text
	username := msg.From.UserName

	if !isValidEmail(strings.TrimSpace(text)) {
		if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Некорректный email, попробуйте снова")); err != nil {
			log.Printf("telegram send error: %v", err)
		}
		return
	}

	err := sendToTelegramChannel(bot, chatID, username, user.Plan, text, channelID)
	if err != nil {
		log.Println(err)
		if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Ошибка отправки 😢")); err != nil {
			log.Printf("telegram send error: %v", err)
		}
		return
	}

	user.State = StateIdle

	if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Заявка отправлена администратору.\nОтвет придёт в течение 24 часов.")); err != nil {
		log.Printf("telegram send error: %v", err)
	}

	sendMainMenu(bot, chatID, username)
}

func handleAdminReply(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, channelID int64) {
	if msg.Chat.ID != channelID {
		// Если пришедшее сообщение не из канала - ничего не делаем
		log.Println("[ADMIN REPLY] message is not from the channel")
		return
	}

	if msg.ReplyToMessage == nil {
		log.Println("[ADMIN_REPLY] no reply")
		return
	}

	// Достать текст сообщения, на которое ответил админ
	// Логика такая, что админ отвечал на форму, в которой упоминался user_id
	original := msg.ReplyToMessage.Text
	log.Printf("[ADMIN_REPLY] original=%q", original)

	// Который тут достаётся регуляркой
	re := regexp.MustCompile(`User ID:\s*(\d+)`)
	matches := re.FindStringSubmatch(original)

	if len(matches) < 2 {
		log.Println("[ADMIN_REPLY] user id not found")
		return
	}

	// Конвертация строки в число (из сообщения всегда достаётся текст)
	// Но для отправки user_id должен быть в числовом типе
	userID, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		log.Println("[ADMIN_REPLY] parse error:", err)
		return
	}

	log.Printf("[ADMIN_REPLY] send to %d", userID)

	// Последний шаг - пересылка текста от админа так, словно написал бот
	reply := tgbotapi.NewMessage(userID, msg.Text)
	if _, err := bot.Send(reply); err != nil {
		log.Println("[ADMIN_REPLY] send error:", err)
	}
}

// Старт диалога с поддержкой
func handleSupportDialog(bot *tgbotapi.BotAPI, chatID int64, user *UserData) {
	user.State = StateSupportChat
	sendStartSupport(bot, chatID)
}

// Обработка сообщения пользователя в поддержку
func handleUserAppeal(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, user *UserData, channelID int64) {
	timeout := time.Hour

	if user.LastSupportInteraction != nil && time.Since(*user.LastSupportInteraction) > timeout {
		user.State = StateIdle
		sendSupportExpired(bot, msg.Chat.ID)
		return
	}

	now := time.Now()
	userID := msg.From.ID
	username := user.Username
	userText := msg.Text
	user.LastSupportInteraction = &now
	sendToSupport(bot, userID, username, userText, channelID)
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
