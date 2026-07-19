package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type TelegramWebhookController struct {
	service services.TelegramWebhookService
}

type TelegramWebhookDeleteJSON struct {
	DropPendingUpdates bool `json:"drop_pending_updates"`
}

func NewTelegramWebhookController(service services.TelegramWebhookService) *TelegramWebhookController {
	return &TelegramWebhookController{service: service}
}

func (h *TelegramWebhookController) GetStatus(c *gin.Context) {
	status, err := h.service.GetWebhookStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *TelegramWebhookController) SyncWebhook(c *gin.Context) {
	status, err := h.service.SyncWebhook()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *TelegramWebhookController) DeleteWebhook(c *gin.Context) {
	var input TelegramWebhookDeleteJSON
	if err := c.ShouldBindJSON(&input); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.DeleteWebhook(input.DropPendingUpdates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
