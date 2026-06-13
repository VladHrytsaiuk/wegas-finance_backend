package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionController_Errors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockTransactionService)
	controller := NewTransactionController(mockService)

	t.Run("GetOne - Not Found", func(t *testing.T) {
		mockService.On("GetByID", "tx-1", mock.Anything).Return(nil, assert.AnError).Once()
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetupTestUser(c, "u-1", "f-1")
		c.Params = gin.Params{{Key: "id", Value: "tx-1"}}
		
		controller.GetOne(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Delete - Error", func(t *testing.T) {
		mockService.On("Delete", "tx-1", mock.Anything).Return(assert.AnError).Once()
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetupTestUser(c, "u-1", "f-1")
		c.Params = gin.Params{{Key: "id", Value: "tx-1"}}
		
		controller.Delete(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Predict - Missing Name", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetupTestUser(c, "u-1", "f-1")
		c.Request, _ = http.NewRequest("GET", "/api/transactions/predict", nil)
		
		controller.PredictCategory(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
