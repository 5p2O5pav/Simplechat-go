package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken            string
	ChatID              string
	ExpireDays          int
	MaxFileSize         int64
	RateLimitCount      int
	RateLimitBanHours   int
	MonitorTopicID      string
	WebhookSecret       string
	Domain              string
	Port                string
	DBPath              string
}

var AppConfig *Config

func Load() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, using environment variables")
	}

	cfg := &Config{
		BotToken:          getEnv("BOT_TOKEN", ""),
		ChatID:            getEnv("CHAT_ID", ""),
		ExpireDays:        getEnvInt("EXPIRE_DAYS", 30),
		MaxFileSize:       int64(getEnvInt("MAX_FILE_SIZE", 50*1024*1024)),
		RateLimitCount:    getEnvInt("RATE_LIMIT_COUNT", 20),
		RateLimitBanHours: getEnvInt("RATE_LIMIT_BAN_HOURS", 48),
		MonitorTopicID:    getEnv("MONITOR_TOPIC_ID", ""),
		WebhookSecret:     getEnv("WEBHOOK_SECRET", ""),
		Domain:            getEnv("DOMAIN", ""),
		Port:              getEnv("PORT", "3000"),
		DBPath:            getEnv("DB_PATH", "./data/chat.db"),
	}

	if cfg.BotToken == "" || cfg.ChatID == "" {
		log.Fatal("BOT_TOKEN and CHAT_ID are required")
	}
	if cfg.WebhookSecret == "" {
		log.Fatal("WEBHOOK_SECRET is required")
	}

	AppConfig = cfg
	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
