package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken   string
	PriceText  string
	TributeURL string

	RedisAddr string // "localhost:6379"
	RedisDB   int    // 0
	RedisPass string // "" если без пароля
}

func Load() Config {
	_ = godotenv.Load()
	c := Config{
		BotToken:   getenv("TELEGRAM_BOT_TOKEN", ""), // ← читаем по имени
		PriceText:  getenv("PRICE_TEXT", "9 900 ₽"),
		TributeURL: getenv("TRIBUTE_URL", "https://pay.tribute.to/your-product"),

		RedisAddr: getenv("REDIS_ADDR", "127.0.0.1:6380"),
		RedisDB:   atoi(getenv("REDIS_DB", "0")),
		RedisPass: getenv("REDIS_PASSWORD", ""),
	}
	if c.BotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is empty")
	}
	return c
}

func atoi(s string) int { i, _ := strconv.Atoi(s); return i }

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
