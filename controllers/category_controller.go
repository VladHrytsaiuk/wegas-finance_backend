package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	service services.CategoryService
}

func NewCategoryController(service services.CategoryService) *CategoryController {
	return &CategoryController{service: service}
}

// --- HANDLERS ---

func (h *CategoryController) Create(c *gin.Context) {
	// 1. Отримуємо юзера для перевірки прав
	currentUser := c.MustGet("user").(*models.User)

	var input services.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.service.Create(input, currentUser)
	if err != nil {
		// Якщо помилка доступу (403) або інша (500)
		status := http.StatusInternalServerError
		if err.Error() == "access denied: only parents can manage categories" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (h *CategoryController) GetAll(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	categories, err := h.service.GetAll(currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (h *CategoryController) Update(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	var input services.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.service.Update(id, input, currentUser)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "access denied: only parents can manage categories" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *CategoryController) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	if err := h.service.Delete(id, currentUser); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "access denied: only parents can manage categories" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}

func (h *CategoryController) GetOne(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	category, err := h.service.GetByID(id, currentUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}