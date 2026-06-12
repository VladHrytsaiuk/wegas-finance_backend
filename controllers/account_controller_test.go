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
	"github.com/stretchr/testify/mock"
)

func TestAccountController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Create Account", func(t *testing.T) {
		mockService := new(services.MockAccountService)
		controller := NewAccountController(mockService)
		
		user := &models.User{Base: models.Base{ID: "user-1"}, FamilyID: "family-1"}
		input := AccountInputJSON{
			Name:     "Test Account",
			Type:     "cash",
			Currency: "UAH",
		}

		mockService.On("Create", mock.Anything, user).Return(&models.Account{
			Base: models.Base{ID: "acc-1"},
			Name: "Test Account",
		}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		
		body, _ := json.Marshal(input)
		c.Request, _ = http.NewRequest("POST", "/accounts", bytes.NewReader(body))
		
		controller.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.Account
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "acc-1", response.ID)
		mockService.AssertExpectations(t)
	})

	t.Run("GetAll Accounts", func(t *testing.T) {
		mockService := new(services.MockAccountService)
		controller := NewAccountController(mockService)
		
		user := &models.User{Base: models.Base{ID: "user-1"}, FamilyID: "family-1"}
		mockService.On("GetAll", user).Return([]models.Account{
			{Base: models.Base{ID: "acc-1"}, Name: "Acc 1"},
			{Base: models.Base{ID: "acc-2"}, Name: "Acc 2"},
		}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		
		controller.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []models.Account
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		mockService.AssertExpectations(t)
	})

	t.Run("GetOne Account", func(t *testing.T) {
		mockService := new(services.MockAccountService)
		controller := NewAccountController(mockService)
		
		user := &models.User{Base: models.Base{ID: "user-1"}, FamilyID: "family-1"}
		mockService.On("GetByID", "acc-1", user).Return(&models.Account{
			Base: models.Base{ID: "acc-1"},
			Name: "Acc 1",
		}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "acc-1"}}
		c.Set("user", user)
		
		controller.GetOne(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.Account
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "acc-1", response.ID)
		mockService.AssertExpectations(t)
	})
}
