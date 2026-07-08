package main

import (
	"log"
	"os"

	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/database"
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/pkg/config"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/kardianos/service"
	"gorm.io/gorm"
)

// program реалізує інтерфейс сервісу
type program struct{}

func (p *program) Start(s service.Service) error {
	// Не блокуємо Start, запускаємо сервер у горутині
	go p.run()
	return nil
}

func (p *program) run() {
	startApp()
}

func (p *program) Stop(s service.Service) error {
	return nil
}

// @title WeGaS Finance API
// @version 1.0
// @description API for WeGaS Finance application.
// @contact.name Vlad Hrytsaiuk

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @BasePath /api

func main() {

	///

	// // // 1. Отримуємо шлях до самого EXE файлу
	// ex, err := os.Executable()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // 2. Отримуємо папку цього файлу (відрізаємо wegas-finance.exe)
	// exPath := filepath.Dir(ex)

	// // 3. Змінюємо робочу директорію на папку з EXE
	// // Це критично для Windows сервісів, які за замовчуванням стартують у System32
	// if err := os.Chdir(exPath); err != nil {
	// 	log.Fatal(err)
	// }
	// log.Printf("📂 Working directory set to: %s", exPath)

	///

	svcConfig := &service.Config{
		Name:        "WeGaSFinance",
		DisplayName: "WeGaS Finance Service",
		Description: "Backend service for WeGaS Finance Application (Lviv Polytechnic)",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Керування сервісом (install, uninstall тощо)
	if len(os.Args) > 1 {
		err := service.Control(s, os.Args[1])
		if err != nil {
			log.Printf("Valid actions: install, uninstall, start, stop, restart")
			log.Fatal(err)
		}
		return
	}

	// ПЕРЕВІРКА: Якщо запускаємо вручну (на Маці або просто через .exe)
	if service.Interactive() {
		log.Println("💻 Running in interactive mode...")
		startApp()
		return
	}

	// Якщо запускаємо як реальний Windows сервіс
	log.Println("⚙️ Running as a service...")
	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func startApp() {
	cfg := config.LoadConfig()
	db := database.InitDB(cfg.DBPath)

	if err := runMigrations(db); err != nil {
		log.Fatal("❌ Migration error:", err)
	}

	startSystemSeed(db)

	runtime, err := buildAppRuntime(cfg, db)
	if err != nil {
		log.Fatal("❌ Failed to initialize app runtime:", err)
	}

	startSchedulers(db, runtime.goalService, runtime.currencyService, runtime.monobankService)
	r := newRouter(cfg, runtime.controllers)

	log.Printf("🚀 Server starting on port %s...", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("❌ Failed to run server:", err)
	}
}

// startSchedulers запускає фонові завдання в горутинах
func startSchedulers(
	db *gorm.DB,
	goalService services.GoalService,
	currencyService services.CurrencyService,
	monobankService services.MonobankService,
) {
	// 1. Перевірка цілей (Goals Check) щоночі
	go func() {
		for {
			now := time.Now()
			// Плануємо на 00:00:01 наступного дня
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 1, 0, now.Location())
			if now.After(nextRun) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			duration := nextRun.Sub(now)
			log.Printf("📅 Goals Check scheduled in %v", duration)
			time.Sleep(duration)

			log.Println("🎯 Goals Check: Waking up! Checking overdue goals...")
			if err := goalService.CheckOverdueGoals(); err != nil {
				log.Printf("❌ Error checking goals: %v", err)
			} else {
				log.Println("✅ Goals check finished successfully")
			}
		}
	}()

	// 2. Оновлення курсів валют (Currency Sync) раз на 24 години
	go func() {
		log.Println("💱 Starting initial currency sync...")
		if err := currencyService.SyncRates(); err != nil {
			log.Printf("❌ Error syncing rates: %v", err)
		}

		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			log.Println("💱 Scheduled: Updating currency rates...")
			if err := currencyService.SyncRates(); err != nil {
				log.Printf("❌ Error updating rates: %v", err)
			}
		}
	}()

	// 3. Авто-синхронізація банків (Monobank) о 03:00
	go func() {
		for {
			now := time.Now()
			// Плануємо на 03:00:00
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
			if now.After(nextRun) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			duration := nextRun.Sub(now)
			log.Printf("🤖 Auto-Sync scheduled in %v", duration)
			time.Sleep(duration)

			log.Println("🤖 Auto-Sync: Waking up! Checking accounts...")
			var connections []models.BankConnection

			// Знаходимо всі активні підключення
			if err := db.Where("is_active = ?", true).Find(&connections).Error; err == nil {
				for _, conn := range connections {
					log.Printf("🤖 Auto-Sync: Starting for user %s", conn.UserID)
					monobankService.Sync(conn.UserID, "")
				}
			}
		}
	}()
}
