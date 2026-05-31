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

// GetRates godoc
// @Summary Get currency rates
// @Description Returns the latest exchange rates for all supported currencies.
// @Tags Currency
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.ExchangeRate
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /currencies [get]
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