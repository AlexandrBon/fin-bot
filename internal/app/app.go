package app

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"log"
	"strings"
)

const (
	SuccessMessage  = "Operation completed successfully!"
	FailedMessage   = "Something went wrong, please try again."
	CanceledMessage = "Operation canceled successfully"
)

type Repository interface {
	CreateUserInfoTable() error
	UpdateBalance(chatID int64, update int64) error
	GetBalance(chatID int64) (int64, error)
}

type App interface {
	Start(chatID int64)
	TopUpBalance(chantID int64, value int64)
	SendMessage(chatID int64, message string)
	Buy(chatID int64, name string, value int64)
	GetHistory(chatID int64, dayCnt int64)
}

type FinancialApp struct {
	bot  *tgbotapi.BotAPI
	repo Repository
}

var AmountButtons = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("+1000"),
		tgbotapi.NewKeyboardButton("+2000"),
		tgbotapi.NewKeyboardButton("+5000"),
	),
)

var EmptyReplyKeyboard = tgbotapi.NewRemoveKeyboard(true)

func getSuccessOperationMessage(chatID int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, "Operation completed successfully!")
	msg.ReplyMarkup = EmptyReplyKeyboard
	return msg
}

func (fb *FinancialApp) Start(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Hi, I'm the bot for your finances")
	_, _ = fb.bot.Send(msg)
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

	for {
		select {
		case val, _ := <-us.operationStateCh:
			if val == CONTINUE {
				msg := strings.TrimSpace(strings.ToLower(<-us.message))
				if msg == "yes" {
					err = fb.repo.UpdateBalance(chatID, curBalance)
					if err != nil {
						fb.SendMessage(chatID, FailedMessage)
						return
					}
					fb.SendMessage(chatID, SuccessMessage)
					return
				} else if msg == "no" {
					fb.SendMessage(chatID, CanceledMessage)
					return
				} else {
					fb.SendMessage(chatID, "Please, enter \"yes\" OR \"no\" ")
				}
			} else {
				fb.SendMessage(chatID, CanceledMessage)
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
}

func (fb *FinancialApp) GetHistory(chatID int64, dayCnt int64) {

}

func New(bot *tgbotapi.BotAPI, repo Repository) App {
	return &FinancialApp{bot: bot, repo: repo}
}
