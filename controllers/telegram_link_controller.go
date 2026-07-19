package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type TelegramLinkController struct {
	service services.TelegramLinkService
}

type TelegramLinkCompleteJSON struct {
	Token          string `json:"token" binding:"required"`
	TelegramUserID int64  `json:"telegram_user_id" binding:"required"`
	TelegramChatID int64  `json:"telegram_chat_id" binding:"required"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
}

func NewTelegramLinkController(service services.TelegramLinkService) *TelegramLinkController {
	return &TelegramLinkController{service: service}
}

func (h *TelegramLinkController) GetStatus(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	status, err := h.service.GetStatus(currentUser.(*models.User))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *TelegramLinkController) CreateLinkToken(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	response, err := h.service.CreateLinkToken(currentUser.(*models.User))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *TelegramLinkController) RevokeLink(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.service.RevokeLink(currentUser.(*models.User)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "revoked"})
}

func (h *TelegramLinkController) CompleteLink(c *gin.Context) {
	var input TelegramLinkCompleteJSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, err := h.service.CompleteLink(services.TelegramLinkCompleteInput{
		Token:          input.Token,
		TelegramUserID: input.TelegramUserID,
		TelegramChatID: input.TelegramChatID,
		Username:       input.Username,
		FirstName:      input.FirstName,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}
