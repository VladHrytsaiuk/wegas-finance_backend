package controllers

import (
	"net/http"
	"strconv"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type AdminAuditController struct {
	adminService services.AdminService
	auditService services.AuditService
}

func NewAdminAuditController(adminService services.AdminService, auditService services.AuditService) *AdminAuditController {
	return &AdminAuditController{
		adminService: adminService,
		auditService: auditService,
	}
}

func (ctrl *AdminAuditController) GetLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	logs, count, err := ctrl.auditService.GetLogs(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": count,
	})
}

func (ctrl *AdminAuditController) GetSettings(c *gin.Context) {
	settings, err := ctrl.adminService.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}
