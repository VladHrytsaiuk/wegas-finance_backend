package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type CounterpartyController struct {
	service services.CounterpartyService
}

func NewCounterpartyController(service services.CounterpartyService) *CounterpartyController {
	return &CounterpartyController{service: service}
}

// === JSON STRUCTS ===

type CpCategoryJSON struct {
	Name  string `json:"name" binding:"required"`
	Type  string `json:"type" binding:"required"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

type CounterpartyJSON struct {
	Name       string  `json:"name" binding:"required"`
	Type       string  `json:"type" binding:"required"`
	CategoryID *string `json:"category_id"`
	Icon       string  `json:"icon"`
	Logo       string  `json:"logo"` // Приймає рядок (Base64 або filename)
}

// === HANDLERS: CATEGORIES ===

func (h *CounterpartyController) CreateCategory(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	var json CpCategoryJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.CpCategoryInput{
		Name:  json.Name,
		Type:  json.Type,
		Icon:  json.Icon,
		Color: json.Color,
	}

	res, err := h.service.CreateCategory(input, currentUser)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *CounterpartyController) GetCategories(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	list, err := h.service.GetCategories(currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *CounterpartyController) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	var json CpCategoryJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.CpCategoryInput{
		Name:  json.Name,
		Type:  json.Type,
		Icon:  json.Icon,
		Color: json.Color,
	}

	res, err := h.service.UpdateCategory(id, input, currentUser)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *CounterpartyController) GetCategory(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	cat, err := h.service.GetCategoryByID(id, currentUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	c.JSON(http.StatusOK, cat)
}

// === HANDLERS: COUNTERPARTIES ===

func (h *CounterpartyController) Create(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	var json CounterpartyJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.CounterpartyInput{
		Name:       json.Name,
		Type:       json.Type,
		CategoryID: json.CategoryID,
		Icon:       json.Icon,
		Logo:       json.Logo,
	}

	res, err := h.service.Create(input, currentUser)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *CounterpartyController) GetAll(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	list, err := h.service.GetAll(currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *CounterpartyController) Update(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	var json CounterpartyJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.CounterpartyInput{
		Name:       json.Name,
		Type:       json.Type,
		CategoryID: json.CategoryID,
		Icon:       json.Icon,
		Logo:       json.Logo,
	}

	res, err := h.service.Update(id, input, currentUser)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *CounterpartyController) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	if err := h.service.Delete(id, currentUser); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Counterparty deleted"})
}

func (h *CounterpartyController) GetOne(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	cp, err := h.service.GetByID(id, currentUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Counterparty not found"})
		return
	}
	c.JSON(http.StatusOK, cp)
}

// Helper for error responses
func handleError(c *gin.Context, err error) {
	if err.Error() == "access denied: only parents can manage counterparties" {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	} else if err.Error() == "logo file is too large (max 20KB)" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}