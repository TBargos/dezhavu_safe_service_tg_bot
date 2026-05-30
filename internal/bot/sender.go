package bot

import (
	"bytes"
	"dezhavu_tg_bot/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Отправляет заявку на подключение в канал Telegram. Выведено из использования.
func sendToTelegramChannel(bot *tgbotapi.BotAPI, userID int64, username string, plan, email string, groupID int64) error {
	text := fmt.Sprintf(
		"📋 <b>Новая заявка</b>\n\n"+
			"👤 User ID: <code>%d</code>\n"+
			"👤 Username: <code>%s</code>\n"+
			"💰 План: %s\n"+
			"📧 Email: <code>%s</code>",
		userID, username, plan, email,
	)

	msg := tgbotapi.NewMessage(groupID, text)
	msg.ParseMode = "HTML"

	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send to channel: %w", err)
	}

	return nil
}

func sendToServer(chatID int64, username string, plan, email string) error {

	serverURL := "http://your-server.com/api/zayavki"

	payload := map[string]interface{}{
		"chat_id":  chatID,
		"username": username,
		"plan":     plan,
		"email":    email,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, serverURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("server returned status %s", resp.Status)
	}

	return nil
}

func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64, username string) {
	// 1. Отправляем картинку
	sendLogo(bot, chatID)
	// 2. Отправляем текст с кнопками
	hello_text := fmt.Sprintf(mainMenuFText, username)

	msg := tgbotapi.NewMessage(chatID, hello_text)

	// Отключаем превью ссылки
	msg.DisableWebPagePreview = true

	// 2.1. Кнопки для получения доступа и продления
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💎 Получить доступ", BtnGetAccess),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Личный кабинет", BtnProfile),
			tgbotapi.NewInlineKeyboardButtonData("🆘 Помощь", BtnHelp),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Конфиденциальность", BtnPrivacy),
		),
	)

	if _, err := bot.Send(msg); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

func sendHelp(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 Начать диалог", BtnSupportDialog),
			tgbotapi.NewInlineKeyboardButtonData("📖 Инструкция", BtnShowTextManual),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👈 Главное меню", BtnBackMain),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

func sendPlanSelectionMenu(bot *tgbotapi.BotAPI, chatID int64) {
	text := `💎1 месяц —> 129₽
💵 3 месяца —> 319₽
💰1 год —> 1190₽

👉🏼Выгода при оплате
за 1 год <350₽` + "\u00A0👈🏼" + `


⬇️⬇️ Выбрать период: ⬇️⬇️`
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗓️1 месяц", BtnPlan1),
			tgbotapi.NewInlineKeyboardButtonData("😎3 месяца", BtnPlan3),
			tgbotapi.NewInlineKeyboardButtonData("🔥1 год", BtnPlan12),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👈 Главное меню", BtnBackMain),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

func sendPrivacy(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, privacyText)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

// Отправляем пользователю сообщение с ссылкой на оплату
func sendInvoice(bot *tgbotapi.BotAPI, userID int64, invoiceURL string) {
	textTemplate := `Рады приветствовать Вас в DezhavuVPN👾

📌 Ссылка на оплату действует ограниченное время.

После оплаты статус подписки в личном кабинете обновится автоматически.

Вам станут доступны:
• 🔗 персональная ссылка для подключения
• ⚡️ доступ к сервису

⬇️ Ссылка на оплату ⬇️

%s`

	text := fmt.Sprintf(textTemplate, invoiceURL)

	msg := tgbotapi.NewMessage(userID, text)
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤Личный кабинет", BtnProfile),
		),
	)

	if _, err := bot.Send(msg); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

// Отправляем пользователю его профиль с данными о подписке
func sendProfile(bot *tgbotapi.BotAPI, chatID int64, info *models.ProfileInfoResponse) {

	text := fmt.Sprintf(
		profileTextTemplate,
		info.Status,
		info.ExpireDate,
		info.RemainTime,
		info.ConnectLink,
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳 Оплатить подписку", BtnPaySubscription),
			tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить статус", BtnUpdateProfile),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📖 Инструкция", BtnShowTextManual),
			tgbotapi.NewInlineKeyboardButtonData("🆘 Помощь", BtnHelp),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👈 Главное меню", BtnBackMain),
		),
	)

	if _, err := bot.Send(msg); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

