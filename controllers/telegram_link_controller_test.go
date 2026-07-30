package controllers

import (
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTelegramLinkController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockTelegramLinkService)
	controller := NewTelegramLinkController(mockService)

	protected := gin.Default()
	protected.Use(func(c *gin.Context) {
		SetupTestUser(c, "user-123", "family-456")
		c.Next()
	})
	protected.GET("/integrations/telegram", controller.GetStatus)
	protected.POST("/integrations/telegram/link-token", controller.CreateLinkToken)
	protected.DELETE("/integrations/telegram/link", controller.RevokeLink)

	public := gin.Default()
	public.POST("/telegram/link/complete", controller.CompleteLink)

	t.Run("Get Status", func(t *testing.T) {
		mockService.On("GetStatus", mock.Anything).Return(&services.TelegramLinkStatus{
			IsLinked:    true,
			BotUsername: "wegas_finance_bot",
		}, nil).Once()

		w := PerformRequest(protected, "GET", "/integrations/telegram", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create Link Token", func(t *testing.T) {
		mockService.On("CreateLinkToken", mock.Anything).Return(&services.TelegramLinkTokenResponse{
			Token:       "token-1",
			DeepLink:    "https://t.me/wegas_finance_bot?start=token-1",
			BotUsername: "wegas_finance_bot",
		}, nil).Once()

		w := PerformRequest(protected, "POST", "/integrations/telegram/link-token", nil)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Revoke Link", func(t *testing.T) {
		mockService.On("RevokeLink", mock.Anything).Return(nil).Once()

		w := PerformRequest(protected, "DELETE", "/integrations/telegram/link", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Complete Link", func(t *testing.T) {
		mockService.On("CompleteLink", services.TelegramLinkCompleteInput{
			Token:          "token-1",
			TelegramUserID: 12345,
			TelegramChatID: 67890,
			Username:       "vlad",
			FirstName:      "Vlad",
		}).Return(&services.TelegramLinkStatus{
			IsLinked:    true,
			BotUsername: "wegas_finance_bot",
		}, nil).Once()

		w := PerformRequest(public, "POST", "/telegram/link/complete", TelegramLinkCompleteJSON{
			Token:          "token-1",
			TelegramUserID: 12345,
			TelegramChatID: 67890,
			Username:       "vlad",
			FirstName:      "Vlad",
		})
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
