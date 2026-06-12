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

func TestTagController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockTagService)
	controller := NewTagController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Next()
	})

	r.GET("/tags", controller.GetAll)
	r.POST("/tags", controller.Create)
	r.DELETE("/tags/:id", controller.Delete)

	t.Run("GetAll Tags Success", func(t *testing.T) {
		mockService.On("GetAll", "f-1").Return([]models.Tag{}, nil).Once()
		w := PerformRequest(r, "GET", "/tags", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create Tag Success", func(t *testing.T) {
		input := TagInputJSON{Name: "Gift"}
		mockService.On("Create", "Gift", "", mock.Anything).Return(&models.Tag{
			Base: models.Base{ID: "t-1"},
			Name: "Gift",
		}, nil).Once()

		w := PerformRequest(r, "POST", "/tags", input)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Delete Tag Success", func(t *testing.T) {
		mockService.On("Delete", "t-1", mock.Anything).Return(nil).Once()
		w := PerformRequest(r, "DELETE", "/tags/t-1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
