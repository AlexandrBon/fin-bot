package main

import (
	"github.com/Syfaro/telegram-bot-api"
	"log"
	"os"
	"tgbot/internal/app"
	"tgbot/internal/repo"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal("NewBotAPI(...) failed, result: ", err)
	}

	finRepo, err := repo.New()
	if err != nil {
		log.Fatal("repo.New() failed, result: ", err)
	}
	finApp := app.New(bot, finRepo)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)
	for upd := range updates {
		app.HandleUpdate(finApp, upd)
	}
}
