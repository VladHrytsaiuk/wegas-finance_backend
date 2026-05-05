package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type ShoppingController struct {
	service *services.ShoppingService
}

func NewShoppingController(service *services.ShoppingService) *ShoppingController {
	return &ShoppingController{service: service}
}

// === LISTS ===

func (c *ShoppingController) CreateList(ctx *gin.Context) {
	var req models.CreateShoppingListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID := ctx.GetString("userID")
	familyID := ctx.GetString("familyID")

	list, err := c.service.CreateList(req, userID, familyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, list)
}

func (c *ShoppingController) GetLists(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	familyID := ctx.GetString("familyID")

	lists, err := c.service.GetLists(familyID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, lists)
}

func (c *ShoppingController) UpdateList(ctx *gin.Context) {
	id := ctx.Param("id")
	familyID := ctx.GetString("familyID")

	var req models.UpdateShoppingListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := c.service.UpdateList(id, req, familyID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "List updated"})
}

func (c *ShoppingController) DeleteList(ctx *gin.Context) {
	id := ctx.Param("id")
	familyID := ctx.GetString("familyID")

	if err := c.service.DeleteList(id, familyID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "List deleted"})
}

// === ITEMS ===

func (c *ShoppingController) AddItem(ctx *gin.Context) {
	listID := ctx.Param("id")

	var req models.CreateShoppingItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	item, err := c.service.AddItemToList(listID, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

func (c *ShoppingController) UpdateItem(ctx *gin.Context) {
	itemID := ctx.Param("id")

	var req models.UpdateShoppingItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := c.service.UpdateItem(itemID, req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Item updated"})
}

func (c *ShoppingController) DeleteItem(ctx *gin.Context) {
	itemID := ctx.Param("id")

	if err := c.service.DeleteItem(itemID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
}

func (c *ShoppingController) ClearCompletedInList(ctx *gin.Context) {
	listID := ctx.Param("id") // ID нотатки

	if err := c.service.ClearCompletedInList(listID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Completed items cleared"})
}