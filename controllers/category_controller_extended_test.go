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

func TestCategoryController_Security(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockCategoryService)
	controller := NewCategoryController(mockService)

	t.Run("Create - Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// No user set
		controller.Create(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Create - Service Access Denied", func(t *testing.T) {
		mockService.On("Create", mock.Anything, mock.Anything).Return(nil, services.ErrAccessDenied).Once()
		
		body, _ := json.Marshal(services.CategoryInput{Name: "Forbidden"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetupTestUser(c, "child-1", "f-1")
		c.Request, _ = http.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		
		controller.Create(c)
		// The controller now correctly returns 403 for access denied errors.
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
