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

func TestCategoryController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockCategoryService)
	controller := NewCategoryController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Next()
	})

	r.POST("/categories", controller.Create)
	r.GET("/categories", controller.GetAll)
	r.GET("/categories/:id", controller.GetOne)
	r.PUT("/categories/:id", controller.Update)
	r.DELETE("/categories/:id", controller.Delete)

	t.Run("Create Category", func(t *testing.T) {
		input := services.CategoryInput{Name: "Food", Type: "expense"}

		mockService.On("Create", input, mock.Anything).Return(&models.Category{
			Base: models.Base{ID: "cat-1"},
			Name: "Food",
		}, nil).Once()

		w := PerformRequest(r, "POST", "/categories", input)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("GetAll Categories", func(t *testing.T) {
		mockService.On("GetAll", mock.Anything).Return([]models.Category{
			{Name: "Food"}, {Name: "Rent"},
		}, nil).Once()

		w := PerformRequest(r, "GET", "/categories", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("GetOne Category", func(t *testing.T) {
		mockService.On("GetByID", "cat-1", mock.Anything).Return(&models.Category{
			Base: models.Base{ID: "cat-1"},
			Name: "Food",
		}, nil).Once()

		w := PerformRequest(r, "GET", "/categories/cat-1", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Update Category", func(t *testing.T) {
		input := services.CategoryInput{Name: "Groceries", Type: "expense"}

		mockService.On("Update", "cat-1", input, mock.Anything).Return(&models.Category{
			Base: models.Base{ID: "cat-1"},
			Name: "Groceries",
		}, nil).Once()

		w := PerformRequest(r, "PUT", "/categories/cat-1", input)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Delete Category", func(t *testing.T) {
		mockService.On("Delete", "cat-1", mock.Anything).Return(nil).Once()

		w := PerformRequest(r, "DELETE", "/categories/cat-1", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
