package routes

import (
	"os"

	"github.com/VladHrytsaiuk/wegas-finance/backend/controllers"
	_ "github.com/VladHrytsaiuk/wegas-finance/backend/docs"
	"github.com/VladHrytsaiuk/wegas-finance/backend/middlewares"
	"github.com/VladHrytsaiuk/wegas-finance/backend/pkg/config"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// AppControllers — контейнер для всіх контролерів
type AppControllers struct {
	Auth             *controllers.AuthController
	User             *controllers.UserController
	Account          *controllers.AccountController
	Category         *controllers.CategoryController
	Counterparty     *controllers.CounterpartyController
	Tag              *controllers.TagController
	Transaction      *controllers.TransactionController
	Inbox            *controllers.InboxController
	Dashboard        *controllers.DashboardController
	Role             *controllers.RoleController
	Settings         *controllers.SettingsController
	Export           *controllers.ExportController
	Import           *controllers.ImportController // <--- Added
	ReceiptIngestion *controllers.ReceiptIngestionController
	TelegramLink     *controllers.TelegramLinkController
	TelegramBot      *controllers.TelegramBotController
	TelegramWebhook  *controllers.TelegramWebhookController
	Monobank         *controllers.MonobankController // <--- Added
	Asset            *controllers.AssetController
	// Medical      *controllers.MedicalController
	Utility      *controllers.UtilityController
	Goal         *controllers.GoalController
	StorageType  *controllers.StorageTypeController
	Currency     *controllers.CurrencyController
	Feedback     *controllers.FeedbackController
	Shopping      *controllers.ShoppingController
	Wishlist      *controllers.WishlistController
	Family        *controllers.FamilyController
	WS            *controllers.WSController
	WebAuthn      *controllers.WebAuthnController
	AdminCatalog  *controllers.AdminCatalogController
	AdminUsers    *controllers.AdminUsersController
	AdminAudit    *controllers.AdminAuditController
	AdminOverview *controllers.AdminOverviewController
}

func SetupRoutes(r *gin.Engine, c AppControllers, cfg *config.Config) {

	_ = os.MkdirAll(cfg.UploadsDir, 0755)

	r.StaticFS("/uploads", gin.Dir(cfg.UploadsDir, true))

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		// === PUBLIC ===
		api.GET("/system/info", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{
				"rp_id":            cfg.RPID,
				"app_url":          cfg.AppURL,
				"status":           "running",
				"maintenance_mode": middlewares.IsMaintenanceMode(),
			})
		})
		api.POST("/users", c.Auth.Register)
		api.POST("/login", c.Auth.Login)
		api.POST("/login/pin", c.Auth.LoginWithPIN)
		api.POST("/feedback", c.Feedback.Submit)
		api.POST("/monobank/webhook", c.Monobank.Webhook)
		api.POST("/telegram/link/complete", c.TelegramLink.CompleteLink)
		api.POST("/telegram/webhook", c.TelegramBot.Webhook)

		// WebAuthn Public Endpoints
		api.POST("/webauthn/login/options", c.WebAuthn.LoginOptions)
		api.POST("/webauthn/login/verify", c.WebAuthn.LoginVerify)
		api.POST("/refresh", c.WebAuthn.Refresh)

		// === PROTECTED ===
		protected := api.Group("/")
		protected.Use(middlewares.AuthMiddleware(cfg.SecretKey))
		protected.Use(middlewares.MaintenanceMiddleware())
		{
			admin := protected.Group("/admin")
			admin.Use(middlewares.RequirePlatformAdmin())
			admin.GET("/status", func(ctx *gin.Context) {
				ctx.JSON(200, gin.H{"status": "ok"})
			})
			admin.GET("/catalog/categories", c.AdminCatalog.GetCategories)
			admin.POST("/catalog/categories", c.AdminCatalog.CreateCategory)
			admin.PUT("/catalog/categories/:id", c.AdminCatalog.UpdateCategory)
			admin.DELETE("/catalog/categories/:id", c.AdminCatalog.ArchiveCategory)
			admin.GET("/catalog/counterparty-categories", c.AdminCatalog.GetCounterpartyCategories)
			admin.POST("/catalog/counterparty-categories", c.AdminCatalog.CreateCounterpartyCategory)
			admin.PUT("/catalog/counterparty-categories/:id", c.AdminCatalog.UpdateCounterpartyCategory)
			admin.DELETE("/catalog/counterparty-categories/:id", c.AdminCatalog.ArchiveCounterpartyCategory)
			admin.GET("/catalog/counterparties", c.AdminCatalog.GetCounterparties)
			admin.POST("/catalog/counterparties", c.AdminCatalog.CreateCounterparty)
			admin.PUT("/catalog/counterparties/:id", c.AdminCatalog.UpdateCounterparty)
			admin.DELETE("/catalog/counterparties/:id", c.AdminCatalog.ArchiveCounterparty)

			// --- ADMIN USERS ---
			admin.GET("/users", c.AdminUsers.GetUsers)
			admin.POST("/users/:id/block", c.AdminUsers.ToggleBlock)
			admin.POST("/users/:id/logout", c.AdminUsers.ForceLogout)
			admin.POST("/users/:id/role", c.AdminUsers.SetRole)

			// --- ADMIN AUDIT & SETTINGS ---
			admin.GET("/audit", c.AdminAudit.GetLogs)
			admin.GET("/settings", c.AdminAudit.GetSettings)

			// --- ADMIN OVERVIEW ---
			admin.GET("/overview/stats", c.AdminOverview.GetStats)
			admin.POST("/maintenance", c.AdminOverview.ToggleMaintenance)

			// WebAuthn Protected Endpoints
			protected.POST("/webauthn/register/options", c.WebAuthn.RegisterOptions)
			protected.POST("/webauthn/register/verify", c.WebAuthn.RegisterVerify)

			protected.GET("/ws", c.WS.HandleWS)

			// --- DASHBOARD ---
			protected.GET("/dashboard/stats", c.Dashboard.GetStats)
			protected.GET("/dashboard/top", c.Dashboard.GetTopStats)
			protected.GET("/dashboard/trend", c.Dashboard.GetTrend)
			protected.GET("/dashboard/recent", c.Dashboard.GetRecent)

			// --- EXPORT / IMPORT ---
			protected.GET("/export/transactions", c.Export.ExportTransactions)
			protected.GET("/export/backup", c.Export.ExportBackup)
			protected.POST("/import/upload", c.Import.UploadStatement)
			protected.POST("/receipt-ingestion/xml", c.ReceiptIngestion.IngestXML)
			protected.POST("/receipt-ingestion/url", c.ReceiptIngestion.IngestURL)
			protected.GET("/integrations/telegram", c.TelegramLink.GetStatus)
			protected.POST("/integrations/telegram/link-token", c.TelegramLink.CreateLinkToken)
			protected.DELETE("/integrations/telegram/link", c.TelegramLink.RevokeLink)
			protected.GET("/integrations/telegram/webhook", c.TelegramWebhook.GetStatus)
			protected.POST("/integrations/telegram/webhook/sync", c.TelegramWebhook.SyncWebhook)
			protected.DELETE("/integrations/telegram/webhook", c.TelegramWebhook.DeleteWebhook)

			// --- PROFILE & USERS ---
			protected.GET("/users/me", c.User.GetMe)
			protected.PUT("/users/me", c.User.UpdateProfile)
			protected.GET("/users/security-status", c.Auth.GetSecurityStatus)
			protected.PUT("/users/password", c.User.ChangePassword)
			protected.POST("/users/pin", c.Auth.SetPIN)
			protected.DELETE("/users/pin", c.Auth.RemovePIN)
			protected.DELETE("/users/passkeys", c.Auth.RemovePasskeys)

			protected.GET("/currencies", c.Currency.GetRates)

			// --- FAMILY ---
			protected.GET("/users", c.User.GetFamilyMembers)
			protected.POST("/family/users", c.User.AddMember)
			protected.POST("/family/leave", c.User.LeaveFamily)
			protected.POST("/families/:id/generate-code", c.Family.GenerateCodeHandler)
			protected.POST("/families/join", c.Family.JoinFamilyHandler)
			protected.PUT("/users/:id", c.User.UpdateUser)
			protected.DELETE("/users/:id", c.User.DeleteMember)

			// --- ACCOUNTS ---
			protected.POST("/accounts", c.Account.Create)
			protected.GET("/accounts", c.Account.GetAll)
			protected.PUT("/accounts/mobile-order", c.Account.UpdateMobileOrder)
			protected.GET("/accounts/:id", c.Account.GetOne)
			protected.PATCH("/accounts/:id/round-up-target", c.Account.SetRoundUpTarget)
			protected.PUT("/accounts/:id", c.Account.Update)
			protected.DELETE("/accounts/:id", c.Account.Delete)

			// --- CATEGORIES ---
			protected.POST("/categories", c.Category.Create)
			protected.GET("/categories", c.Category.GetAll)
			protected.GET("/categories/:id", c.Category.GetOne)
			protected.PUT("/categories/:id", c.Category.Update)
			protected.DELETE("/categories/:id", c.Category.Delete)

			// --- TRANSACTIONS ---
			protected.POST("/transactions", c.Transaction.Create)
			protected.GET("/transactions", c.Transaction.GetAll)
			protected.GET("/transactions/:id", c.Transaction.GetOne)
			protected.DELETE("/transactions/:id", c.Transaction.Delete)
			protected.PUT("/transactions/:id", c.Transaction.Update)
			protected.POST("/transactions/batch", c.Transaction.BatchCreate)
			protected.DELETE("/transactions/photos/:id", c.Transaction.DeletePhoto)
			protected.GET("/transactions/predict", c.Transaction.PredictCategory)
			protected.GET("/transactions/:id/receipt-sources", c.Transaction.GetLinkedReceipts)
			protected.POST("/transactions/:id/receipt-sources/unlink", c.Transaction.UnlinkReceiptSource)

			// Receipts
			protected.POST("/transactions/:id/receipt", c.Transaction.UploadReceipt)
			protected.DELETE("/transactions/:id/receipt", c.Transaction.DeleteReceipt)

			// --- INBOX ---
			protected.POST("/inbox", c.Inbox.Create)
			protected.POST("/inbox/photo", c.Inbox.CreatePhoto)
			protected.GET("/inbox", c.Inbox.GetAll)
			protected.GET("/inbox/:id/account-candidates", c.Inbox.GetAccountCandidates)
			protected.GET("/inbox/:id/transaction-candidates", c.Inbox.GetTransactionCandidates)
			protected.GET("/inbox/:id", c.Inbox.GetOne)
			protected.DELETE("/inbox/:id", c.Inbox.Delete)
			protected.PATCH("/inbox/:id/account", c.Inbox.SelectAccount)
			protected.POST("/inbox/:id/link", c.Inbox.Link)
			protected.POST("/inbox/:id/unlink", c.Inbox.Unlink)

			// --- MONOBANK (Виправлено конфлікт шляхів) ---
			protected.POST("/monobank/connect", c.Monobank.Connect)
			protected.POST("/monobank/settings", c.Monobank.SaveSettings) // Зберегти налаштування
			protected.GET("/monobank/settings", c.Monobank.GetSettings)

			// 2. Refresh (API Request) - викликається по кнопці "Налаштувати"
			protected.POST("/monobank/refresh", c.Monobank.RefreshClientInfo)
			protected.POST("/monobank/sync-confirm", c.Monobank.ConfirmSync)
			protected.POST("/monobank/disconnect", c.Monobank.Disconnect)
			protected.GET("/monobank/status", c.Monobank.GetStatus) // <--- Додай
			protected.POST("/monobank/sync", c.Monobank.ForceSync)
			protected.POST("/monobank/force-resync", c.Monobank.ForceResyncCounterparties)

			// --- COUNTERPARTIES ---
			protected.POST("/counterparties", c.Counterparty.Create)
			protected.GET("/counterparties", c.Counterparty.GetAll)
			protected.GET("/counterparties/:id", c.Counterparty.GetOne)
			protected.DELETE("/counterparties/:id", c.Counterparty.Delete)
			protected.PUT("/counterparties/:id", c.Counterparty.Update)

			protected.POST("/counterparty-categories", c.Counterparty.CreateCategory)
			protected.GET("/counterparty-categories", c.Counterparty.GetCategories)
			protected.GET("/counterparty-categories/:id", c.Counterparty.GetCategory)
			protected.PUT("/counterparty-categories/:id", c.Counterparty.UpdateCategory)

			// --- TAGS ---
			protected.POST("/tags", c.Tag.Create)
			protected.GET("/tags", c.Tag.GetAll)
			protected.DELETE("/tags/:id", c.Tag.Delete)

			// --- ROLES ---
			protected.POST("/roles", c.Role.Create)
			protected.GET("/roles", c.Role.GetAll)
			protected.DELETE("/roles/:id", c.Role.Delete)

			// --- GENERAL SETTINGS ---
			protected.GET("/settings", c.Settings.GetSettingsHTTP) // Загальні налаштування
			protected.POST("/settings", c.Settings.SaveSettingsHTTP)

			// --- ASSETS ---
			protected.GET("/assets", c.Asset.GetAll)
			protected.POST("/assets", c.Asset.Create)
			protected.GET("/assets/:id", c.Asset.GetOne)
			protected.PUT("/assets/:id", c.Asset.Update)
			protected.DELETE("/assets/:id", c.Asset.Delete)
			protected.POST("/assets/:id/photo", c.Asset.UploadPhoto)
			protected.PATCH("/assets/:id/mileage", c.Asset.UpdateMileage)
			protected.DELETE("/assets/:id/photo", c.Asset.RemovePhoto)
			protected.POST("/assets/:id/documents", c.Asset.UploadDocument)
			protected.DELETE("/assets/:id/documents/:doc_id", c.Asset.RemoveDocument)

			// --- UTILITY ---
			protected.GET("/utility/meters", c.Utility.GetMeters)
			protected.POST("/utility/meters", c.Utility.CreateMeter)
			protected.GET("/utility/meters/:id", c.Utility.GetMeter)
			protected.PUT("/utility/meters/:id", c.Utility.UpdateMeter)
			protected.DELETE("/utility/meters/:id", c.Utility.DeleteMeter)

			protected.GET("/utility/readings", c.Utility.GetReadings)
			protected.POST("/utility/readings", c.Utility.CreateReading)
			protected.PATCH("/utility/readings/:id", c.Utility.UpdateReading)
			protected.PUT("/utility/readings/:id", c.Utility.UpdateReading)
			protected.DELETE("/utility/readings/:id", c.Utility.DeleteReading)
			protected.POST("/utility/readings/:id/pay", c.Utility.PayReading)

			protected.GET("/utility/stats/global", c.Utility.GetGlobalStats)
			protected.GET("/utility/stats/meter/:id", c.Utility.GetMeterStats)

			// --- GOALS (Фінансові цілі) ---
			protected.POST("/goals", c.Goal.Create)
			protected.GET("/goals", c.Goal.GetAll)
			protected.GET("/goals/:id", c.Goal.GetOne)
			protected.PUT("/goals/:id", c.Goal.Update)
			protected.DELETE("/goals/:id", c.Goal.Delete)
			protected.POST("/goals/:id/photo", c.Goal.UploadPhoto)

			// Спеціальний роут для відв'язування рахунку від цілі (або це можна робити через PUT goals)
			protected.POST("/goals/:id/link-account", c.Goal.LinkAccount)
			protected.POST("/goals/:id/unlink-account", c.Goal.UnlinkAccount)

			// --- STORAGE TYPES (Типи скарбничок) ---
			protected.GET("/storage-types", c.StorageType.GetAll) // Отримати і системні, і свої
			protected.POST("/storage-types", c.StorageType.Create)
			protected.DELETE("/storage-types/:id", c.StorageType.Delete)

			// --- СПИСКИ (НОТАТКИ) ---
			protected.GET("/shopping-lists", c.Shopping.GetLists)
			protected.POST("/shopping-lists", c.Shopping.CreateList)
			protected.PUT("/shopping-lists/:id", c.Shopping.UpdateList)
			protected.DELETE("/shopping-lists/:id", c.Shopping.DeleteList)

			// --- ЕЛЕМЕНТИ ВСЕРЕДИНІ НОТАТКИ ---
			protected.POST("/shopping-lists/:id/items", c.Shopping.AddItem)
			protected.PUT("/shopping-items/:id", c.Shopping.UpdateItem)
			protected.DELETE("/shopping-items/:id", c.Shopping.DeleteItem)
			protected.DELETE("/shopping-lists/:id/completed", c.Shopping.ClearCompletedInList)

			// --- WISHLIST ---
			protected.GET("/wishlist-groups", c.Wishlist.GetGroups)
			protected.POST("/wishlist-groups", c.Wishlist.CreateGroup)
			protected.PUT("/wishlist-groups/:id", c.Wishlist.UpdateGroup)
			protected.DELETE("/wishlist-groups/:id", c.Wishlist.DeleteGroup)

			protected.GET("/wishlist", c.Wishlist.GetAll)
			protected.POST("/wishlist", c.Wishlist.Create)
			protected.GET("/wishlist/:id", c.Wishlist.GetOne)
			protected.PUT("/wishlist/:id", c.Wishlist.Update)
			protected.DELETE("/wishlist/:id", c.Wishlist.Delete)
			protected.POST("/wishlist/:id/reserve", c.Wishlist.ToggleReservation)
			protected.POST("/wishlist/:id/photo", c.Wishlist.UploadPhoto)
			protected.DELETE("/wishlist/:id/photo", c.Wishlist.DeletePhoto)
		}
	}
}
