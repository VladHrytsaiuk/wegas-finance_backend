package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type ShoppingController struct {
	service services.ShoppingService
}

func NewShoppingController(service services.ShoppingService) *ShoppingController {
	return &ShoppingController{service: service}
}

// === LISTS ===

// CreateList godoc
// @Summary Create a new shopping list
// @Description Creates a new shopping list/note.
// @Tags Shopping
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.CreateShoppingListRequest true "List details"
// @Success 201 {object} models.ShoppingList
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shopping-lists [post]
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

// GetLists godoc
// @Summary Get all shopping lists
// @Description Returns all shopping lists accessible to the current user.
// @Tags Shopping
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.ShoppingList
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shopping-lists [get]
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

// UpdateList godoc
// @Summary Update a shopping list
// @Description Updates an existing shopping list by its ID.
// @Tags Shopping
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "List ID"
// @Param body body models.UpdateShoppingListRequest true "Updated list details"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shopping-lists/{id} [put]
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

// DeleteList godoc
// @Summary Delete a shopping list
// @Description Deletes a shopping list by its ID.
// @Tags Shopping
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "List ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shopping-lists/{id} [delete]
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

// AddItem godoc
// @Summary Add item to shopping list
// @Description Adds a new item to an existing shopping list.
// @Tags Shopping
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "List ID"
// @Param body body models.CreateShoppingItemRequest true "Item details"
// @Success 201 {object} models.ShoppingItem
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shopping-lists/{id}/items [post]
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

// UpdateItem godoc
// @Summary Update a shopping item
// @Description Updates an item (e.g., mark as bought) by its ID.
// @Tags Shopping
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Item ID"
// @Param body body models.UpdateShoppingItemRequest true "Updated item details"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shopping-items/{id} [put]
func (h *ShoppingController) UpdateItem(ctx *gin.Context) {
	itemID := ctx.Param("id")

	var req models.UpdateShoppingItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.UpdateItem(itemID, req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Item updated"})
}

// DeleteItem godoc
// @Summary Delete a shopping item
// @Description Deletes an item from a shopping list by its ID.
// @Tags Shopping
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Item ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shopping-items/{id} [delete]
func (c *ShoppingController) DeleteItem(ctx *gin.Context) {
	itemID := ctx.Param("id")

	if err := c.service.DeleteItem(itemID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
}

// ClearCompletedInList godoc
// @Summary Clear completed items in a list
// @Description Deletes all items marked as bought in a specific shopping list.
// @Tags Shopping
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "List ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shopping-lists/{id}/completed [delete]
func (c *ShoppingController) ClearCompletedInList(ctx *gin.Context) {
	listID := ctx.Param("id") // ID нотатки

	if err := c.service.ClearCompletedInList(listID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Completed items cleared"})
}