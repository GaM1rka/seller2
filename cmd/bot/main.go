package main

import (
	"context"
	"log"
	"os"

	"seller2/config"
	"seller2/internal/bot"
	"seller2/internal/store"
)

func main() {
	cfg := config.Load()
	debug := os.Getenv("DEBUG") == "1"

	// Redis
	rds := store.NewRedis(cfg.RedisAddr, cfg.RedisPass, cfg.RedisDB)
	if _, err := rds.Ping(context.Background()); err != nil {
		// добавь метод Ping в store (обертку rdb.Ping)
		log.Println("redis ping:", err)
	}

	// Bot
	b := bot.New(cfg.BotToken, debug)
	h := bot.NewHandlerWithStore(b, cfg, rds)
	// запускаем планировщик удаления
	go h.RunDeletionScheduler(context.Background())

	log.Printf("Bot authorized as @%s", b.API.Self.UserName)
	h.Start()
}
