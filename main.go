package main

import (
	"log"
	"os"

	// 🔥 ДОДАЙ ЦЕЙ ІМПОРТ
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/controllers"
	"github.com/VladHrytsaiuk/wegas-finance/backend/database"
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/pkg/config"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/routes"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/gin-gonic/gin"
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


////////

// ex, err := os.Executable()
  // if err != nil {
  //   log.Fatal(err)
  // }
  // // 2. Отримуємо папку цього файлу (відрізаємо wegas-finance.exe)
  // exPath := filepath.Dir(ex)

	// if err := os.Chdir(exPath); err != nil {
  //     log.Fatal(err)
  // }
  // log.Printf("📂 Working directory set to: %s", exPath)

	//////////



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

	log.Println("🔄 Running database migrations...")
	err := db.AutoMigrate(
		&models.Family{}, 
		&models.User{}, 
		&models.Role{}, 
		&models.Account{},
		&models.Goal{}, 
		&models.StorageType{}, 
		&models.Category{}, 
		&models.Counterparty{},
		&models.CounterpartyCategory{}, 
		&models.CounterpartyBalance{}, 
		&models.Tag{},
		&models.ExchangeRate{}, 
		&models.Transaction{}, 
		&models.TransactionItem{},
		&models.TransactionTag{}, 
		&models.TransactionPhoto{}, 
		&models.Asset{},
		&models.AssetPhoto{}, 
		&models.AssetDocument{},
		&models.UtilityMeter{}, 
		&models.UtilityReading{}, 
		&models.BankConnection{},
		&models.BankAccountMapping{},
		&models.ShoppingList{}, // <--- ОБОВ'ЯЗКОВО ДОДАЙ ЦЕ
    &models.ShoppingItem{},
		&models.WishlistGroup{},
		&models.WishlistItem{},
		&models.FamilyJoinCode{},
		&models.MedicalRecord{},
		&models.MedicalFile{},
	)
	if err != nil {
		log.Fatal("❌ Migration error:", err)
	}

	go func() {
		if err := utils.SeedSystemStorageTypes(db); err != nil {
			log.Printf("⚠️ Seed error: %v", err)
		}
	}()

	// Ініціалізація всіх шарів (твій оригінальний код)
	userRepo := repositories.NewUserRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	cpRepo := repositories.NewCounterpartyRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	txRepo := repositories.NewTransactionRepository(db)
	statsRepo := repositories.NewStatsRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	exportRepo := repositories.NewExportRepository(db)
	assetRepo := repositories.NewAssetRepository(db)
	utilityRepo := repositories.NewUtilityRepository(db)
	goalRepo := repositories.NewGoalRepository(db)
	storageTypeRepo := repositories.NewStorageTypeRepository(db)
	shoppingRepo := repositories.NewShoppingRepo(db)
	wishlistRepo := repositories.NewWishlistRepo(db) // <--- 2. ДОДАНО REPO
	familyJoinRepo := repositories.NewFamilyJoinRepository(db)

	wsHub := utils.NewWSHub()
	go wsHub.Run()

	storageService := services.NewLocalStorageService("./uploads")

	importService := services.NewImportService(db)
	currencyService := services.NewCurrencyService(db)
	categoryService := services.NewCategoryService(categoryRepo)
	cpService := services.NewCounterpartyService(cpRepo)
	userService := services.NewUserService(userRepo, wsHub, db)
	authService := services.NewAuthService(userRepo, cfg.SecretKey, cfg.RegistrationCode)
	accountService := services.NewAccountService(accountRepo, db)
	tagService := services.NewTagService(tagRepo)
	txService := services.NewTransactionService(db, txRepo, cpRepo, assetRepo, storageService)
	exportService := services.NewExportService(exportRepo)
	roleService := services.NewRoleService(roleRepo)
	statsService := services.NewStatsService(statsRepo, currencyService)
	assetService := services.NewAssetService(assetRepo, txRepo, storageService)
	utilityService := services.NewUtilityService(utilityRepo, txRepo, assetRepo)
	monobankService := services.NewMonobankService(db, txService, accountRepo)
	goalService := services.NewGoalService(goalRepo, accountRepo, currencyService)
	storageTypeService := services.NewStorageTypeService(storageTypeRepo)
	feedbackService := services.NewFeedbackService(cfg.TgBotToken, cfg.TgChatID)
	shoppingService := services.NewShoppingService(shoppingRepo)
	wishlistService := services.NewWishlistService(wishlistRepo) // <--- 3. ДОДАНО SERVICE
	familyJoinService := services.NewFamilyJoinService(familyJoinRepo, userRepo, wsHub, db)

	startSchedulers(db, goalService, currencyService, monobankService)

	appControllers := routes.AppControllers{
		Auth: controllers.NewAuthController(authService), User: controllers.NewUserController(userService),
		Account: controllers.NewAccountController(accountService), Category: controllers.NewCategoryController(categoryService),
		Counterparty: controllers.NewCounterpartyController(cpService), Tag: controllers.NewTagController(tagService),
		Transaction: controllers.NewTransactionController(txService), Dashboard: controllers.NewDashboardController(statsService, userRepo),
		Role: controllers.NewRoleController(roleService), Export: controllers.NewExportController(exportService),
		Import: controllers.NewImportController(importService), Settings: controllers.NewSettingsController(userRepo),
		Monobank: controllers.NewMonobankController(monobankService), Asset: controllers.NewAssetController(assetService),
		Utility: controllers.NewUtilityController(utilityService), Goal: controllers.NewGoalController(goalService),
		StorageType: controllers.NewStorageTypeController(storageTypeService), Currency: controllers.NewCurrencyController(currencyService),
		Feedback: controllers.NewFeedbackController(feedbackService),
		Shopping: controllers.NewShoppingController(shoppingService),
		Wishlist:     controllers.NewWishlistController(wishlistService), // <--- 4. ДОДАНО КОНТРОЛЕР
		Family:       controllers.NewFamilyController(familyJoinService),
		WS:           controllers.NewWSController(wsHub),
	}

	r := gin.Default()
	r.MaxMultipartMemory = 10 << 20

	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Дозволяємо будь-який Origin, який прийшов (щоб точно працювало)
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Якщо Origin порожній (іноді буває через проксі), ставимо зірочку або Vercel
			c.Writer.Header().Set("Access-Control-Allow-Origin", "https://wegas-finance.vercel.app")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		// Обробка Preflight запиту (OPTIONS)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	

	routes.SetupRoutes(r, appControllers, cfg.UploadsDir, cfg.SecretKey)

	log.Printf("🚀 Server starting on port %s...", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("❌ Failed to run server:", err)
	}
}

// startSchedulers запускає фонові завдання в горутинах
func startSchedulers(
	db *gorm.DB,
	goalService *services.GoalService,
	currencyService services.CurrencyService,
	monobankService *services.MonobankService,
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