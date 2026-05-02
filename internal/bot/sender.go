package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func sendToTelegramChannel(bot *tgbotapi.BotAPI, userID int64, username string, plan, email string, channelID int64) error {
	text := fmt.Sprintf(
		"📋 <b>Новая заявка</b>\n\n"+
			"👤 User ID: <code>%d</code>\n"+
			"👤 Username: <code>%s</code>\n"+
			"💰 План: %s\n"+
			"📧 Email: <code>%s</code>",
		userID, username, plan, email,
	)

	msg := tgbotapi.NewMessage(channelID, text)
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
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath("assets/images/logo.jpg"))
	_, err := bot.Send(photo)
	if err != nil {
		log.Printf("send photo: %v", err)
	}

	// 2. Отправляем текст с кнопками
	hello_text := fmt.Sprintf(
		`Добро пожаловать в dezhavuVPN, %s!

📈 высокая скорость
🕵️ доступ ко всем сайтам
🛟 поддержка 24/7

👫 Пригласите друзей в наш сервис!

📌 Обязательно (!!) добавьте наш сайт http://dezhavu-rest.ru/ себе в закладки/избранное/ярлык на рабочий экран телефона.
Так Вы точно не потеряете свой vpn, чтобы не произошло.

⬇️⬇️ Получить доступ: ⬇️⬇️`,
		username)

	msg := tgbotapi.NewMessage(chatID, hello_text)

	// Отключаем превью ссылки
	msg.DisableWebPagePreview = true

	// 2.1. Кнопки для получения доступа и продления
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💸 Тарифы", BtnGetAccess),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🤔 Не работает ЛК", BtnNotWork),
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
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👈 Назад", BtnBackMain),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

func sendNotWorking(bot *tgbotapi.BotAPI, chatID int64) {
	helpText := `
Если вы не можете зайти в личный кабинет или он работает некорректно, попробуйте:

 👉если вы заходите с WiFi, выключите его (или наоборот, включите)

 👉попробуйте зайти в кабинет с включенным VPN

 👉попробуйте открыть ссылку на кабинет в другом браузере (Длинное нажатие на ссылку - "Открыть в...")

 👉обновите страницу (меню браузера - Обновить ⟳)`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🆘 Поддержка", BtnHelp),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👈 Назад", BtnBackMain),
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
			tgbotapi.NewInlineKeyboardButtonData("👈 Назад", BtnBackMain),
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
			tgbotapi.NewInlineKeyboardButtonData("👈 В главное меню", BtnBackMain),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

// Отправка сообщения в поддержку
// Сейчас реализовано как отправка формы в чат заявок, что немного дублирует sendToTelegramChannel
func sendToSupport(bot *tgbotapi.BotAPI, userID int64, username string, userMessage string, channelID int64) {
	text := fmt.Sprintf(
		"📋 <b>Новое сообщение</b>\n\n"+
			"👤 User ID: <code>%d</code>\n"+
			"👤 Username: <code>%s</code>\n"+
			"✉️ Сообщение:\n<blockquote>%s</blockquote>",
		userID, username, userMessage,
	)

	msg := tgbotapi.NewMessage(channelID, text)
	msg.ParseMode = "HTML"

	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("send to channel error: %v", err)
	}
}
