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

func TestAssetController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockAssetService)
	controller := NewAssetController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Next()
	})

	r.GET("/assets", controller.GetAll)
	r.POST("/assets", controller.Create)
	r.GET("/assets/:id", controller.GetOne)

	t.Run("GetAll", func(t *testing.T) {
		mockService.On("GetAll", mock.Anything).Return([]models.Asset{{Name: "Laptop"}}, nil).Once()

		w := PerformRequest(r, "GET", "/assets", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Create Success", func(t *testing.T) {
		mockService.On("Create", mock.Anything, mock.Anything).Return("a-1", nil).Once()

		body := map[string]string{"name": "Phone"}
		w := PerformRequest(r, "POST", "/assets", body)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("GetOne Success", func(t *testing.T) {
		mockService.On("GetByID", "a-1", mock.Anything).Return(&services.AssetWithStats{
			Asset: &models.Asset{Base: models.Base{ID: "a-1"}},
		}, nil).Once()

		w := PerformRequest(r, "GET", "/assets/a-1", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
