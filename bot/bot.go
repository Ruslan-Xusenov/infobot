package bot

import (
	"log"
	"sync"
	"time"

	"github.com/company/infobot/config"
	"gopkg.in/telebot.v3"
)

var (
	b         *telebot.Bot
	appConfig *config.Config
)

type AdminState struct {
	State string
	Data  map[string]interface{}
}

var (
	adminStates = make(map[int64]*AdminState)
	stateMu     sync.RWMutex
)

func SetAdminState(userID int64, state string, data map[string]interface{}) {
	stateMu.Lock()
	defer stateMu.Unlock()
	adminStates[userID] = &AdminState{State: state, Data: data}
}

func GetAdminState(userID int64) *AdminState {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return adminStates[userID]
}

func ClearAdminState(userID int64) {
	stateMu.Lock()
	defer stateMu.Unlock()
	delete(adminStates, userID)
}

func Start(cfg *config.Config) {
	appConfig = cfg

	pref := telebot.Settings{
		Token:  cfg.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	var err error
	b, err = telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	RegisterHandlers()

	log.Println("Bot is running...")
	b.Start()
}
