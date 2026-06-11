package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionController_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockTransactionService)
	controller := NewTransactionController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "user-123", "family-456")
		c.Next()
	})
	r.POST("/transactions", controller.Create)

	t.Run("Success JSON", func(t *testing.T) {
		inputJSON := CreateTxJSON{
			AccountID: "acc-1",
			Amount:    1000,
			Date:      123456789,
			Type:      "expense",
		}

		mockService.On("Create", mock.MatchedBy(func(in services.CreateTransactionInput) bool {
			return in.AccountID == "acc-1" && in.Amount == 1000 && in.Type == "expense"
		}), mock.Anything, mock.Anything).Return("tx-1", nil).Once()

		w := PerformRequest(r, "POST", "/transactions", inputJSON)

		assert.Equal(t, http.StatusCreated, w.Code)
		var res map[string]string
		json.Unmarshal(w.Body.Bytes(), &res)
		assert.Equal(t, "success", res["status"])
		assert.Equal(t, "tx-1", res["id"])
		mockService.AssertExpectations(t)
	})

	t.Run("Bind Error", func(t *testing.T) {
		body := map[string]interface{}{
			"amount": "invalid", // should be int64
		}
		w := PerformRequest(r, "POST", "/transactions", body)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
