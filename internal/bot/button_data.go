package bot

// Здесь хранятся константы для кнопок, чтобы не дублировать текст и callback data в коде
// Пока не внедрено

type ButtonKey string

const (
	// Кнопки главного меню
	BtnGetAccess = "get_access"
	BtnProfile   = "profile" // Используется не только в главном меню
	BtnHelp      = "help"    // Используется не только в главном меню
	BtnPrivacy   = "privacy"
	BtnBackMain  = "back_main"

	// Кнопки личного кабинета
	BtnPaySubscription = BtnGetAccess // Тот же сценарий, что и у кнопки получения доступа. Существует, т.к. визуально отличается
	BtnUpdateProfile   = "update_profile"
	BtnShowTextManual  = "show_text_manual"

	// Кнопки выбора плана
	BtnPlan1  = "plan_1"
	BtnPlan3  = "plan_3"
	BtnPlan12 = "plan_12"

	// Кнопки поддержки
	BtnSupportDialog = "support_dialog"

	// Кнопки инструкции
	BtnSendVideoManual = "video_manual" // TODO: не реализовано
)
