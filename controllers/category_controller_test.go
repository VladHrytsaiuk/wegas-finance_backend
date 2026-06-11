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

func TestCategoryController_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockCategoryService)
	controller := NewCategoryController(mockService)

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		input := services.CategoryInput{Name: "Food", Type: "expense"}
		user := &models.User{Base: models.Base{ID: "user1"}, RoleID: "admin"}
		
		r.POST("/categories", func(c *gin.Context) {
			c.Set("user", user)
			controller.Create(c)
		})

		mockService.On("Create", input, user).Return(&models.Category{Name: "Food"}, nil)

		req := httptest.NewRequest("POST", "/categories", PerformBody(input))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp models.Category
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Food", resp.Name)
		mockService.AssertExpectations(t)
	})

	t.Run("Forbidden", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		input := services.CategoryInput{Name: "Food"}
		user := &models.User{Base: models.Base{ID: "user1"}, RoleID: "child"}

		r.POST("/categories", func(c *gin.Context) {
			c.Set("user", user)
			controller.Create(c)
		})

		mockService.On("Create", input, user).Return(nil, services.ErrAccessDenied)

		req := httptest.NewRequest("POST", "/categories", PerformBody(input))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestCategoryController_GetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockCategoryService)
	controller := NewCategoryController(mockService)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	user := &models.User{Base: models.Base{ID: "user1"}}
	categories := []models.Category{{Name: "Food"}, {Name: "Rent"}}

	r.GET("/categories", func(c *gin.Context) {
		c.Set("user", user)
		controller.GetAll(c)
	})

	mockService.On("GetAll", user).Return(categories, nil)

	req := httptest.NewRequest("GET", "/categories", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []models.Category
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 2, len(resp))
}

// Helper to encode body
func PerformBody(body interface{}) *bytes.Buffer {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)
	return &buf
}
