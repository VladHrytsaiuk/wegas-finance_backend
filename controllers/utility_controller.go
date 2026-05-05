package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type UtilityController struct {
	service *services.UtilityService
}

func NewUtilityController(s *services.UtilityService) *UtilityController {
	return &UtilityController{service: s}
}

// === METERS (Лічильники) ===

// GET /utility/meters
func (h *UtilityController) GetMeters(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	meters, err := h.service.GetMeters(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, meters)
}

// POST /utility/meters
func (h *UtilityController) CreateMeter(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	var input models.UtilityMeter
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.CreateMeter(input, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

// PUT /utility/meters/:id
func (h *UtilityController) UpdateMeter(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("id")
	var input models.UtilityMeter
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateMeter(id, input, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// DELETE /utility/meters/:id
func (h *UtilityController) DeleteMeter(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("id")
	if err := h.service.DeleteMeter(id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// === READINGS (Показники) ===

// GET /utility/readings?meter_id=...
func (h *UtilityController) GetReadings(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	meterID := c.Query("meter_id")
	// Якщо meter_id пустий, сервіс поверне всі або помилку, залежно від логіки.
	// Тут краще валідувати, але сервіс впорається.
	readings, err := h.service.GetReadings(user, meterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, readings)
}

// POST /utility/readings
func (h *UtilityController) CreateReading(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	var input models.UtilityReading
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.CreateReading(input, user); err != nil {
		// Часто буває помилка "нове значення менше попереднього" - це 400 Bad Request
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

// PUT /utility/readings/:id
func (h *UtilityController) UpdateReading(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("id")

	var input models.UtilityReading
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateReading(id, input, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// DELETE /utility/readings/:id
func (h *UtilityController) DeleteReading(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("id")
	if err := h.service.DeleteReading(id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
func (h *UtilityController) GetMeter(c *gin.Context) {
    user := c.MustGet("user").(*models.User)
    id := c.Param("id")
    
    meter, err := h.service.GetMeterByID(id, user)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Meter not found"})
        return
    }
    c.JSON(http.StatusOK, meter)
}

func (h *UtilityController) PayReading(c *gin.Context) {
  user := c.MustGet("user").(*models.User)
  id := c.Param("id") // ID показника

  var input struct {
    AccountID string `json:"account_id" binding:"required"`
  }
  if err := c.ShouldBindJSON(&input); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  if err := h.service.PayReading(id, input.AccountID, user); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"status": "paid"})
}

func (h *UtilityController) GetGlobalStats(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	stats, err := h.service.GetGlobalStats(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GET /utility/stats/meter/:id
func (h *UtilityController) GetMeterStats(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	meterID := c.Param("id")

	stats, err := h.service.GetMeterStats(meterID, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}