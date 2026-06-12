package controllers

import (
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCurrencyController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockCurrencyService)
	controller := NewCurrencyController(mockService)

	r := gin.Default()
	r.GET("/currencies", controller.GetRates)

	t.Run("GetRates Success", func(t *testing.T) {
		mockService.On("GetAllRates").Return([]models.ExchangeRate{
			{CurrencyCode: "USD", Rate: 40.5},
		}, nil).Once()

		w := PerformRequest(r, "GET", "/currencies", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
