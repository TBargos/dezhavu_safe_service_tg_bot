package bot

import (
	"errors"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// -----------------------------------------------------------------------//
// ------------------------------UserData---------------------------------//
// -----------------------------------------------------------------------//
type Plan string

const (
	Plan1Month   Plan = "1 месяц"
	Plan3Months  Plan = "3 месяца"
	Plan12Months Plan = "1 год"
)

func (p Plan) Int() int {
	switch p {
	case Plan1Month:
		return 1
	case Plan3Months:
		return 3
	case Plan12Months:
		return 12
	}
	return 1
}

type UserData struct {
	Username               string
	State                  State
	Plan                   Plan
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

// -----------------------------------------------------------------------//
// -----------------------------------------------------------------------//
// -----------------------------------------------------------------------//

type Asset string

const (
	Logo        Asset = "logo"
	VideoManual Asset = "video_manual"
)

var assetPaths = map[Asset]string{
	Logo:        "assets/images/logo.jpg",
	VideoManual: "assets/videos/manual.mp4",
}

type Assets struct {
	Files map[Asset]*string
	mu    sync.RWMutex
}

var assets = func() Assets {
	files := make(map[Asset]*string)

	for asset := range assetPaths {
		files[asset] = nil
	}

	return Assets{
		Files: files,
	}
}()

// Возвращает FileID ассета или пустую строку, если он ещё не закэширован.
func GetFileID(asset Asset) string {
	assets.mu.RLock()
	fileID := assets.Files[asset]
	assets.mu.RUnlock()

	if fileID == nil {
		return ""
	}

	return *fileID
}

// Сохраняет FileID ассета в кэш.
func SetFileID(asset Asset, fileID string) {
	assets.mu.Lock()
	defer assets.mu.Unlock()

	assets.Files[asset] = &fileID
}

// Возвращает путь к ассету.
func GetPathAsset(asset Asset) string {
	return assetPaths[asset]
}

// Проверяет наличие всех зарегистрированных ассетов на диске.
func CheckAssets() error {
	var err error

	for _, path := range assetPaths {
		if _, subErr := os.Stat(path); subErr != nil {
			err = errors.Join(err, subErr)
		}
	}

	return err
}
