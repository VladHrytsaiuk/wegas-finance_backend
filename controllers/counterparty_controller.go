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

// CreateCategory godoc
// @Summary Create a new counterparty category
// @Description Creates a new category for counterparties.
// @Tags Counterparties
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body CpCategoryJSON true "Category details"
// @Success 201 {object} models.CounterpartyCategory
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /counterparty-categories [post]
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

// GetCategories godoc
// @Summary Get all counterparty categories
// @Description Returns a list of all counterparty categories for the family.
// @Tags Counterparties
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.CounterpartyCategory
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /counterparty-categories [get]
func (h *CounterpartyController) GetCategories(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	list, err := h.service.GetCategories(currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// UpdateCategory godoc
// @Summary Update counterparty category
// @Description Updates an existing counterparty category by its ID.
// @Tags Counterparties
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Category ID"
// @Param body body CpCategoryJSON true "Updated category details"
// @Success 200 {object} models.CounterpartyCategory
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /counterparty-categories/{id} [put]
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

// GetCategory godoc
// @Summary Get counterparty category by ID
// @Description Returns a single counterparty category by its ID.
// @Tags Counterparties
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Category ID"
// @Success 200 {object} models.CounterpartyCategory
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Category not found"
// @Router /counterparty-categories/{id} [get]
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

// Create godoc
// @Summary Create a new counterparty
// @Description Creates a new counterparty.
// @Tags Counterparties
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body CounterpartyJSON true "Counterparty details"
// @Success 201 {object} models.Counterparty
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /counterparties [post]
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

// GetAll godoc
// @Summary Get all counterparties
// @Description Returns a list of all counterparties for the family.
// @Tags Counterparties
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.Counterparty
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /counterparties [get]
func (h *CounterpartyController) GetAll(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	list, err := h.service.GetAll(currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// Update godoc
// @Summary Update counterparty
// @Description Updates an existing counterparty by its ID.
// @Tags Counterparties
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Counterparty ID"
// @Param body body CounterpartyJSON true "Updated counterparty details"
// @Success 200 {object} models.Counterparty
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /counterparties/{id} [put]
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

// Delete godoc
// @Summary Delete counterparty
// @Description Deletes a counterparty by its ID.
// @Tags Counterparties
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Counterparty ID"
// @Success 200 {object} map[string]string "Counterparty deleted"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /counterparties/{id} [delete]
func (h *CounterpartyController) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	if err := h.service.Delete(id, currentUser); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Counterparty deleted"})
}

// GetOne godoc
// @Summary Get counterparty by ID
// @Description Returns a single counterparty by its ID.
// @Tags Counterparties
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Counterparty ID"
// @Success 200 {object} models.Counterparty
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Counterparty not found"
// @Router /counterparties/{id} [get]
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