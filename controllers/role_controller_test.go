package controllers

import (
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRoleController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockRoleService)
	controller := NewRoleController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Next()
	})

	r.POST("/roles", controller.Create)
	r.GET("/roles", controller.GetAll)
	r.DELETE("/roles/:id", controller.Delete)

	t.Run("Create Role Success", func(t *testing.T) {
		input := RoleInputJSON{Name: "Manager", Description: "Can manage"}
		mockService.On("Create", mock.Anything).Return(&models.Role{
			Base: models.Base{ID: "r-1"},
			Name: "Manager",
		}, nil).Once()

		w := PerformRequest(r, "POST", "/roles", input)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("GetAll Roles Success", func(t *testing.T) {
		mockService.On("GetAll").Return([]models.Role{}, nil).Once()
		w := PerformRequest(r, "GET", "/roles", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Delete Role Success", func(t *testing.T) {
		mockService.On("Delete", "r-1").Return(nil).Once()
		w := PerformRequest(r, "DELETE", "/roles/r-1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create Role Forbidden for Child", func(t *testing.T) {
		childRouter := gin.Default()
		childRouter.Use(func(c *gin.Context) {
			SetupTestUser(c, "u-child", "f-1")
			user := c.MustGet("user").(*models.User)
			user.RoleID = "child"
			c.Next()
		})
		childRouter.POST("/roles", controller.Create)

		w := PerformRequest(childRouter, "POST", "/roles", RoleInputJSON{Name: "Bad"})
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
