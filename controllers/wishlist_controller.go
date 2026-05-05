package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type WishlistController struct {
	service *services.WishlistService
}

func NewWishlistController(service *services.WishlistService) *WishlistController {
	return &WishlistController{service: service}
}

// --- GROUPS ENDPOINTS ---

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

func (c *WishlistController) DeletePhoto(ctx *gin.Context) {
	id := ctx.Param("id")
	familyID := ctx.GetString("familyID")

	if err := c.service.RemovePhoto(id, familyID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Photo removed"})
}


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