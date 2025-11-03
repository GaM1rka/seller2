package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API *tgbotapi.BotAPI
}

func New(token string, debug bool) *Bot {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	api.Debug = debug
	return &Bot{API: api}
}

func (b *Bot) Updates() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	return b.API.GetUpdatesChan(u)
}
