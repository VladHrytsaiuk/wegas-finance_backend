package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type StorageTypeController struct {
	service services.StorageTypeService
}

func NewStorageTypeController(service services.StorageTypeService) *StorageTypeController {
	return &StorageTypeController{service: service}
}

// GetAll godoc
// @Summary Get all storage types
// @Description Returns a list of all storage types (system and custom) for the current family.
// @Tags StorageTypes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.StorageType
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /storage-types [get]
func (c *StorageTypeController) GetAll(ctx *gin.Context) {
	familyID, _ := ctx.Get("familyID")
	types, err := c.service.GetAll(familyID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch storage types"})
		return
	}
	ctx.JSON(http.StatusOK, types)
}

// Create godoc
// @Summary Create a new storage type
// @Description Creates a new custom storage type for the family.
// @Tags StorageTypes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.StorageType true "Storage type details"
// @Success 200 {object} models.StorageType
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /storage-types [post]
func (c *StorageTypeController) Create(ctx *gin.Context) {
	var input models.StorageType
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	familyID, _ := ctx.Get("familyID")
	fID := familyID.(string)
	input.FamilyID = &fID // Кастомні типи належать родині

	if err := c.service.Create(&input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create type"})
		return
	}

	ctx.JSON(http.StatusOK, input)
}

// Delete godoc
// @Summary Delete storage type
// @Description Deletes a custom storage type by its ID.
// @Tags StorageTypes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Storage Type ID"
// @Success 200 {object} map[string]string "Delete status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /storage-types/{id} [delete]
func (c *StorageTypeController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	// Тут варто було б перевірити, чи має юзер права і чи це не системний тип
	if err := c.service.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete type"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Type deleted"})
}
