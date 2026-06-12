package controllers

import (
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFamilyController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockFamilyJoinService)
	controller := NewFamilyController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Set("familyID", "f-1") // Set explicitly as it's used in handler
		c.Next()
	})

	r.POST("/families/:id/generate-code", controller.GenerateCodeHandler)
	r.POST("/families/join", controller.JoinFamilyHandler)

	t.Run("GenerateCode Success", func(t *testing.T) {
		mockService.On("GenerateCode", "f-1", "member").Return("123456", nil).Once()

		body := map[string]string{"role_id": "member"}
		w := PerformRequest(r, "POST", "/families/f-1/generate-code", body)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("GenerateCode Forbidden", func(t *testing.T) {
		body := map[string]string{"role_id": "member"}
		w := PerformRequest(r, "POST", "/families/f-wrong/generate-code", body)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("JoinFamily Success", func(t *testing.T) {
		mockService.On("JoinFamily", "u-1", "123456").Return(&models.Family{Name: "Joined Family"}, nil).Once()

		body := map[string]string{"code": "123456"}
		w := PerformRequest(r, "POST", "/families/join", body)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
