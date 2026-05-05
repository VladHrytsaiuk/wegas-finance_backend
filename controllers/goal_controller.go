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

// ... Link/Unlink methods ...

func (c *GoalController) LinkAccount(ctx *gin.Context) {
	goalID := ctx.Param("id")

	var req struct {
		AccountID string `json:"account_id" binding:"required"`
	}
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

func (c *GoalController) UnlinkAccount(ctx *gin.Context) {
	var req struct {
		AccountID string `json:"account_id" binding:"required"`
	}
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