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
	Name           string `json:"name" binding:"required"`
	Type           string `json:"type" binding:"required"`
	Currency       string `json:"currency" binding:"required"`
	InitialBalance int64  `json:"initial_balance"`
	Color          string `json:"color"`
	CardNumber     string `json:"card_number"`
	PaymentSystem  string `json:"payment_system"`
	BankName       string `json:"bank_name"` 
	CardType       string `json:"card_type"` // Ви просили CardType
	OwnerID        string `json:"user_id"`
	StorageTypeID *string `json:"storage_type_id"`
	GoalID        *string `json:"goal_id"`

}

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
		PaymentSystem:  jsonInput.PaymentSystem,
		OwnerID:        jsonInput.OwnerID,
		StorageTypeID: jsonInput.StorageTypeID,
		GoalID:        jsonInput.GoalID,
	}

	// Передаємо юзера в сервіс
	account, err := h.service.Create(serviceInput, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}

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
		PaymentSystem:  jsonInput.PaymentSystem,
		OwnerID:        jsonInput.OwnerID,
		StorageTypeID: jsonInput.StorageTypeID,
		GoalID:        jsonInput.GoalID,
	}

	account, err := h.service.Update(id, serviceInput, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

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