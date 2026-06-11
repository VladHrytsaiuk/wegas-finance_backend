package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthController_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockAuthService)
	controller := NewAuthController(mockService)

	r := gin.Default()
	r.POST("/login", controller.Login)

	t.Run("Success", func(t *testing.T) {
		loginInput := services.LoginInput{
			Email:    "test@example.com",
			Password: "password123",
		}
		expectedRes := &services.LoginResponse{
			Token: "test-token",
			User:  models.User{Name: "Test User"},
		}

		mockService.On("Login", loginInput).Return(expectedRes, nil).Once()

		body := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		w := PerformRequest(r, "POST", "/login", body)

		assert.Equal(t, http.StatusOK, w.Code)
		var res services.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, expectedRes.Token, res.Token)
		assert.Equal(t, expectedRes.User.Name, res.User.Name)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		loginInput := services.LoginInput{
			Email:    "wrong@example.com",
			Password: "password123",
		}
		mockService.On("Login", loginInput).Return(nil, assert.AnError).Once()

		body := map[string]string{
			"email":    "wrong@example.com",
			"password": "password123",
		}
		w := PerformRequest(r, "POST", "/login", body)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockService.AssertExpectations(t)
	})
}
