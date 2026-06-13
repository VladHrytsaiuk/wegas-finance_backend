package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserController_Errors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockUserService)
	controller := NewUserController(mockService)

	t.Run("AddMember - Forbidden", func(t *testing.T) {
		user := &models.User{Base: models.Base{ID: "child-1"}, RoleID: "child"}
		mockService.On("AddMember", user, mock.Anything).Return(nil, errors.New("permission denied: only parents can manage members")).Once()

		body, _ := json.Marshal(AddMemberJSON{Name: "New", Email: "new@test.com", Password: "password"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("POST", "/family/users", bytes.NewBuffer(body))
		
		controller.AddMember(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("AddMember - Bad Request", func(t *testing.T) {
		user := &models.User{Base: models.Base{ID: "parent-1"}, RoleID: "admin"}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("POST", "/family/users", bytes.NewBuffer([]byte("invalid")))
		
		controller.AddMember(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateProfile - Service Error", func(t *testing.T) {
		user := &models.User{Base: models.Base{ID: "u-1"}}
		mockService.On("UpdateProfile", "u-1", "Name", "email@test.com").Return(nil, assert.AnError).Once()

		body, _ := json.Marshal(UpdateProfileJSON{Name: "Name", Email: "email@test.com"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("PUT", "/users/me", bytes.NewBuffer(body))
		
		controller.UpdateProfile(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
