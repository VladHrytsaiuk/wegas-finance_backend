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

func TestStorageTypeController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockStorageTypeService)
	controller := NewStorageTypeController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Next()
	})

	r.GET("/storage-types", controller.GetAll)
	r.POST("/storage-types", controller.Create)
	r.DELETE("/storage-types/:id", controller.Delete)

	t.Run("GetAll StorageTypes Success", func(t *testing.T) {
		mockService.On("GetAll", "f-1").Return([]models.StorageType{}, nil).Once()
		w := PerformRequest(r, "GET", "/storage-types", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create StorageType Success", func(t *testing.T) {
		input := models.StorageType{Name: "Safe"}
		mockService.On("Create", mock.Anything).Return(nil).Once()

		w := PerformRequest(r, "POST", "/storage-types", input)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Delete StorageType Success", func(t *testing.T) {
		mockService.On("Delete", "st-1").Return(nil).Once()
		w := PerformRequest(r, "DELETE", "/storage-types/st-1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
