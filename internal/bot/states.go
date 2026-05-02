package bot

type State string

const (
	StateIdle        State = "idle"
	StateChoosePlan  State = "choose_plan"
	StateWaitEmail   State = "wait_email"
	StateSupportChat State = "support_chat"
)
