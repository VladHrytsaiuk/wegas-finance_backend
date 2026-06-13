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

func TestUserController_Extended(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ChangePassword - Success", func(t *testing.T) {
		mockService := new(services.MockUserService)
		controller := NewUserController(mockService)
		user := &models.User{Base: models.Base{ID: "u-1"}}
		
		mockService.On("ChangePassword", "u-1", "old-pass", "new-password-123").Return(nil).Once()

		body, _ := json.Marshal(map[string]string{
			"old_password": "old-pass",
			"new_password": "new-password-123",
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("POST", "/api/users/change-password", bytes.NewBuffer(body))
		
		controller.ChangePassword(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("DeleteMember - Success", func(t *testing.T) {
		mockService := new(services.MockUserService)
		controller := NewUserController(mockService)
		actor := &models.User{Base: models.Base{ID: "parent-1"}, RoleID: "admin"}
		
		mockService.On("DeleteMember", actor, "target-1").Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", actor)
		c.Params = gin.Params{{Key: "id", Value: "target-1"}}
		
		controller.DeleteMember(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("LeaveFamily - Success", func(t *testing.T) {
		mockService := new(services.MockUserService)
		controller := NewUserController(mockService)
		user := &models.User{Base: models.Base{ID: "u-1"}}
		
		mockService.On("LeaveFamily", user).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		
		controller.LeaveFamily(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("UpdateUser - Success", func(t *testing.T) {
		mockService := new(services.MockUserService)
		controller := NewUserController(mockService)
		actor := &models.User{Base: models.Base{ID: "parent-1"}}
		input := services.CreateUserInput{Name: "New Name", RoleID: "child"}
		
		mockService.On("UpdateUser", actor, "target-1", mock.Anything).Return(&models.User{Name: "New Name"}, nil).Once()

		body, _ := json.Marshal(input)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", actor)
		c.Params = gin.Params{{Key: "id", Value: "target-1"}}
		c.Request, _ = http.NewRequest("PUT", "/api/users/target-1", bytes.NewBuffer(body))
		
		controller.UpdateUser(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
