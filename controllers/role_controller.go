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

// Create godoc
// @Summary Create a new role
// @Description Creates a new user role. Only accessible by family parents/admins.
// @Tags Roles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body RoleInputJSON true "Role details"
// @Success 201 {object} models.Role
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Permission denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /roles [post]
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

// GetAll godoc
// @Summary Get all roles
// @Description Returns a list of all available roles.
// @Tags Roles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.Role
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /roles [get]
func (h *RoleController) GetAll(c *gin.Context) {
	// Читати ролі можуть усі (щоб фронтенд знав права)
	roles, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}

// Delete godoc
// @Summary Delete role
// @Description Deletes a role by its ID. Only accessible by family parents/admins.
// @Tags Roles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Role ID"
// @Success 200 {object} map[string]string "Role deleted"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Permission denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /roles/{id} [delete]
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