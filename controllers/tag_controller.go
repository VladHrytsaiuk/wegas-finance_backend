package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type TagController struct {
	service services.TagService
}

func NewTagController(service services.TagService) *TagController {
	return &TagController{service: service}
}

type TagInputJSON struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

// Create godoc
// @Summary Create a new tag
// @Description Creates a new tag for transactions.
// @Tags Tags
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body TagInputJSON true "Tag details"
// @Success 201 {object} models.Tag
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /tags [post]
func (h *TagController) Create(c *gin.Context) {
	// Отримуємо юзера (і батьки, і діти мають право)
	currentUser := c.MustGet("user").(*models.User)

	var json TagInputJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.service.Create(json.Name, json.Color, currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tag)
}

// GetAll godoc
// @Summary Get all tags
// @Description Returns a list of all tags for the current family.
// @Tags Tags
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.Tag
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /tags [get]
func (h *TagController) GetAll(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	tags, err := h.service.GetAll(currentUser.FamilyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tags)
}

// Delete godoc
// @Summary Delete tag
// @Description Deletes a tag by its ID.
// @Tags Tags
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Tag ID"
// @Success 200 {object} map[string]string "Tag deleted"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /tags/{id} [delete]
func (h *TagController) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	// Дозволяємо видаляти і дітям теж
	if err := h.service.Delete(id, currentUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted"})
}


