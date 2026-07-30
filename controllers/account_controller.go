package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type AccountController struct {
	service services.AccountService
}

func NewAccountController(service services.AccountService) *AccountController {
	return &AccountController{service: service}
}

// --- HANDLERS ---

type AccountInputJSON struct {
	Name           string   `json:"name" binding:"required"`
	Type           string   `json:"type" binding:"required"`
	Currency       string   `json:"currency" binding:"required"`
	InitialBalance int64    `json:"initial_balance"`
	Color          string   `json:"color"`
	CardNumber     string   `json:"card_number"`
	CardNumbers    []string `json:"card_numbers"`
	PaymentSystem  string   `json:"payment_system"`
	BankName       string   `json:"bank_name"`
	CardType       string   `json:"card_type"` // Ви просили CardType
	OwnerID        string   `json:"user_id"`
	StorageTypeID  *string  `json:"storage_type_id"`
	GoalID         *string  `json:"goal_id"`
}

type UpdateMobileAccountOrderJSON struct {
	AccountIDs []string `json:"account_ids" binding:"required"`
}

// Create godoc
// @Summary Create a new account
// @Description Creates a new financial account for the current user.
// @Tags Accounts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body AccountInputJSON true "Account details"
// @Success 201 {object} models.Account
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /accounts [post]
func (h *AccountController) Create(c *gin.Context) {
	// 1. Отримуємо об'єкт юзера з контексту
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var jsonInput AccountInputJSON
	if err := c.ShouldBindJSON(&jsonInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviceInput := services.CreateAccountInput{
		Name:           jsonInput.Name,
		Type:           jsonInput.Type,
		Currency:       jsonInput.Currency,
		InitialBalance: jsonInput.InitialBalance,
		Color:          jsonInput.Color,
		BankName:       jsonInput.BankName,
		CardType:       jsonInput.CardType,
		CardNumber:     jsonInput.CardNumber,
		CardNumbers:    jsonInput.CardNumbers,
		PaymentSystem:  jsonInput.PaymentSystem,
		OwnerID:        jsonInput.OwnerID,
		StorageTypeID:  jsonInput.StorageTypeID,
		GoalID:         jsonInput.GoalID,
	}

	// Передаємо юзера в сервіс
	account, err := h.service.Create(serviceInput, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}

// GetAll godoc
// @Summary Get all accounts
// @Description Returns a list of all accounts accessible to the current user.
// @Tags Accounts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.Account
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /accounts [get]
func (h *AccountController) GetAll(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	// Сервіс сам вирішить, що показати цьому юзеру
	accounts, err := h.service.GetAll(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// GetOne godoc
// @Summary Get account by ID
// @Description Returns a single account by its ID.
// @Tags Accounts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Account ID"
// @Success 200 {object} models.Account
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Account not found"
// @Router /accounts/{id} [get]
func (h *AccountController) GetOne(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	account, err := h.service.GetByID(id, user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, account)
}

// Update godoc
// @Summary Update account
// @Description Updates an existing account by its ID.
// @Tags Accounts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Account ID"
// @Param body body AccountInputJSON true "Updated account details"
// @Success 200 {object} models.Account
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /accounts/{id} [put]
func (h *AccountController) Update(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var jsonInput AccountInputJSON
	if err := c.ShouldBindJSON(&jsonInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviceInput := services.CreateAccountInput{
		Name:           jsonInput.Name,
		Type:           jsonInput.Type,
		Currency:       jsonInput.Currency,
		InitialBalance: jsonInput.InitialBalance,
		Color:          jsonInput.Color,
		BankName:       jsonInput.BankName,
		CardType:       jsonInput.CardType,
		CardNumber:     jsonInput.CardNumber,
		CardNumbers:    jsonInput.CardNumbers,
		PaymentSystem:  jsonInput.PaymentSystem,
		OwnerID:        jsonInput.OwnerID,
		StorageTypeID:  jsonInput.StorageTypeID,
		GoalID:         jsonInput.GoalID,
	}

	account, err := h.service.Update(id, serviceInput, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

// Delete godoc
// @Summary Delete account
// @Description Deletes an account by its ID.
// @Tags Accounts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Account ID"
// @Success 200 {object} map[string]string "Account deleted"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /accounts/{id} [delete]
func (h *AccountController) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	if err := h.service.Delete(id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted"})
}

// UpdateMobileOrder godoc
// @Summary Update mobile accounts order
// @Description Saves the current user's custom mobile account order.
// @Tags Accounts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body UpdateMobileAccountOrderJSON true "Ordered account ids"
// @Success 200 {object} map[string]string "Order saved"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /accounts/mobile-order [put]
func (h *AccountController) UpdateMobileOrder(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var jsonInput UpdateMobileAccountOrderJSON
	if err := c.ShouldBindJSON(&jsonInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateMobileOrder(jsonInput.AccountIDs, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "saved"})
}
