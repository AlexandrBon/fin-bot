package app

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"strconv"
	"strings"
	"sync"
)

type UserState struct {
	state            string
	operationStateCh chan int16
	message          chan string
}

const (
	CANCELED = -1
	CONTINUE = 0

	EmptyState = ""
)

var (
	UserStates = make(map[int64]UserState)
	mu         sync.Mutex
)

func NewUserState(state string) UserState {
	return UserState{operationStateCh: make(chan int16), state: state, message: make(chan string)}
}

func ChangeUserState(chatID int64, state string) {
	mu.Lock()
	defer mu.Unlock()

	UserStates[chatID] = NewUserState(state)
}

func ClearUserState(chatID int64) {
	mu.Lock()
	defer mu.Unlock()

	value, ok := UserStates[chatID]
	if !ok {
		return
	}
	value.state = EmptyState
	UserStates[chatID] = value
}

func HandleUpdate(a App, upd tgbotapi.Update) {
	if upd.Message == nil {
		return
	}

	splitText := strings.Split(upd.Message.Text, " ")
	cmd := splitText[0]

	us, ok := UserStates[upd.Message.Chat.ID]
	if ok && us.state != EmptyState { // тут проработать несколько команд одновременной от 1 юзера
		if strings.HasPrefix(cmd, "/") && cmd != "/cancel" {
			a.SendMessage(upd.Message.Chat.ID, "You have not completed the previous command. \n"+
				"Enter /cancel to cancel it, or follow the instructions for using the command")
		} else {
			us.operationStateCh <- CONTINUE
			us.message <- upd.Message.Text
		}
	}

	switch cmd {
	case "/start":
		go a.Start(upd.Message.Chat.ID)
	case "/topup_balance":
		if len(splitText) != 2 {
			a.SendMessage(upd.Message.Chat.ID, "Usage: /topup_balance {amount of rubles}. "+
				"For example: \n/topup_balance +500\n/topup_balance 450")
			return
		}
		value, err := strconv.ParseInt(splitText[1], 10, 64)
		if err != nil || value <= 0 {
			a.SendMessage(upd.Message.Chat.ID, "The number of rubles can only be a positive number. Try again.")
			return
		}
		ChangeUserState(upd.Message.Chat.ID, cmd)
		go a.TopUpBalance(upd.Message.Chat.ID, value)
	case "/buy":
		n := len(splitText)
		value, err := strconv.ParseInt(splitText[n-1], 10, 64)
		if len(splitText) < 3 || err != nil {
			a.SendMessage(upd.Message.Chat.ID, "Usage: /buy {name} {price}. For example: /buy banana 50")
			return
		} else if value < 0 {
			a.SendMessage(upd.Message.Chat.ID, "Price can only be non-negative. Try again.")
			return
		}
		ChangeUserState(upd.Message.Chat.ID, cmd)
		go a.Buy(upd.Message.Chat.ID, strings.TrimSpace(strings.Join(splitText[1:n-1], " ")), value)
	case "/get_balance":
		go a.GetBalance(upd.Message.Chat.ID)
	case "/cancel":
		if us.state != "" {
			us.operationStateCh <- CANCELED
		}
	}
}
