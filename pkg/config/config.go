package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort   string
	DBPath       string
	UploadsDir   string
	SecretKey    string
	RegistrationCode string // <--- Додали
	RPID             string
	AppURL           string
	
	// Telegram
	TgBotToken          string
	TgChatID            string
	TgReceiptBotToken   string
	TgReceiptBotUsername string
	TgReceiptWebhookSecret string
}

func LoadConfig() *Config {
	// Спробуємо завантажити .env файл (якщо ми локально)
	// Якщо файлу немає (на сервері), помилку ігноруємо і читаємо змінні середовища
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		DBPath:       getEnv("DB_PATH", "finance.db"),
		UploadsDir:   getEnv("UPLOADS_DIR", "./uploads"),
		SecretKey:        getEnv("JWT_SECRET", "secret"),
    RegistrationCode: getEnv("REGISTRATION_CODE", "admin"),
		RPID:             getEnv("RP_ID", "localhost"),
		AppURL:           getEnv("APP_URL", "http://localhost:3000"),
		
		TgBotToken:              getEnv("TG_BOT_TOKEN", ""),
		TgChatID:                getEnv("TG_CHAT_ID", ""),
		TgReceiptBotToken:       getEnv("TG_RECEIPT_BOT_TOKEN", ""),
		TgReceiptBotUsername:    getEnv("TG_RECEIPT_BOT_USERNAME", ""),
		TgReceiptWebhookSecret:  getEnv("TG_RECEIPT_WEBHOOK_SECRET", ""),
	}
}

// Допоміжна функція: читає змінну або повертає дефолт
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
