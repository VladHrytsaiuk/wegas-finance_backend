package controllers

import (
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTelegramWebhookController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockTelegramWebhookService)
	controller := NewTelegramWebhookController(mockService)

	r := gin.Default()
	r.GET("/integrations/telegram/webhook", controller.GetStatus)
	r.POST("/integrations/telegram/webhook/sync", controller.SyncWebhook)
	r.DELETE("/integrations/telegram/webhook", controller.DeleteWebhook)

	t.Run("Get Status", func(t *testing.T) {
		mockService.On("GetWebhookStatus").Return(&services.TelegramWebhookStatus{
			Configured: true,
			WebhookURL: "https://app.example.com/api/telegram/webhook",
		}, nil).Once()

		w := PerformRequest(r, "GET", "/integrations/telegram/webhook", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Sync Webhook", func(t *testing.T) {
		mockService.On("SyncWebhook").Return(&services.TelegramWebhookStatus{
			Configured: true,
			WebhookURL: "https://app.example.com/api/telegram/webhook",
		}, nil).Once()

		w := PerformRequest(r, "POST", "/integrations/telegram/webhook/sync", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Delete Webhook", func(t *testing.T) {
		mockService.On("DeleteWebhook", true).Return(nil).Once()

		w := PerformRequest(r, "DELETE", "/integrations/telegram/webhook", TelegramWebhookDeleteJSON{
			DropPendingUpdates: true,
		})
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
