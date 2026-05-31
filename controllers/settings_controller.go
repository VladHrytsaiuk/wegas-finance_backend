	package controllers

	import (
		"net/http"

		"github.com/VladHrytsaiuk/wegas-finance/backend/models"
		"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
		"github.com/gin-gonic/gin"
	)

	type SettingsController struct {
		userRepo repositories.UserRepository
	}

	func NewSettingsController(userRepo repositories.UserRepository) *SettingsController {
		return &SettingsController{userRepo: userRepo}
	}

	type UpdateSettingsRequest struct {
		BaseCurrency string `json:"base_currency"`
		Language     string `json:"language"`
		Theme        string `json:"theme"`
	}

	// GetSettingsHTTP godoc
	// @Summary Get user settings
	// @Description Returns the settings for the current user.
	// @Tags Settings
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Success 200 {object} map[string]string "User settings"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Router /settings [get]
	func (c *SettingsController) GetSettingsHTTP(ctx *gin.Context) {
		// 🔥 ОПТИМІЗАЦІЯ: Беремо готового юзера з контексту
		// (AuthMiddleware вже зробив запит в БД за нас)
		currentUser := ctx.MustGet("user").(*models.User)

		ctx.JSON(http.StatusOK, gin.H{
			"base_currency": currentUser.BaseCurrency,
			"language":      currentUser.Language,
			"theme":         currentUser.Theme,
		})
	}

	// SaveSettingsHTTP godoc
	// @Summary Save user settings
	// @Description Updates the settings for the current user.
	// @Tags Settings
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param body body UpdateSettingsRequest true "Updated settings"
	// @Success 200 {object} map[string]string "Success status"
	// @Failure 400 {object} map[string]string "Invalid input"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /settings [post]
	func (c *SettingsController) SaveSettingsHTTP(ctx *gin.Context) {
		currentUser := ctx.MustGet("user").(*models.User)

		var req UpdateSettingsRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Оновлюємо поля (дитина може міняти собі тему/мову — це ок)
		if req.BaseCurrency != "" {
			currentUser.BaseCurrency = req.BaseCurrency
		}
		if req.Language != "" {
			currentUser.Language = req.Language
		}
		if req.Theme != "" {
			currentUser.Theme = req.Theme
		}

		// Зберігаємо зміни
		if err := c.userRepo.Update(currentUser); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "saved"})
	}