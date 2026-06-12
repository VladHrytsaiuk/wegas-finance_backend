package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockUserService)
	controller := NewUserController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Next()
	})
	
	r.GET("/users/me", controller.GetMe)
	r.GET("/users", controller.GetFamilyMembers)
	r.POST("/family/users", controller.AddMember)
	r.PUT("/users/me", controller.UpdateProfile)
	r.PUT("/users/password", controller.ChangePassword)

	t.Run("GetMe", func(t *testing.T) {
		w := PerformRequest(r, "GET", "/users/me", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		var user models.User
		json.Unmarshal(w.Body.Bytes(), &user)
		assert.Equal(t, "u-1", user.ID)
	})

	t.Run("GetFamilyMembers", func(t *testing.T) {
		mockService.On("GetFamilyMembers", mock.Anything).Return([]models.User{
			{Base: models.Base{ID: "u-1"}},
			{Base: models.Base{ID: "u-2"}},
		}, nil).Once()

		w := PerformRequest(r, "GET", "/users", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		var members []models.User
		json.Unmarshal(w.Body.Bytes(), &members)
		assert.Len(t, members, 2)
		mockService.AssertExpectations(t)
	})

	t.Run("AddMember Success", func(t *testing.T) {
		input := AddMemberJSON{
			Name: "Child", Email: "c@t.com", Password: "password",
		}
		mockService.On("AddMember", mock.Anything, mock.MatchedBy(func(in services.CreateUserInput) bool {
			return in.Name == "Child"
		})).Return(&models.User{Base: models.Base{ID: "u-child"}}, nil).Once()

		w := PerformRequest(r, "POST", "/family/users", input)
		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("UpdateProfile", func(t *testing.T) {
		input := UpdateProfileJSON{Name: "New Name"}
		mockService.On("UpdateProfile", "u-1", "New Name", "").Return(&models.User{Name: "New Name"}, nil).Once()

		w := PerformRequest(r, "PUT", "/users/me", input)
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
