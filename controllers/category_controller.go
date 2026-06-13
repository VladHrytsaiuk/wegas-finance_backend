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

// Create godoc
// @Summary Create a new category
// @Description Creates a new transaction category. Only accessible by family parents/admins.
// @Tags Categories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body services.CategoryInput true "Category details"
// @Success 201 {object} models.Category
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Permission denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /categories [post]
func (h *CategoryController) Create(c *gin.Context) {
	// 1. Отримуємо юзера для перевірки прав
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var input services.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.service.Create(input, user)
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

// GetAll godoc
// @Summary Get all categories
// @Description Returns a list of all transaction categories for the current family.
// @Tags Categories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.Category
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /categories [get]
func (h *CategoryController) GetAll(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	categories, err := h.service.GetAll(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// Update godoc
// @Summary Update category
// @Description Updates an existing category by its ID. Only accessible by family parents/admins.
// @Tags Categories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Category ID"
// @Param body body services.CategoryInput true "Updated category details"
// @Success 200 {object} models.Category
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Permission denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /categories/{id} [put]
func (h *CategoryController) Update(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var input services.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.service.Update(id, input, user)
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

// Delete godoc
// @Summary Delete category
// @Description Deletes a category by its ID. Only accessible by family parents/admins.
// @Tags Categories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Category ID"
// @Success 200 {object} map[string]string "Category deleted"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Permission denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /categories/{id} [delete]
func (h *CategoryController) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	if err := h.service.Delete(id, user); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "access denied: only parents can manage categories" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}

// GetOne godoc
// @Summary Get category by ID
// @Description Returns a single category by its ID.
// @Tags Categories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Category ID"
// @Success 200 {object} models.Category
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Category not found"
// @Router /categories/{id} [get]
func (h *CategoryController) GetOne(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	category, err := h.service.GetByID(id, user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}