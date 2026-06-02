package bot

import (
	"dezhavu_tg_bot/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Обработчик нажатий на кнопки главного меню
func handleMainMenu(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, user *UserData) {
	username := callback.From.UserName
	userID := callback.From.ID
	data := callback.Data
	switch data {

	case BtnGetAccess:
		user.State = StateChoosePlan
		sendPlanSelectionMenu(bot, userID)

	case BtnProfile:
		user.State = StateIdle
		handleProfile(bot, userID, user)

	case BtnHelp:
		sendHelp(bot, userID)

	case BtnPrivacy:
		sendPrivacy(bot, userID)

	case BtnBackMain:
		user.State = StateIdle
		sendMainMenu(bot, userID, username)
	}
}

// Обработка вызова личного кабинета
func handleProfile(bot *tgbotapi.BotAPI, userID int64, user *UserData) {
	user.State = StateIdle
	info := prepareProfileInfo(userID)
	sendProfile(bot, userID, info)
}

// handleUpdateProfile(bot, chatID, callback.From.ID, callback.Message.MessageID)
func handleProfileUpdate(bot *tgbotapi.BotAPI, userID int64, callback *tgbotapi.CallbackQuery) {
	info := prepareProfileInfo(userID)
	editProfile(bot, callback, info)
}

// Обработчик нажатий на кнопки выбора плана
func handlePlan(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, user *UserData) {
	userID := callback.From.ID
	username := callback.From.UserName

	data := callback.Data
	switch data {

	case BtnBackMain:
		user.State = StateIdle
		sendMainMenu(bot, userID, username)

	case BtnPlan1:
		user.Plan = Plan1Month
	case BtnPlan3:
		user.Plan = Plan3Months
	case BtnPlan12:
		user.Plan = Plan12Months
	}

	user.State = StateWaitEmail
	if _, err := bot.Send(tgbotapi.NewMessage(userID, "Введите ваш email:")); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

// Обработчик ввода email пользователем
func handleEmail(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, user *UserData) {
	userID := msg.Chat.ID // В данном случае chatID и userID совпадают, так как сообщение от пользователя
	email := msg.Text

	if !isValidEmail(strings.TrimSpace(email)) {
		if _, err := bot.Send(tgbotapi.NewMessage(userID, "Некорректный email, попробуйте снова")); err != nil {
			log.Printf("telegram send error: %v", err)
		}
		return
	}
	// Если email уже валидный, в любом случае меняем стэйт на дефолтный, чтобы не обрабатывать следующие сообщения как email
	user.State = StateIdle

	// Достаём количество месяцев для счёта
	months := user.Plan.Int()

	invoiceURL, err := requestInvoice(userID, months, email)
	if err != nil {
		log.Printf("invoice request error: %v", err)
		if _, err := bot.Send(tgbotapi.NewMessage(userID, "Ошибка создания счёта 😢")); err != nil {
			log.Printf("telegram send error: %v", err)
		}
		return
	}

	sendInvoice(bot, userID, invoiceURL)
}

// Обработчик ответов админа в канале поддержки
func handleAdminReply(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, groupID int64) {
	if msg.Chat.ID != groupID {
		// Если пришедшее сообщение не из группы - ничего не делаем
		log.Println("[ADMIN REPLY] message is not from the group")
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
func handleSupportDialog(bot *tgbotapi.BotAPI, userID int64, user *UserData) {
	user.State = StateSupportChat
	now := time.Now()
	user.LastSupportInteraction = &now
	sendStartSupport(bot, userID)
}

// Обработка сообщения пользователя в поддержку
func handleUserAppeal(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, user *UserData, groupID int64) {
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
	sendToSupport(bot, userID, username, userText, groupID)
}

// Создаёт счёт на оплату подписки через backend API.
// При успехе возвращает ссылку на страницу оплаты.
func requestInvoice(userID int64, months int, email string) (string, error) {
	emailJSON, err := json.Marshal(email)
	if err != nil {
		return "", fmt.Errorf("marshal email: %w", err)
	}

	reqBody := fmt.Sprintf(`{
		"external_id": %d,
		"provider": "telegram",
		"duration_month": %d,
		"customer_email": %s,
		"return_url": null
	}`,
		userID, months, emailJSON,
	)

	resp, err := http.Post(
		backendBaseUrl+"/api/v1/payments",
		"application/json",
		strings.NewReader(reqBody),
	)
	if err != nil {
		return "", fmt.Errorf("invoice request error: %w", err)
	}
	defer resp.Body.Close()

	var result models.BackendResponse[struct {
		URL string `json:"url"`
	}]

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("invoice decode error: %w", err)
	}

	if !result.Success || result.Obj == nil {
		return "", fmt.Errorf("invoice request failed: %s, obj is null=%t", result.Msg, result.Obj == nil)
	}

	return result.Obj.URL, nil
}

// Спрашивает у сервиса данные о подписке и возвращает их в виде структуры для удобного использования
func prepareProfileInfo(userID int64) *models.ProfileInfoResponse {
	// Переменные для заполнения шаблона. Изначально - заглушки, если запрос не удался
	info := models.ProfileInfoResponse{
		Status:      "Не удалось загрузить данные",
		ExpireDate:  "Неизвестно",
		RemainTime:  "",
		ConnectLink: "Недоступна",
	}

	// Сборка запроса и отправка на сервер для получения данных о подписке
	reqBody := fmt.Sprintf(
		`{"provider":"telegram","external_id":%d}`,
		userID,
	)

	resp, err := http.Post(
		backendBaseUrl+"/api/v1/subscription-info",
		"application/json",
		strings.NewReader(reqBody),
	)
	if err == nil {
		defer resp.Body.Close()

		// Анонимная структура для чтения тела ответа. Определяем Obj и подставляем в
		var result models.BackendResponse[struct {
			SubscriptionID string     `json:"subscription_id"`
			Status         string     `json:"status"`
			ExpiredAt      *time.Time `json:"expired_at"`
		}]

		err := json.NewDecoder(resp.Body).Decode(&result)
		// if построены на отлов ошибок или неуспеха. Если же ни один не срабатывает - всё ок
		if err != nil {
			// Ошибка декодирования может возникать, если обратились не туда, неправильно или контракт изменился
			log.Printf(
				"subscription decode error: %v",
				err,
			)
		} else if resp.StatusCode == http.StatusNotFound {
			// Если ответ прочитан и там 404, значит у пользователя нет подписки.
			info.Status = "Подписка отсутствует"
			info.ExpireDate = "—"
			info.RemainTime = "—"
			info.ConnectLink = "Оформите подписку, чтобы получить индивидуальную ссылку"

			log.Printf(
				"subscription not found for user %d",
				userID,
			)

		} else if !result.Success || result.Obj == nil {
			// Если ответ прочитан, ответ не 200 и не 404, подразумевается, что из msg можно вытащить причину ошибки
			log.Printf(
				"subscription api error: %s",
				result.Msg,
			)
		} else {

			// Определяем статус подписки
			info.Status = helperFormatSubscriptionStatus(result.Obj.Status)

			// Если есть дата окончания, форматируем её и считаем оставшееся время
			if result.Obj.ExpiredAt != nil {
				infiniteTime := time.Date(9999, 12, 31, 23, 59, 0, 0, time.UTC)
				if !result.Obj.ExpiredAt.Before(infiniteTime) {
					// Если дата окончания очень далеко в будущем, считаем подписку бессрочной
					info.ExpireDate = "Бесконечно"
					info.RemainTime = "∞"
				} else {
					// Приведение к московскому времени (UTC+3). Telegram API не предоставляет часовой пояс пользователя
					localExpireTime := result.Obj.ExpiredAt.In(
						time.FixedZone("MSK", 3*60*60),
					)
					info.ExpireDate = helperformatRussianDate(localExpireTime)
					info.RemainTime = helperFormatRussianRemainingTime(localExpireTime)
				}

				// Формируем ссылку для подключения
				info.ConnectLink = helperBuildSubscriptionLink(result.Obj.SubscriptionID)
			} else {
				// Подразумевается, что если даты окончания нет, то подписка неактивна или не оплачена
				info.ExpireDate = "—"
				info.RemainTime = "—"
				info.ConnectLink = "Оплатите подписку, чтобы получить индивидуальную ссылку"
			}
		}
	} else {
		log.Printf("subscription request error: %v", err)
	}
	return &info
}

// Обработчик нажатия на кнопку помощи
func handleHelp(bot *tgbotapi.BotAPI, userID int64, user *UserData) {
	user.State = StateIdle
	sendHelp(bot, userID)
}

// Обработчик нажатия на кнопку политики конфиденциальности
func handlePrivacy(bot *tgbotapi.BotAPI, userID int64, user *UserData) {
	user.State = StateIdle
	sendPrivacy(bot, userID)
}

// Обработчик нажатия на кнопку инструкции
func handleShowTextManual(bot *tgbotapi.BotAPI, userID int64, user *UserData) {
	user.State = StateIdle
	sendManual(bot, userID)
}

func handleSendVideoManual(bot *tgbotapi.BotAPI, userID int64, user_data *UserData) {
	user_data.State = StateIdle
	sendVideoManual(bot, userID)
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

//--------------------------------------------------//
//-----------------helper functions-----------------//
//--------------------------------------------------//

// Форматирование даты в виде "25 июня 2026, 23:59 МСК". t нужно передать с часовым поясом Москвы
func helperformatRussianDate(t time.Time) string {
	months := russianMonths[:] // Осторожно, через срез можно случайно изменить глобальный список месяцев

	return fmt.Sprintf(
		"%d %s %d, %02d:%02d МСК",
		t.Day(),
		months[t.Month()-1],
		t.Year(),
		t.Hour(),
		t.Minute(),
	)
}

// t нужно передать с часовым поясом Москвы
func helperFormatRussianRemainingTime(t time.Time) string {

	d := time.Until(t)

	if d <= 0 {
		return "Подписка истекла"
	}

	days := int(d.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf(
			"Осталось %d дней",
			days,
		)
	}

	hours := int(d.Hours())
	if hours > 0 {
		return fmt.Sprintf(
			"Осталось %d часов",
			hours,
		)
	}

	minutes := int(d.Minutes())

	return fmt.Sprintf(
		"Осталось %d минут",
		minutes,
	)
}

// Форматирование статуса подписки в человекочитаемый вид
func helperFormatSubscriptionStatus(status string) string {
	switch status {
	case "pending":
		return "Ожидает оплаты"
	case "active":
		return "Активна"
	case "expired":
		return "Просрочена"
	case "canceled":
		return "Приостановлена"
	default:
		return "Неизвестно"
	}
}

// Формирование ссылки подключения для VPN-клиента
func helperBuildSubscriptionLink(subscriptionID string) string {
	return backendBaseUrl + "/subs/" + subscriptionID
}
