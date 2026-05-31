package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type GoalController struct {
	service *services.GoalService
}

func NewGoalController(service *services.GoalService) *GoalController {
	return &GoalController{service: service}
}

// Create godoc
// @Summary Create a new financial goal
// @Description Creates a new financial goal for the family.
// @Tags Goals
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.Goal true "Goal details"
// @Success 201 {object} models.Goal
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /goals [post]
func (c *GoalController) Create(ctx *gin.Context) {
	var input models.Goal
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	familyID, _ := ctx.Get("familyID")
	userID := ctx.GetString("userID")
    
	input.FamilyID = familyID.(string)

	if err := c.service.Create(&input, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create goal"})
		return
	}
	ctx.JSON(http.StatusCreated, input)
}

// GetAll godoc
// @Summary Get all goals
// @Description Returns a list of all financial goals for the family.
// @Tags Goals
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.Goal
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /goals [get]
func (c *GoalController) GetAll(ctx *gin.Context) {
	familyID, _ := ctx.Get("familyID")
	userID := ctx.GetString("userID")

	goals, err := c.service.GetAll(familyID.(string), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, goals)
}

// GetOne godoc
// @Summary Get goal by ID
// @Description Returns a single financial goal by its ID.
// @Tags Goals
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Goal ID"
// @Success 200 {object} models.Goal
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Goal not found"
// @Router /goals/{id} [get]
func (c *GoalController) GetOne(ctx *gin.Context) {
	id := ctx.Param("id")
	userID := ctx.GetString("userID")

	goal, err := c.service.GetOne(id, userID)
	if err != nil {
        if err == services.ErrForbidden {
            ctx.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
            return
        }
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Goal not found"})
		return
	}
	ctx.JSON(http.StatusOK, goal)
}

// Update godoc
// @Summary Update goal
// @Description Updates an existing financial goal by its ID.
// @Tags Goals
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Goal ID"
// @Param body body models.Goal true "Updated goal details"
// @Success 200 {object} models.Goal
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Permission denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /goals/{id} [put]
func (c *GoalController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	userID := ctx.GetString("userID")

	var input models.Goal
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id

	if err := c.service.Update(&input, userID); err != nil {
        if err == services.ErrForbidden {
            ctx.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can edit this goal"})
            return
        }
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, input)
}

// Delete godoc
// @Summary Delete goal
// @Description Deletes a financial goal by its ID.
// @Tags Goals
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Goal ID"
// @Success 200 {object} map[string]string "Delete status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Permission denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /goals/{id} [delete]
func (c *GoalController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	userID := ctx.GetString("userID")

	if err := c.service.Delete(id, userID); err != nil {
        if err == services.ErrForbidden {
            ctx.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can delete this goal"})
            return
        }
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Goal deleted"})
}

// UploadPhoto godoc
// @Summary Upload goal photo
// @Description Uploads a photo for a financial goal.
// @Tags Goals
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Goal ID"
// @Param file formData file true "Photo file"
// @Success 200 {object} map[string]string "Success status and photo URL"
// @Failure 400 {object} map[string]string "No file uploaded"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Permission denied"
// @Failure 404 {object} map[string]string "Goal not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /goals/{id}/photo [post]
func (c *GoalController) UploadPhoto(ctx *gin.Context) {
	goalID := ctx.Param("id")
	userID := ctx.GetString("userID")

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

    // Тепер передаємо userID і отримуємо готовий URL
	photoURL, err := c.service.UploadGoalPhoto(goalID, userID, file)
	if err != nil {
        status := http.StatusInternalServerError
        if err == services.ErrForbidden {
            status = http.StatusForbidden
        } else if err == services.ErrGoalNotFound {
            status = http.StatusNotFound
        }
		ctx.JSON(status, gin.H{"error": err.Error()})
		return
	}

    // Більше не треба викликати Update окремо, Service вже все зробив
	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Photo uploaded successfully",
		"photo_url": photoURL,
	})
}

type LinkAccountJSON struct {
	AccountID string `json:"account_id" binding:"required"`
}

// LinkAccount godoc
// @Summary Link account to goal
// @Description Associates a financial account with a financial goal.
// @Tags Goals
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Goal ID"
// @Param body body LinkAccountJSON true "Account link details"
// @Success 200 {object} map[string]string "Success status"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /goals/{id}/link-account [post]
func (c *GoalController) LinkAccount(ctx *gin.Context) {
	goalID := ctx.Param("id")

	var req LinkAccountJSON
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.LinkAccount(goalID, req.AccountID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link account"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Account linked successfully"})
}

// UnlinkAccount godoc
// @Summary Unlink account from goal
// @Description Removes the association between a financial account and any goal.
// @Tags Goals
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Goal ID (unused, but kept for route symmetry)"
// @Param body body LinkAccountJSON true "Account unlink details"
// @Success 200 {object} map[string]string "Success status"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /goals/{id}/unlink-account [post]
func (c *GoalController) UnlinkAccount(ctx *gin.Context) {
	var req LinkAccountJSON
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.UnlinkAccount(req.AccountID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlink account"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Account unlinked successfully"})
}