// Редактируем сообщение с профилем, чтобы обновить данные о подписке
func editProfile(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, info *models.ProfileInfoResponse) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	text := fmt.Sprintf(
		profileTextTemplate,
		info.Status,
		info.ExpireDate,
		info.RemainTime,
		info.ConnectLink,
	)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "HTML"
	edit.ReplyMarkup = cb.Message.ReplyMarkup // сохраняем кнопки

	_, err := bot.Send(edit)
	if err != nil {
		if strings.Contains(err.Error(), "message is not modified") {
			log.Printf(
				"profile update skipped: message is not modified",
			)
			return
		}

		log.Printf(
			"telegram edit profile error: %v",
			err,
		)
		return
	}

	log.Printf(
		"profile updated: chat_id=%d message_id=%d",
		chatID,
		messageID,
	)
}

// Уведомление о начала диалога с поддержкой
func sendStartSupport(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID,
		"🎯 Вы запустили диалог с поддержкой.\n\n"+
			"Опишите свою проблему в сообщении.\n"+
			"Чтобы вернуться в начало — нажмите кнопку в меню.")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

// Уведомление о том, что диалог просрочен
func sendSupportExpired(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Этот диалог уже закрыт ⏳\n" +
		"Если нужна помощь — откройте новый через меню 👇"
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👈 Главное меню", BtnBackMain),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

// Отправка сообщения в поддержку
// Сейчас реализовано как отправка формы в групповой чат заявок, что немного дублирует sendToTelegramChannel
// Выведено из использования
func sendToSupport(bot *tgbotapi.BotAPI, userID int64, username string, userMessage string, groupID int64) {
	text := fmt.Sprintf(
		"📋 <b>Новое сообщение</b>\n\n"+
			"👤 User ID: <code>%d</code>\n"+
			"👤 Username: <code>%s</code>\n"+
			"✉️ Сообщение:\n<blockquote>%s</blockquote>",
		userID, username, userMessage,
	)

	msg := tgbotapi.NewMessage(groupID, text)
	msg.ParseMode = "HTML"

	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("send to channel error: %v", err)
	}
}

// Отправляем пользователю текстовую инструкцию по подключению
func sendManual(bot *tgbotapi.BotAPI, userID int64) {
	msg := tgbotapi.NewMessage(userID, manualText)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Личный кабинет", BtnProfile),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎬 Видеоинструкция", BtnSendVideoManual),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👈 Главное меню", BtnBackMain),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

func sendVideoManual(bot *tgbotapi.BotAPI, chatID int64) {
	video := tgbotapi.NewVideo(chatID, tgbotapi.FilePath("assets/videos/manual.mp4"))
	_, err := bot.Send(video)
	if err != nil {
		log.Printf("send video: %v", err)
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Видеоинструкция сейчас недоступна 😔"))
	}
}

// Отправляет логотип отдельным сообщением
func sendLogo(bot *tgbotapi.BotAPI, chatID int64) {
	// Пытаемся взять FileID из кэша.
	fileID := GetFileID(Logo)

	var photo tgbotapi.PhotoConfig
	needUpdateCache := false

	if fileID != "" {
		photo = tgbotapi.NewPhoto(chatID, tgbotapi.FileID(fileID))
	} else {
		// Если FileID ещё неизвестен, отправляем локальный файл.
		photo = tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(GetPathAsset(Logo)))
		needUpdateCache = true
	}

	msg, err := bot.Send(photo)
	if err != nil {
		log.Printf("send photo: %v", err)
		return
	}

	// Если фото было отправлено из локального файла, сохраняем полученный от Telegram FileID
	if needUpdateCache && len(msg.Photo) > 0 {
		SetFileID(Logo, msg.Photo[len(msg.Photo)-1].FileID)
	}
}
