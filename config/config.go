package config

import (
	"log"
	"os"
)

type Config struct {
	BotToken   string
	PriceText  string // подставляется в оффер вместо [цена]
	TributeURL string // ссылка «Взять доступ»
}

func Load() Config {
	c := Config{
		BotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		PriceText:  getenv("PRICE_TEXT", "9 900 ₽"),
		TributeURL: getenv("TRIBUTE_URL", "https://pay.tribute.to/your-product"),
	}
	if c.BotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is empty")
	}
	return c
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
