package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort       string
	DBPath           string
	UploadsDir       string
	SecretKey        string
	RegistrationCode string // <--- Додали
	RPID             string
	AppURL           string
	AllowedOrigins   []string

	// Telegram
	TgBotToken string
	TgChatID   string
}

func LoadConfig() *Config {
	// Спробуємо завантажити .env файл (якщо ми локально)
	// Якщо файлу немає (на сервері), помилку ігноруємо і читаємо змінні середовища
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	appURL := getEnv("APP_URL", "http://localhost:3000")

	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		DBPath:           sanitizePath(getEnv("DB_PATH", "finance.db"), "finance.db"),
		UploadsDir:       sanitizePath(getEnv("UPLOADS_DIR", "./uploads"), "./uploads"),
		SecretKey:        mustGetEnv("JWT_SECRET"),
		RegistrationCode: mustGetEnv("REGISTRATION_CODE"),
		RPID:             getEnv("RP_ID", "localhost"),
		AppURL:           appURL,
		AllowedOrigins:   buildAllowedOrigins(getEnv("CORS_ALLOWED_ORIGINS", ""), appURL),

		TgBotToken: getEnv("TG_BOT_TOKEN", ""),
		TgChatID:   getEnv("TG_CHAT_ID", ""),
	}
}

// Допоміжна функція: читає змінну або повертає дефолт
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func mustGetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Fatal(fmt.Sprintf("required environment variable %s is not set", key))
	}

	return value
}

func sanitizePath(value, fallback string) string {
	if value == "" {
		return fallback
	}

	if runtime.GOOS != "windows" && isWindowsPath(value) {
		log.Printf("invalid Windows-style path %q on %s, using %q instead", value, runtime.GOOS, fallback)
		return fallback
	}

	return value
}

func isWindowsPath(value string) bool {
	return strings.Contains(value, "\\") || strings.Contains(value, ":")
}

func buildAllowedOrigins(rawOrigins, appURL string) []string {
	origins := splitAndNormalizeOrigins(rawOrigins)
	if len(origins) > 0 {
		return origins
	}

	defaults := []string{
		normalizeOrigin(appURL),
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://localhost:3000",
		"http://127.0.0.1:3000",
	}

	return dedupeOrigins(defaults)
}

func splitAndNormalizeOrigins(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := normalizeOrigin(strings.TrimSpace(part))
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	return dedupeOrigins(origins)
}

func normalizeOrigin(raw string) string {
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	return parsed.Scheme + "://" + parsed.Host
}

func dedupeOrigins(origins []string) []string {
	result := make([]string, 0, len(origins))
	for _, origin := range origins {
		if origin == "" || slices.Contains(result, origin) {
			continue
		}
		result = append(result, origin)
	}
	return result
}
