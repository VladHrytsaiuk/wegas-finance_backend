package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type CurrencyController struct {
	service services.CurrencyService
}

func NewCurrencyController(service services.CurrencyService) *CurrencyController {
	return &CurrencyController{service: service}
}

// GetRates - Handler for GET /api/currencies
func (cc *CurrencyController) GetRates(c *gin.Context) {
	rates, err := cc.service.GetAllRates()
	if err != nil {
		// ❌ БУЛО: gin.Map (помилка)
		// ✅ СТАЛО: gin.H
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch rates",
		})
		return
	}

	c.JSON(http.StatusOK, rates)
}