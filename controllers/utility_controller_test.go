package controllers

import (
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUtilityController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockUtilityService)
	controller := NewUtilityController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Next()
	})

	r.GET("/utility/meters", controller.GetMeters)
	r.POST("/utility/meters", controller.CreateMeter)
	r.GET("/utility/readings", controller.GetReadings)
	r.POST("/utility/readings", controller.CreateReading)
	r.POST("/utility/readings/:id/pay", controller.PayReading)

	t.Run("GetMeters Success", func(t *testing.T) {
		mockService.On("GetMeters", mock.Anything).Return([]models.UtilityMeter{}, nil).Once()
		w := PerformRequest(r, "GET", "/utility/meters", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CreateMeter Success", func(t *testing.T) {
		input := models.UtilityMeter{Name: "Electricity"}
		mockService.On("CreateMeter", mock.Anything, mock.Anything).Return(nil).Once()
		w := PerformRequest(r, "POST", "/utility/meters", input)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("GetReadings Success", func(t *testing.T) {
		mockService.On("GetReadings", mock.Anything, "").Return([]models.UtilityReading{}, nil).Once()
		w := PerformRequest(r, "GET", "/utility/readings", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CreateReading Success", func(t *testing.T) {
		input := models.UtilityReading{Value: 100}
		mockService.On("CreateReading", mock.Anything, mock.Anything).Return(nil).Once()
		w := PerformRequest(r, "POST", "/utility/readings", input)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("PayReading Success", func(t *testing.T) {
		input := PayReadingJSON{AccountID: "acc-1"}
		mockService.On("PayReading", "r-1", "acc-1", mock.Anything).Return(nil).Once()
		w := PerformRequest(r, "POST", "/utility/readings/r-1/pay", input)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
