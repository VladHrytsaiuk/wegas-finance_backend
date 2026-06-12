package controllers

import (
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSettingsController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(services.MockUserRepository)
	controller := NewSettingsController(mockRepo)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Next()
	})

	r.GET("/settings", controller.GetSettingsHTTP)
	r.POST("/settings", controller.SaveSettingsHTTP)

	t.Run("Get Settings Success", func(t *testing.T) {
		w := PerformRequest(r, "GET", "/settings", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Save Settings Success", func(t *testing.T) {
		input := UpdateSettingsRequest{BaseCurrency: "USD", Language: "en", Theme: "dark"}
		mockRepo.On("Update", mock.Anything).Return(nil).Once()

		w := PerformRequest(r, "POST", "/settings", input)
		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
