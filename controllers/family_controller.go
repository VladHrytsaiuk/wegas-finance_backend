package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type FamilyController struct {
	service services.FamilyJoinService
}

func NewFamilyController(service services.FamilyJoinService) *FamilyController {
	return &FamilyController{service: service}
}

// GenerateCodeHandler godoc
// @Summary Generate a 6-digit join code for a family
// @Tags families
// @Security BearerAuth
// @Param id path string true "Family ID"
// @Param request body map[string]string true "Role ID"
// @Success 200 {object} map[string]string
// @Router /api/families/{id}/generate-code [post]
func (ctrl *FamilyController) GenerateCodeHandler(c *gin.Context) {
	familyID := c.Param("id")
	if familyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "family ID is required"})
		return
	}

	var req struct {
		RoleID string `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role_id is required"})
		return
	}

	// Перевірка прав (тільки адмін сім'ї або якщо userID має цю familyID)
	// Для простоти поки що перевіримо тільки чи юзер в цій сім'ї
	userFamilyID := c.GetString("familyID")
	if userFamilyID != familyID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you don't have permission to generate code for this family"})
		return
	}

	code, err := ctrl.service.GenerateCode(familyID, req.RoleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": code})
}

// JoinFamilyHandler godoc
// @Summary Join a family using a 6-digit code
// @Tags families
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Join Code"
// @Success 200 {object} models.Family
// @Router /api/families/join [post]
func (ctrl *FamilyController) JoinFamilyHandler(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required,len=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request, 6-digit code is required"})
		return
	}

	userID := c.GetString("userID")
	family, err := ctrl.service.JoinFamily(userID, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, family)
}
