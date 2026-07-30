package controllers

import (
	"log"
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/pkg/telegram"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type TelegramBotController struct {
	service       services.TelegramBotService
	webhookSecret string
}

func NewTelegramBotController(service services.TelegramBotService, webhookSecret string) *TelegramBotController {
	return &TelegramBotController{service: service, webhookSecret: webhookSecret}
}

func (h *TelegramBotController) Webhook(c *gin.Context) {
	if h.webhookSecret != "" {
		if c.GetHeader("X-Telegram-Bot-Api-Secret-Token") != h.webhookSecret {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
	}

	var update telegram.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf(
		"telegram webhook update: id=%d message=%t edited_message=%t channel_post=%t edited_channel_post=%t callback=%t",
		update.UpdateID,
		update.Message != nil,
		update.EditedMessage != nil,
		update.ChannelPost != nil,
		update.EditedChannelPost != nil,
		update.CallbackQuery != nil,
	)

	if err := h.service.HandleUpdate(&update); err != nil {
		log.Printf("telegram webhook processing error: %v", err)
		c.JSON(http.StatusOK, gin.H{"ok": true, "handled": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
