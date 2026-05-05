package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type StorageTypeController struct {
	service *services.StorageTypeService
}

func NewStorageTypeController(service *services.StorageTypeService) *StorageTypeController {
	return &StorageTypeController{service: service}
}

func (c *StorageTypeController) GetAll(ctx *gin.Context) {
	familyID, _ := ctx.Get("familyID")
	types, err := c.service.GetAll(familyID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch storage types"})
		return
	}
	ctx.JSON(http.StatusOK, types)
}

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

func (c *StorageTypeController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	// Тут варто було б перевірити, чи має юзер права і чи це не системний тип
	if err := c.service.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete type"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Type deleted"})
}
