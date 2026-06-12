package controllers

import (
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWishlistController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockWishlistService)
	controller := NewWishlistController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Next()
	})

	r.GET("/wishlist", controller.GetAll)
	r.POST("/wishlist", controller.Create)
	r.POST("/wishlist/:id/reserve", controller.ToggleReservation)

	t.Run("GetAll", func(t *testing.T) {
		mockService.On("GetItems", "f-1", "u-1", "", "").Return([]models.WishlistItem{{Name: "Toy"}}, nil).Once()

		w := PerformRequest(r, "GET", "/wishlist", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Create Success", func(t *testing.T) {
		req := models.CreateWishlistRequest{Name: "Toy"}
		mockService.On("CreateItem", req, "u-1", "f-1").Return(&models.WishlistItem{Name: "Toy"}, nil).Once()

		w := PerformRequest(r, "POST", "/wishlist", req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ToggleReservation Success", func(t *testing.T) {
		mockService.On("ToggleReservation", "w-1", "u-1").Return(nil).Once()

		w := PerformRequest(r, "POST", "/wishlist/w-1/reserve", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
