package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCategoryController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Create Category", func(t *testing.T) {
		mockService := new(services.MockCategoryService)
		controller := NewCategoryController(mockService)
		
		user := &models.User{Base: models.Base{ID: "u-1"}, FamilyID: "f-1"}
		input := services.CategoryInput{Name: "Food", Type: "expense"}

		mockService.On("Create", input, user).Return(&models.Category{
			Base: models.Base{ID: "cat-1"},
			Name: "Food",
		}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		
		body, _ := json.Marshal(input)
		c.Request, _ = http.NewRequest("POST", "/api/categories", bytes.NewReader(body))
		
		controller.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("GetAll Categories", func(t *testing.T) {
		mockService := new(services.MockCategoryService)
		controller := NewCategoryController(mockService)
		
		user := &models.User{Base: models.Base{ID: "u-1"}}
		mockService.On("GetAll", user).Return([]models.Category{
			{Name: "Food"}, {Name: "Rent"},
		}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("GET", "/api/categories", nil)
		
		controller.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
