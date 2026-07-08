package main

import (
	"log"

	"github.com/VladHrytsaiuk/wegas-finance/backend/controllers"
	"github.com/VladHrytsaiuk/wegas-finance/backend/middlewares"
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/pkg/config"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/routes"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type appRuntime struct {
	controllers     routes.AppControllers
	goalService     services.GoalService
	currencyService services.CurrencyService
	monobankService services.MonobankService
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
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
		&models.ShoppingList{},
		&models.ShoppingItem{},
		&models.WishlistGroup{},
		&models.WishlistItem{},
		&models.FamilyJoinCode{},
		&models.MedicalRecord{},
		&models.MedicalFile{},
		&models.WebAuthnCredential{},
	)
}

func startSystemSeed(db *gorm.DB) {
	go func() {
		if err := utils.SeedSystemStorageTypes(db); err != nil {
			log.Printf("⚠️ Seed error: %v", err)
		}
	}()
}

func buildAppRuntime(cfg *config.Config, db *gorm.DB) (*appRuntime, error) {
	userRepo := repositories.NewUserRepository(db)
	waRepo := repositories.NewWebAuthnRepository(db)
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
	wishlistRepo := repositories.NewWishlistRepo(db)
	familyJoinRepo := repositories.NewFamilyJoinRepository(db)

	wsHub := utils.NewWSHub()
	go wsHub.Run()

	storageService := services.NewLocalStorageService(cfg.UploadsDir)
	clock := utils.NewRealClock()

	importService := services.NewImportService(db)
	currencyService := services.NewCurrencyService(db)
	categoryService := services.NewCategoryService(categoryRepo)
	cpService := services.NewCounterpartyService(cpRepo)
	jwtService := services.NewJWTService(cfg.SecretKey)

	userService := services.NewUserService(userRepo, wsHub, db)
	authService := services.NewAuthService(userRepo, waRepo, jwtService, cfg.SecretKey, cfg.RegistrationCode)
	accountService := services.NewAccountService(accountRepo, db)
	tagService := services.NewTagService(tagRepo)
	txService := services.NewTransactionService(db, txRepo, cpRepo, assetRepo, storageService, clock)
	exportService := services.NewExportService(exportRepo)
	roleService := services.NewRoleService(roleRepo)
	statsService := services.NewStatsService(statsRepo, currencyService)
	assetService := services.NewAssetService(assetRepo, txRepo, storageService, clock)
	utilityService := services.NewUtilityService(utilityRepo, txRepo, assetRepo)
	monobankService := services.NewMonobankService(db, txService, accountRepo, clock)
	goalService := services.NewGoalService(goalRepo, accountRepo, currencyService)
	storageTypeService := services.NewStorageTypeService(storageTypeRepo)
	feedbackService := services.NewFeedbackService(cfg.TgBotToken, cfg.TgChatID)
	shoppingService := services.NewShoppingService(shoppingRepo)
	wishlistService := services.NewWishlistService(wishlistRepo)
	familyJoinService := services.NewFamilyJoinService(familyJoinRepo, userRepo, wsHub, db)

	waService, err := services.NewWebAuthnService(cfg.RPID, cfg.AppURL, waRepo, userRepo)
	if err != nil {
		return nil, err
	}

	appControllers := routes.AppControllers{
		Auth:         controllers.NewAuthController(authService),
		User:         controllers.NewUserController(userService),
		Account:      controllers.NewAccountController(accountService),
		Category:     controllers.NewCategoryController(categoryService),
		Counterparty: controllers.NewCounterpartyController(cpService),
		Tag:          controllers.NewTagController(tagService),
		Transaction:  controllers.NewTransactionController(txService),
		Dashboard:    controllers.NewDashboardController(statsService, userRepo),
		Role:         controllers.NewRoleController(roleService),
		Export:       controllers.NewExportController(exportService),
		Import:       controllers.NewImportController(importService),
		Settings:     controllers.NewSettingsController(userRepo),
		Monobank:     controllers.NewMonobankController(monobankService),
		Asset:        controllers.NewAssetController(assetService),
		Utility:      controllers.NewUtilityController(utilityService),
		Goal:         controllers.NewGoalController(goalService),
		StorageType:  controllers.NewStorageTypeController(storageTypeService),
		Currency:     controllers.NewCurrencyController(currencyService),
		Feedback:     controllers.NewFeedbackController(feedbackService),
		Shopping:     controllers.NewShoppingController(shoppingService),
		Wishlist:     controllers.NewWishlistController(wishlistService),
		Family:       controllers.NewFamilyController(familyJoinService),
		WS:           controllers.NewWSController(wsHub, cfg.AllowedOrigins),
		WebAuthn:     controllers.NewWebAuthnController(waService, jwtService, userRepo),
	}

	return &appRuntime{
		controllers:     appControllers,
		goalService:     goalService,
		currencyService: currencyService,
		monobankService: monobankService,
	}, nil
}

func newRouter(cfg *config.Config, appControllers routes.AppControllers) *gin.Engine {
	router := gin.Default()
	router.MaxMultipartMemory = 10 << 20
	router.Use(middlewares.CORSMiddleware(cfg.AllowedOrigins))
	routes.SetupRoutes(router, appControllers, cfg)
	return router
}
