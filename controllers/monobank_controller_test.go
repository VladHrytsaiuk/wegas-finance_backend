package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMonobankController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockMonobankService)
	controller := NewMonobankController(mockService)

	r := gin.Default()
	
	// Public route
	r.POST("/monobank/webhook", controller.Webhook)
	
	// Protected route
	protected := r.Group("/")
	protected.Use(func(c *gin.Context) {
		SetupTestUser(c, "user-123", "family-456")
		c.Next()
	})
	protected.GET("/monobank/status", controller.GetStatus)

	t.Run("Webhook Success", func(t *testing.T) {
		payload := services.MonoWebhookPayload{
			Type: "StatementItem",
		}
		mockService.On("ProcessWebhook", payload).Return(nil).Once()

		w := PerformRequest(r, "POST", "/monobank/webhook", payload)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
		mockService.AssertExpectations(t)
	})

	t.Run("GetStatus Success", func(t *testing.T) {
		expectedStatus := services.SyncStatus{
			IsRunning:     true,
			Message:       "Syncing...",
			TotalImported: 5,
		}
		mockService.On("GetSyncStatus", "user-123").Return(expectedStatus).Once()

		w := PerformRequest(r, "GET", "/monobank/status", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var res services.SyncStatus
		json.Unmarshal(w.Body.Bytes(), &res)
		assert.Equal(t, expectedStatus.IsRunning, res.IsRunning)
		assert.Equal(t, expectedStatus.Message, res.Message)
		assert.Equal(t, expectedStatus.TotalImported, res.TotalImported)
		mockService.AssertExpectations(t)
	})
}
