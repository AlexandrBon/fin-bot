package app

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"log"
	"strconv"
	"strings"
)

const (
	SuccessMessage  = "Operation completed successfully!"
	FailedMessage   = "Something went wrong, please try again."
	CanceledMessage = "Operation canceled successfully"
)

type Repository interface {
	CreateUserInfoTable() error
	CreateUserHistoryTable() error
	UpdateBalance(chatID int64, update int64) error
	GetBalance(chatID int64) (int64, error)
	AddUser(chatID int64) error
	UserExists(chatID int64) (bool, error)
	AddToHistory(chatID int64, data string) error
}

type App interface {
	Start(chatID int64)
	TopUpBalance(chantID int64, value int64)
	SendMessage(chatID int64, message string)
	Buy(chatID int64, name string, value int64)
	GetBalance(chatID int64)
}

type FinancialApp struct {
	bot  *tgbotapi.BotAPI
	repo Repository
}

var YesNoButtons = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Yes"),
		tgbotapi.NewKeyboardButton("No"),
	),
)

var CategoriesButtons = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Entertainment"),
		tgbotapi.NewKeyboardButton("Food"),
		tgbotapi.NewKeyboardButton("Housing and communal services"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Transport"),
		tgbotapi.NewKeyboardButton("Cafes and restaurants"),
		tgbotapi.NewKeyboardButton("Other"),
	),
)

var EmptyReplyKeyboard = tgbotapi.NewRemoveKeyboard(true)

func (fb *FinancialApp) Start(chatID int64) {
	defer fb.SendMessage(chatID, "Hi, I'm the bot for your finances")

	if ok, err := fb.repo.UserExists(chatID); err != nil {
		log.Println(err)
	} else if ok {
		return
	}
	if err := fb.repo.AddUser(chatID); err != nil {
		log.Println(err)
		return
	}
}

func (fb *FinancialApp) TopUpBalance(chatID int64, value int64) {
	defer ClearUserState(chatID)

	curBalance, err := fb.repo.GetBalance(chatID)
	if err != nil {
		log.Println("GetBalance failed, result: ", err)
		fb.SendMessage(chatID, FailedMessage)
		return
	}
	curBalance += value

	us, _ := UserStates[chatID]

	msg := tgbotapi.NewMessage(chatID,
		"\nAre you sure you want to increase your balance by "+strconv.FormatInt(value, 10)+" rubles?\n"+
			"Enter yes OR no")
	msg.ReplyMarkup = YesNoButtons
	_, _ = fb.bot.Send(msg)

	var finalMsg string

	defer func(msg *string) {
		tgMsg := tgbotapi.NewMessage(chatID, finalMsg)
		tgMsg.ReplyMarkup = EmptyReplyKeyboard
		fb.bot.Send(tgMsg)
	}(&finalMsg)

	for {
		select {
		case val, _ := <-us.operationStateCh:
			if val == CONTINUE {
				msg := strings.TrimSpace(strings.ToLower(<-us.message))
				if msg == "yes" {
					err = fb.repo.UpdateBalance(chatID, curBalance)
					if err != nil {
						finalMsg = FailedMessage
						return
					}
					finalMsg = SuccessMessage
					return
				} else if msg == "no" {
					finalMsg = CanceledMessage
					return
				} else {
					fb.SendMessage(chatID, "Please, enter \"yes\" OR \"no\" ")
				}
			} else {
				finalMsg = CanceledMessage
				return
			}
		}
	}
}

func (fb *FinancialApp) SendMessage(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	_, _ = fb.bot.Send(msg)
}

func (fb *FinancialApp) Buy(chatID int64, name string, value int64) {
	defer ClearUserState(chatID)

	tgMsg := tgbotapi.NewMessage(chatID, "Please select the category to which the "+name+" belongs")
	tgMsg.ReplyMarkup = CategoriesButtons
	_, _ = fb.bot.Send(tgMsg)

	us, _ := UserStates[chatID]

	var finalMsg string

	defer func(msg *string) {
		tgMsg := tgbotapi.NewMessage(chatID, finalMsg)
		tgMsg.ReplyMarkup = EmptyReplyKeyboard
		fb.bot.Send(tgMsg)
	}(&finalMsg)

	for {
		select {
		case val, _ := <-us.operationStateCh:
			if val == CONTINUE {
				curBalance, _ := fb.repo.GetBalance(chatID)

				curBalance += value
				err := fb.repo.UpdateBalance(chatID, curBalance)
				if err != nil {
					finalMsg = FailedMessage
					return
				}
				msg := <-us.message
				msg += " " + name + " " + strconv.FormatInt(value, 10)
				err = fb.repo.AddToHistory(chatID, msg)
				if err != nil {
					finalMsg = FailedMessage
					return
				}
				finalMsg = SuccessMessage
				return
			} else {
				finalMsg = CanceledMessage
				return
			}
		}
	}
}

func (fb *FinancialApp) GetBalance(chatID int64) {
	balance, err := fb.repo.GetBalance(chatID)
	if err != nil {
		fb.SendMessage(chatID, "Something went wrong")
		return
	}
	fb.SendMessage(chatID, strconv.FormatInt(balance, 10))
}

func New(bot *tgbotapi.BotAPI, repo Repository) App {
	return &FinancialApp{bot: bot, repo: repo}
}
