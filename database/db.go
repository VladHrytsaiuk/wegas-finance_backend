package database

import (
	"log"
	"os"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
func InitDB(dbPath string) *gorm.DB {
	var err error
	
	// Очищення шляху для Mac/Linux, якщо там затесалися Windows-символи
	finalPath := dbPath
	if !strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
		if strings.Contains(finalPath, ":") || strings.Contains(finalPath, "\\") {
			log.Println("⚠️ Виявлено шлях Windows на Unix-системі. Використовую локальну базу finance.db")
			finalPath = "finance.db"
		}
	}

	dsn := finalPath
	if !strings.Contains(dsn, "?") {
		dsn += "?_busy_timeout=5000&_journal_mode=WAL"
	}

	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("❌ Не вдалося підключитися до бази даних:", err)
	}

	log.Println("✅ База даних підключена успішно! Шлях:", finalPath)
	return DB
}