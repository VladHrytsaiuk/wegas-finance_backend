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

func TestCounterpartyController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockCounterpartyService)
	controller := NewCounterpartyController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Next()
	})

	r.POST("/counterparty-categories", controller.CreateCategory)
	r.GET("/counterparty-categories", controller.GetCategories)
	r.GET("/counterparty-categories/:id", controller.GetCategory)
	r.PUT("/counterparty-categories/:id", controller.UpdateCategory)

	r.POST("/counterparties", controller.Create)
	r.GET("/counterparties", controller.GetAll)
	r.GET("/counterparties/:id", controller.GetOne)
	r.PUT("/counterparties/:id", controller.Update)
	r.DELETE("/counterparties/:id", controller.Delete)

	t.Run("Create Category", func(t *testing.T) {
		input := CpCategoryJSON{Name: "Retail", Type: "expense"}
		mockService.On("CreateCategory", mock.Anything, mock.Anything).Return(&models.CounterpartyCategory{
			Base: models.Base{ID: "cat-1"},
			Name: "Retail",
		}, nil).Once()

		w := PerformRequest(r, "POST", "/counterparty-categories", input)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Get Categories", func(t *testing.T) {
		mockService.On("GetCategories", mock.Anything).Return([]models.CounterpartyCategory{}, nil).Once()
		w := PerformRequest(r, "GET", "/counterparty-categories", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create Counterparty", func(t *testing.T) {
		input := CounterpartyJSON{Name: "Store", Type: "expense"}
		mockService.On("Create", mock.Anything, mock.Anything).Return(&models.Counterparty{
			Base: models.Base{ID: "cp-1"},
			Name: "Store",
		}, nil).Once()

		w := PerformRequest(r, "POST", "/counterparties", input)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("GetAll Counterparties", func(t *testing.T) {
		mockService.On("GetAll", mock.Anything).Return([]models.Counterparty{}, nil).Once()
		w := PerformRequest(r, "GET", "/counterparties", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetOne Counterparty", func(t *testing.T) {
		mockService.On("GetByID", "cp-1", mock.Anything).Return(&models.Counterparty{
			Base: models.Base{ID: "cp-1"},
		}, nil).Once()
		w := PerformRequest(r, "GET", "/counterparties/cp-1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Update Counterparty", func(t *testing.T) {
		input := CounterpartyJSON{Name: "New Store", Type: "expense"}
		mockService.On("Update", "cp-1", mock.Anything, mock.Anything).Return(&models.Counterparty{
			Base: models.Base{ID: "cp-1"},
			Name: "New Store",
		}, nil).Once()

		w := PerformRequest(r, "PUT", "/counterparties/cp-1", input)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Delete Counterparty", func(t *testing.T) {
		mockService.On("Delete", "cp-1", mock.Anything).Return(nil).Once()
		w := PerformRequest(r, "DELETE", "/counterparties/cp-1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
