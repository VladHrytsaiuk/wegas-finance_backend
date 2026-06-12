package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type WishlistController struct {
	service services.WishlistService
}

func NewWishlistController(service services.WishlistService) *WishlistController {
	return &WishlistController{service: service}
}

// --- GROUPS ENDPOINTS ---

// CreateGroup godoc
// @Summary Create a new wishlist group
// @Description Creates a new group for wishlist items.
// @Tags Wishlist
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.CreateWishlistGroupRequest true "Group details"
// @Success 201 {object} models.WishlistGroup
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /wishlist-groups [post]
func (c *WishlistController) CreateGroup(ctx *gin.Context) {
	var req models.CreateWishlistGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	userID := ctx.GetString("userID")
	familyID := ctx.GetString("familyID")

	group, err := c.service.CreateGroup(req.Name, req.Color, req.Icon, req.Visibility, req.HiddenFrom, userID, familyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, group)
}

// GetGroups godoc
// @Summary Get all wishlist groups
// @Description Returns all wishlist groups accessible to the current user.
// @Tags Wishlist
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.WishlistGroup
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /wishlist-groups [get]
func (c *WishlistController) GetGroups(ctx *gin.Context) {
	familyID := ctx.GetString("familyID")
	userID := ctx.GetString("userID")
	groups, err := c.service.GetGroups(familyID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, groups)
}

// UpdateGroup godoc
// @Summary Update wishlist group
// @Description Updates an existing wishlist group by its ID.
// @Tags Wishlist
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Group ID"
// @Param body body models.UpdateWishlistGroupRequest true "Updated group details"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Invalid JSON"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /wishlist-groups/{id} [put]
func (c *WishlistController) UpdateGroup(ctx *gin.Context) {
	id := ctx.Param("id")
	familyID := ctx.GetString("familyID")

	var req models.UpdateWishlistGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if err := c.service.UpdateGroup(id, req, familyID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Group updated"})
}

// DeleteGroup godoc
// @Summary Delete wishlist group
// @Description Deletes a wishlist group by its ID.
// @Tags Wishlist
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Group ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /wishlist-groups/{id} [delete]
func (c *WishlistController) DeleteGroup(ctx *gin.Context) {
	id := ctx.Param("id")
	familyID := ctx.GetString("familyID")
	
	if err := c.service.DeleteGroup(id, familyID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Group deleted"})
}


// --- ITEMS ENDPOINTS ---

// Create godoc
// @Summary Create a new wishlist item
// @Description Creates a new item in the wishlist.
// @Tags Wishlist
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.CreateWishlistRequest true "Item details"
// @Success 201 {object} models.WishlistItem
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /wishlist [post]
func (c *WishlistController) Create(ctx *gin.Context) {
	var req models.CreateWishlistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID := ctx.GetString("userID")
	familyID := ctx.GetString("familyID")

	item, err := c.service.CreateItem(req, userID, familyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, item)
}

// GetAll godoc
// @Summary Get all wishlist items
// @Description Returns all wishlist items, optionally filtered by group or user.
// @Tags Wishlist
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param group_id query string false "Filter by group ID"
// @Param user_id query string false "Filter by user ID"
// @Success 200 {array} models.WishlistItem
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /wishlist [get]
func (c *WishlistController) GetAll(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	familyID := ctx.GetString("familyID")

	// Отримуємо параметри фільтрації з URL (?group_id=...&user_id=...)
	groupID := ctx.Query("group_id")
	targetUserID := ctx.Query("user_id")

	items, err := c.service.GetItems(familyID, userID, groupID, targetUserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, items)
}

// GetOne godoc
// @Summary Get wishlist item by ID
// @Description Returns a single wishlist item by its ID.
// @Tags Wishlist
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Item ID"
// @Success 200 {object} models.WishlistItem
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Router /wishlist/{id} [get]
func (c *WishlistController) GetOne(ctx *gin.Context) {
	id := ctx.Param("id")
	familyID := ctx.GetString("familyID")

	item, err := c.service.GetItem(id, familyID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

// Update godoc
// @Summary Update wishlist item
// @Description Updates an existing wishlist item by its ID.
// @Tags Wishlist
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Item ID"
// @Param body body models.UpdateWishlistRequest true "Updated item details"
// @Success 200 {object} models.WishlistItem
// @Failure 400 {object} map[string]string "Invalid JSON"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /wishlist/{id} [put]
func (c *WishlistController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	familyID := ctx.GetString("familyID")

	var req models.UpdateWishlistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	item, err := c.service.UpdateItem(id, req, familyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

// Delete godoc
// @Summary Delete wishlist item
// @Description Deletes a wishlist item by its ID.
// @Tags Wishlist
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Item ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /wishlist/{id} [delete]
func (c *WishlistController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	familyID := ctx.GetString("familyID")
	userID := ctx.GetString("userID") // <--- 1. Отримуємо ID юзера

	// <--- 2. Передаємо userID третім аргументом
	if err := c.service.DeleteItem(id, familyID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// UploadPhoto godoc
// @Summary Upload wishlist item photo
// @Description Uploads a photo for a wishlist item.
// @Tags Wishlist
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Item ID"
// @Param photo formData file true "Photo file"
// @Success 200 {object} map[string]string "URL to the uploaded photo"
// @Failure 400 {object} map[string]string "No file uploaded"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /wishlist/{id}/photo [post]
func (c *WishlistController) UploadPhoto(ctx *gin.Context) {
	id := ctx.Param("id")
	familyID := ctx.GetString("familyID")

	fileHeader, err := ctx.FormFile("photo")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to open file"})
		return
	}
	defer file.Close()

	url, err := c.service.UploadPhoto(id, familyID, file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"url": url})
}

// DeletePhoto godoc
// @Summary Delete wishlist item photo
// @Description Removes the photo associated with a wishlist item.
// @Tags Wishlist
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Item ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /wishlist/{id}/photo [delete]
func (c *WishlistController) DeletePhoto(ctx *gin.Context) {
	id := ctx.Param("id")
	familyID := ctx.GetString("familyID")

	if err := c.service.RemovePhoto(id, familyID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Photo removed"})
}


// ToggleReservation godoc
// @Summary Toggle wishlist item reservation
// @Description Reserves or unreserves a wishlist item for the current user.
// @Tags Wishlist
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Item ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Error message"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /wishlist/{id}/reserve [post]
func (c *WishlistController) ToggleReservation(ctx *gin.Context) {
	id := ctx.Param("id")
	userID := ctx.GetString("userID")

	if err := c.service.ToggleReservation(id, userID); err != nil {
		// Повертаємо 409 Conflict або 400 Bad Request
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Reservation toggled"})
}