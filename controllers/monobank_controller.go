package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type MonobankController struct {
	service *services.MonobankService
}

func NewMonobankController(service *services.MonobankService) *MonobankController {
	return &MonobankController{service: service}
}

type MonobankConnectJSON struct {
	Token string `json:"token" binding:"required"`
}

// Connect godoc
// @Summary Connect to Monobank
// @Description Establishes a connection to Monobank using a provided API token.
// @Tags Monobank
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body MonobankConnectJSON true "Monobank API token"
// @Success 200 {object} map[string]interface{} "List of bank accounts"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 429 {object} map[string]string "Rate limit"
// @Router /monobank/connect [post]
func (ctrl *MonobankController) Connect(c *gin.Context) {
	// Отримуємо ID надійно через рядок (як у Middleware)
	userID := c.GetString("userID")
	// Також дістаємо FamilyID, якщо потрібно, або беремо з структури user
	// Для надійності можна взяти об'єкт, якщо middleware гарантує його наявність:
	var familyID string
	if u, exists := c.Get("user"); exists {
		familyID = u.(*models.User).FamilyID
	}

	var input MonobankConnectJSON

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accounts, err := ctrl.service.Connect(userID, familyID, input.Token)
	if err != nil {
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Monobank rate limit. Please wait 60 seconds."})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

// GetSettings godoc
// @Summary Get Monobank settings
// @Description Returns the current Monobank connection settings and account mappings from the database.
// @Tags Monobank
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Accounts and mappings"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Connection not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monobank/settings [get]
func (ctrl *MonobankController) GetSettings(c *gin.Context) {
	userID := c.GetString("userID") // Використовуємо userID з контексту

	accounts, mappings, err := ctrl.service.GetUserData(userID)
	if err != nil {
		// Якщо запис не знайдено — це 404 (Not Found), а не помилка сервера
		if strings.Contains(err.Error(), "no active connection") || strings.Contains(err.Error(), "record not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "No connection found"})
			return
		}

		// Якщо раптом помилка бази
		fmt.Printf("⚠️ DB Error in GetSettings: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service busy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
		"mappings": mappings,
	})
}

// RefreshClientInfo godoc
// @Summary Refresh Monobank client info
// @Description Fetches fresh account and mapping data directly from Monobank API.
// @Tags Monobank
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Refreshed accounts and mappings"
// @Failure 401 {object} map[string]string "Unauthorized or invalid token"
// @Failure 404 {object} map[string]string "Connection not found"
// @Failure 429 {object} map[string]string "Rate limit"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monobank/refresh [post]
func (c *MonobankController) RefreshClientInfo(ctx *gin.Context) {
	userID := ctx.GetString("userID")

	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User ID missing in context"})
		return
	}

	// Викликаємо метод сервісу, який робить реальний запит до Mono API
	accounts, mappings, err := c.service.RefreshClientInfo(userID)
	
	if err != nil {
		errStr := err.Error()

		// Обробка помилок
		if strings.Contains(errStr, "connection not found") || strings.Contains(errStr, "record not found") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Monobank connection not found"})
			return
		}

		// 🔥 Rate Limit (429)
		if strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "429") {
			ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "Monobank Rate Limit. Please wait 60s."})
			return
		}

		// Invalid Token (403/401)
		if strings.Contains(errStr, "403") || strings.Contains(errStr, "401") {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Invalid Monobank Token"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errStr})
		return
	}

	// Повертаємо свіжі дані
	ctx.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
		"mappings": mappings,
	})
}

type MonobankSettingsJSON struct {
	Accounts []models.BankAccountMapping `json:"accounts"`
}

// SaveSettings godoc
// @Summary Save Monobank settings
// @Description Saves the mapping between Monobank accounts and internal family accounts.
// @Tags Monobank
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body MonobankSettingsJSON true "Account mappings"
// @Success 200 {object} map[string]string "Success status"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monobank/settings [post]
func (ctrl *MonobankController) SaveSettings(c *gin.Context) {
	// Тут беремо об'єкт юзера, щоб отримати FamilyID
	user := c.MustGet("user").(*models.User)
	
	var input MonobankSettingsJSON

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.service.SaveSettings(user.ID, user.FamilyID, input.Accounts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ConfirmSync godoc
// @Summary Confirm and start synchronization
// @Description Saves settings (if provided) and starts a background synchronization with Monobank.
// @Tags Monobank
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body MonobankSettingsJSON false "Optional account mappings"
// @Success 200 {object} map[string]string "Started status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monobank/sync-confirm [post]
func (ctrl *MonobankController) ConfirmSync(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var input MonobankSettingsJSON

	// 1. Зберігаємо налаштування (якщо вони передані)
	if err := c.ShouldBindJSON(&input); err == nil && len(input.Accounts) > 0 {
		err := ctrl.service.SaveSettings(user.ID, user.FamilyID, input.Accounts)
		if err != nil {
			fmt.Printf("❌ Failed to save settings inside ConfirmSync: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save account settings: " + err.Error()})
			return
		}
	}

	// 2. Запускаємо синхронізацію у фоні (goroutine)
	go func() {
		_, err := ctrl.service.Sync(user.ID, "")
		if err != nil {
			fmt.Printf("❌ Background Sync Error for user %s: %v\n", user.Name, err)
		} else {
			fmt.Printf("✅ Background Sync Finished for user %s\n", user.Name)
		}
	}()

	// 3. Миттєва відповідь
	c.JSON(http.StatusOK, gin.H{
		"status":  "started",
		"message": "Синхронізацію розпочато у фоновому режимі",
	})
}

// ForceSync godoc
// @Summary Force manual synchronization
// @Description Manually triggers a synchronization for a specific account or all accounts of the user.
// @Tags Monobank
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param account_id query string false "Optional account ID to sync"
// @Success 200 {object} map[string]string "Sync started status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /monobank/sync [post]
func (ctrl *MonobankController) ForceSync(c *gin.Context) {
	userID := c.GetString("userID")
	accountID := c.Query("account_id")

	// Запускаємо асинхронно, щоб не блокувати UI
	go func() {
		ctrl.service.Sync(userID, accountID)
	}()

	c.JSON(http.StatusOK, gin.H{"status": "sync_started"})
}

// Disconnect godoc
// @Summary Disconnect Monobank
// @Description Removes the Monobank connection for the current user.
// @Tags Monobank
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "Disconnected status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monobank/disconnect [post]
func (ctrl *MonobankController) Disconnect(c *gin.Context) {
	userID := c.GetString("userID")

	if err := ctrl.service.Disconnect(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "disconnected"})
}

// GetStatus godoc
// @Summary Get Monobank sync status
// @Description Returns the current status of all background synchronizations.
// @Tags Monobank
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} services.SyncStatus
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /monobank/status [get]
func (ctrl *MonobankController) GetStatus(c *gin.Context) {
	status := ctrl.service.GetSyncStatus()
	c.JSON(http.StatusOK, status)
}


// ====================================================================================

	// ForceResyncCounterparties godoc
	// @Summary Global resync of counterparties
	// @Description Triggers a global resync of counterparties for all transactions across all families. Only for admins.
	// @Tags Monobank
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Success 200 {object} map[string]interface{} "Update message and count"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /monobank/force-resync [post]
	func (ctrl *MonobankController) ForceResyncCounterparties(c *gin.Context) {
		// Роут захищений мідлварею, тому просто так його не смикнуть,
		// але сама синхронізація тепер глобальна.
		
		updatedCount, err := ctrl.service.GlobalResyncCounterparties()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":                    "Базу успішно оновлено для ВСІХ сімей!",
			"updated_transactions_count": updatedCount,
		})
	}
