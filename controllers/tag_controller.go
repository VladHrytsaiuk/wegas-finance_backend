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

func (h *TagController) GetAll(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	tags, err := h.service.GetAll(currentUser.FamilyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tags)
}

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


