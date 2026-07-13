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

func TestAccountController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockAccountService)
	controller := NewAccountController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "user-1", "family-1")
		c.Next()
	})

	r.POST("/accounts", controller.Create)
	r.GET("/accounts", controller.GetAll)
	r.GET("/accounts/:id", controller.GetOne)
	r.PUT("/accounts/:id", controller.Update)
	r.PUT("/accounts/mobile-order", controller.UpdateMobileOrder)
	r.DELETE("/accounts/:id", controller.Delete)

	t.Run("Create Account", func(t *testing.T) {
		input := AccountInputJSON{
			Name:     "Test Account",
			Type:     "cash",
			Currency: "UAH",
		}

		mockService.On("Create", mock.Anything, mock.Anything).Return(&models.Account{
			Base: models.Base{ID: "acc-1"},
			Name: "Test Account",
		}, nil).Once()

		w := PerformRequest(r, "POST", "/accounts", input)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("GetAll Accounts", func(t *testing.T) {
		mockService.On("GetAll", mock.Anything).Return([]models.Account{
			{Base: models.Base{ID: "acc-1"}, Name: "Acc 1"},
		}, nil).Once()

		w := PerformRequest(r, "GET", "/accounts", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("GetOne Account", func(t *testing.T) {
		mockService.On("GetByID", "acc-1", mock.Anything).Return(&models.Account{
			Base: models.Base{ID: "acc-1"},
			Name: "Acc 1",
		}, nil).Once()

		w := PerformRequest(r, "GET", "/accounts/acc-1", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Update Account", func(t *testing.T) {
		input := AccountInputJSON{
			Name:     "Updated Account",
			Type:     "bank",
			Currency: "USD",
		}

		mockService.On("Update", "acc-1", mock.Anything, mock.Anything).Return(&models.Account{
			Base: models.Base{ID: "acc-1"},
			Name: "Updated Account",
		}, nil).Once()

		w := PerformRequest(r, "PUT", "/accounts/acc-1", input)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Delete Account", func(t *testing.T) {
		mockService.On("Delete", "acc-1", mock.Anything).Return(nil).Once()

		w := PerformRequest(r, "DELETE", "/accounts/acc-1", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Update Mobile Order", func(t *testing.T) {
		mockService.On("UpdateMobileOrder", []string{"acc-2", "acc-1"}, mock.Anything).Return(nil).Once()

		w := PerformRequest(r, "PUT", "/accounts/mobile-order", gin.H{
			"account_ids": []string{"acc-2", "acc-1"},
		})

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
