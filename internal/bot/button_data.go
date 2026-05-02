package bot

// Здесь хранятся константы для кнопок, чтобы не дублировать текст и callback data в коде
// Пока не внедрено

type ButtonKey string

const (
	// Кнопки главного меню
	BtnGetAccess = "get_access"
	BtnNotWork   = "not_working"
	BtnHelp      = "help"
	BtnPrivacy   = "privacy"
	BtnBackMain  = "back_main"

	// Кнопки выбора плана
	BtnPlan1  = "plan_1"
	BtnPlan3  = "plan_3"
	BtnPlan12 = "plan_12"

	// Кнопки поддержки
	BtnSupportDialog = "support_dialog"
)
