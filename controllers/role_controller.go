package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type RoleController struct {
	service services.RoleService
}

func NewRoleController(service services.RoleService) *RoleController {
	return &RoleController{service: service}
}

type RoleInputJSON struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	CanManageUsers bool   `json:"can_manage_users"`
	CanEditSchema  bool   `json:"can_edit_schema"`
}

func (h *RoleController) Create(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	// 🛑 ЗАХИСТ: Дитина не лізе в налаштування ролей
	if currentUser.RoleID == "child" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var json RoleInputJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.service.Create(services.CreateRoleInput{
		Name:           json.Name,
		Description:    json.Description,
		CanManageUsers: json.CanManageUsers,
		CanEditSchema:  json.CanEditSchema,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *RoleController) GetAll(c *gin.Context) {
	// Читати ролі можуть усі (щоб фронтенд знав права)
	roles, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *RoleController) Delete(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	// 🛑 ЗАХИСТ
	if currentUser.RoleID == "child" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	id := c.Param("id")
	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted"})
}