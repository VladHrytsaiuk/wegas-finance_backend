package controllers

import (
	"net/http"
	"strconv"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type AdminUsersController struct {
	adminService services.AdminService
	auditService services.AuditService
}

func NewAdminUsersController(adminService services.AdminService, auditService services.AuditService) *AdminUsersController {
	return &AdminUsersController{
		adminService: adminService,
		auditService: auditService,
	}
}

func (ctrl *AdminUsersController) GetUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	search := c.Query("search")

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	users, count, err := ctrl.adminService.GetUsers(limit, offset, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": count,
	})
}

func (ctrl *AdminUsersController) ToggleBlock(c *gin.Context) {
	adminID := c.GetString("userID")
	userID := c.Param("id")

	if adminID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot block yourself"})
		return
	}

	if err := ctrl.adminService.ToggleUserBlock(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctrl.auditService.LogAction(adminID, "toggle_block", "user", userID, nil, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (ctrl *AdminUsersController) ForceLogout(c *gin.Context) {
	adminID := c.GetString("userID")
	userID := c.Param("id")

	if err := ctrl.adminService.ForceLogoutUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctrl.auditService.LogAction(adminID, "force_logout", "user", userID, nil, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (ctrl *AdminUsersController) SetRole(c *gin.Context) {
	adminID := c.GetString("userID")
	userID := c.Param("id")
	
	if adminID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot change your own admin status"})
		return
	}

	var req struct {
		IsPlatformAdmin bool `json:"is_platform_admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.adminService.SetUserPlatformAdmin(userID, req.IsPlatformAdmin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctrl.auditService.LogAction(adminID, "update_role", "user", userID, req, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
