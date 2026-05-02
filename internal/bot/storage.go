package bot

import (
	"log"
	"regexp"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserData struct {
	Username               string
	State                  State
	Plan                   string
	LastSupportInteraction *time.Time
}

var (
	users      = make(map[int64]*UserData)
	usersMu    sync.RWMutex
	emailRegex = regexp.MustCompile(`(?i)^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
)

var (
	lastInteract = make(map[int64]time.Time)
	ltMu         sync.RWMutex
)

// Возвращает объект данных пользователя.
// При работе с callback настоятельно рекомендуется передавать callback.From
func getUser(sender *tgbotapi.User) *UserData {
	if sender == nil {
		log.Printf("getUser: sender is nil")
		return nil
	}

	id := sender.ID

	if id == 0 {
		log.Printf("getUser: zero user id")
	}

	if users == nil {
		log.Printf("getUser: users map is nil")
	}

	if sender.UserName == "" {
		log.Printf("getUser: empty username user=%d", id)
	}

	usersMu.RLock()
	user := users[id]
	usersMu.RUnlock()

	if user != nil {
		return user
	}

	usersMu.Lock()
	defer usersMu.Unlock()

	if users[id] == nil {
		log.Printf("getUser: creating user=%d username=%s", id, sender.UserName)
		users[id] = &UserData{
			Username: sender.UserName,
			State:    StateIdle,
		}
	} else {
		log.Printf("getUser: user already created concurrently user=%d", id)
	}

	return users[id]
}
