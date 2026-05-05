package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service services.AuthService
}

func NewAuthController(service services.AuthService) *AuthController {
	return &AuthController{service: service}
}

type RegisterJSON struct {
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"invite_code" binding:"required"` // 🔥
}

type LoginJSON struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthController) Register(c *gin.Context) {
	var json RegisterJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.service.Register(services.RegisterInput{
		Name:       json.Name,
		Email:      json.Email,
		Password:   json.Password,
		InviteCode: json.InviteCode, // 🔥
	})

	if err != nil {
		// Якщо помилка коду - 403 Forbidden
		if err.Error() == "invalid invite code" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Невірний код запрошення"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *AuthController) Login(c *gin.Context) {
	var json LoginJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.service.Login(services.LoginInput{
		Email:    json.Email,
		Password: json.Password,
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Невірний логін або пароль"})
		return
	}

	c.JSON(http.StatusOK, res)
}