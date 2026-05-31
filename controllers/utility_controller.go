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

// GetMeters godoc
// @Summary Get all utility meters
// @Description Returns a list of all utility meters for the current family.
// @Tags Utility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.UtilityMeter
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /utility/meters [get]
func (h *UtilityController) GetMeters(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	meters, err := h.service.GetMeters(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, meters)
}

// CreateMeter godoc
// @Summary Create a new utility meter
// @Description Creates a new utility meter (electricity, water, etc.).
// @Tags Utility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.UtilityMeter true "Meter details"
// @Success 201 "Created"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /utility/meters [post]
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

// UpdateMeter godoc
// @Summary Update utility meter
// @Description Updates an existing utility meter by its ID.
// @Tags Utility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Meter ID"
// @Param body body models.UtilityMeter true "Updated meter details"
// @Success 200 {object} map[string]string "Update status"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /utility/meters/{id} [put]
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

// DeleteMeter godoc
// @Summary Delete utility meter
// @Description Deletes a utility meter by its ID.
// @Tags Utility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Meter ID"
// @Success 200 {object} map[string]string "Delete status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /utility/meters/{id} [delete]
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

// GetReadings godoc
// @Summary Get utility readings
// @Description Returns a list of utility readings, optionally filtered by meter ID.
// @Tags Utility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param meter_id query string false "Filter by meter ID"
// @Success 200 {array} models.UtilityReading
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /utility/readings [get]
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

// CreateReading godoc
// @Summary Create a new utility reading
// @Description Records a new reading for a utility meter.
// @Tags Utility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.UtilityReading true "Reading details"
// @Success 201 "Created"
// @Failure 400 {object} map[string]string "Invalid input or reading sequence"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /utility/readings [post]
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

// UpdateReading godoc
// @Summary Update utility reading
// @Description Updates an existing utility reading by its ID.
// @Tags Utility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Reading ID"
// @Param body body models.UtilityReading true "Updated reading details"
// @Success 200 {object} map[string]string "Update status"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /utility/readings/{id} [put]
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

// DeleteReading godoc
// @Summary Delete utility reading
// @Description Deletes a utility reading by its ID.
// @Tags Utility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Reading ID"
// @Success 200 {object} map[string]string "Delete status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /utility/readings/{id} [delete]
func (h *UtilityController) DeleteReading(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	id := c.Param("id")
	if err := h.service.DeleteReading(id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// GetMeter godoc
// @Summary Get utility meter by ID
// @Description Returns a single utility meter by its ID.
// @Tags Utility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Meter ID"
// @Success 200 {object} models.UtilityMeter
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Meter not found"
// @Router /utility/meters/{id} [get]
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

type PayReadingJSON struct {
	AccountID string `json:"account_id" binding:"required"`
}

// PayReading godoc
// @Summary Pay for a utility reading
// @Description Processes a payment for a specific utility reading using a financial account.
// @Tags Utility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Reading ID"
// @Param body body PayReadingJSON true "Payment details"
// @Success 200 {object} map[string]string "Payment status"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /utility/readings/{id}/pay [post]
func (h *UtilityController) PayReading(c *gin.Context) {
  user := c.MustGet("user").(*models.User)
  id := c.Param("id") // ID показника

  var input PayReadingJSON
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

// GetGlobalStats godoc
// @Summary Get global utility statistics
// @Description Returns aggregated consumption and cost statistics for all utility meters.
// @Tags Utility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.UtilityStatGlobalDTO
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /utility/stats/global [get]
func (h *UtilityController) GetGlobalStats(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	stats, err := h.service.GetGlobalStats(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetMeterStats godoc
// @Summary Get statistics for a specific utility meter
// @Description Returns consumption and cost trend statistics for a single utility meter.
// @Tags Utility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Meter ID"
// @Success 200 {array} models.UtilityStatMeterDTO
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /utility/stats/meter/{id} [get]
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