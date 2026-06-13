package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccountController_Errors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockAccountService)
	controller := NewAccountController(mockService)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "user-1", "family-1")
		c.Next()
	})
	r.POST("/accounts", controller.Create)
	r.PUT("/accounts/:id", controller.Update)
	r.GET("/accounts/:id", controller.GetOne)

	t.Run("Create - Bind Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetupTestUser(c, "user-1", "family-1")
		c.Request, _ = http.NewRequest("POST", "/accounts", bytes.NewBuffer([]byte("invalid json")))
		
		controller.Create(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Update - Not Found", func(t *testing.T) {
		mockService.On("Update", "wrong-id", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()
		
		body, _ := json.Marshal(AccountInputJSON{Name: "Test", Type: "cash", Currency: "UAH"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetupTestUser(c, "user-1", "family-1")
		c.Params = gin.Params{{Key: "id", Value: "wrong-id"}}
		c.Request, _ = http.NewRequest("PUT", "/accounts/wrong-id", bytes.NewBuffer(body))
		
		controller.Update(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("GetOne - Not Found", func(t *testing.T) {
		mockService.On("GetByID", "wrong-id", mock.Anything).Return(nil, assert.AnError).Once()
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetupTestUser(c, "user-1", "family-1")
		c.Params = gin.Params{{Key: "id", Value: "wrong-id"}}
		
		controller.GetOne(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
