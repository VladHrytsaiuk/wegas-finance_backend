package controllers

import (
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestShoppingController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockShoppingService)
	controller := NewShoppingController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Next()
	})

	r.GET("/shopping-lists", controller.GetLists)
	r.POST("/shopping-lists", controller.CreateList)
	r.POST("/shopping-lists/:id/items", controller.AddItem)

	t.Run("GetLists", func(t *testing.T) {
		mockService.On("GetLists", "f-1", "u-1").Return([]models.ShoppingList{{Title: "L1"}}, nil).Once()

		w := PerformRequest(r, "GET", "/shopping-lists", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("CreateList Success", func(t *testing.T) {
		req := models.CreateShoppingListRequest{Title: "New List"}
		mockService.On("CreateList", req, "u-1", "f-1").Return(&models.ShoppingList{Title: "New List"}, nil).Once()

		w := PerformRequest(r, "POST", "/shopping-lists", req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("AddItem Success", func(t *testing.T) {
		req := models.CreateShoppingItemRequest{Name: "Milk"}
		mockService.On("AddItemToList", "l-1", req).Return(&models.ShoppingItem{Name: "Milk"}, nil).Once()

		w := PerformRequest(r, "POST", "/shopping-lists/l-1/items", req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})
}
