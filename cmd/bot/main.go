package main

import (
	"log"
	"os"

	"seller2/config"
	"seller2/internal/bot"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // прочитает .env из корня проекта
	cfg := config.Load()

	debug := os.Getenv("DEBUG") == "1"
	b := bot.New(cfg.BotToken, debug)

	log.Printf("Bot authorized as @%s", b.API.Self.UserName)

	h := bot.NewHandler(b, cfg)
	h.Start()
}
