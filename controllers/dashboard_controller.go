package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	service  services.StatsService
	userRepo repositories.UserRepository
}

func NewDashboardController(s services.StatsService, u repositories.UserRepository) *DashboardController {
	return &DashboardController{service: s, userRepo: u}
}

// Хелпер
func parseAccountIDs(ctx *gin.Context) []string {
	idsParam := ctx.Query("account_ids")
	if idsParam != "" {
		return strings.Split(idsParam, ",")
	}
	idsParamCamel := ctx.Query("accountIds")
	if idsParamCamel != "" {
		return strings.Split(idsParamCamel, ",")
	}
	return []string{}
}

// 1. Головна статистика
func (c *DashboardController) GetStats(ctx *gin.Context) {
	currentUser := ctx.MustGet("user").(*models.User)

	targetCurrency := currentUser.BaseCurrency
	if targetCurrency == "" {
		targetCurrency = "UAH"
	}

	from, to := parseDates(ctx)
	accountIDs := parseAccountIDs(ctx)

	stats, err := c.service.GetDashboardData(currentUser, targetCurrency, from, to, accountIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, stats)
}

// 2. ТОП-и
func (c *DashboardController) GetTopStats(ctx *gin.Context) {
	currentUser := ctx.MustGet("user").(*models.User)

	flowType := ctx.Query("type")
	entityType := ctx.Query("entity")

	if flowType == "" || entityType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing type or entity params"})
		return
	}

	targetCurrency := currentUser.BaseCurrency
	if targetCurrency == "" {
		targetCurrency = "UAH"
	}

	from, to := parseDates(ctx)
	accountIDs := parseAccountIDs(ctx)

	result, err := c.service.GetTopStats(currentUser, flowType, entityType, targetCurrency, from, to, accountIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

// 3. Тренд
func (c *DashboardController) GetTrend(ctx *gin.Context) {
	currentUser := ctx.MustGet("user").(*models.User)
	flowType := ctx.Query("type")

	from, to := parseDates(ctx)
	accountIDs := parseAccountIDs(ctx)

	// 🔥 Логіка визначення валюти для графіка
	// 1. URL (?currency=USD)
	targetCurrency := ctx.Query("currency")

	// 2. User Settings
	if targetCurrency == "" {
		targetCurrency = currentUser.BaseCurrency
	}

	// 3. Default
	if targetCurrency == "" {
		targetCurrency = "UAH"
	}

	trend, err := c.service.GetTrendStats(currentUser, flowType, targetCurrency, from, to, accountIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, trend)
}

// 4. Останні транзакції
func (c *DashboardController) GetRecent(ctx *gin.Context) {
	currentUser := ctx.MustGet("user").(*models.User)
	accountIDs := parseAccountIDs(ctx)

	txs, err := c.service.GetRecentTransactions(currentUser, accountIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, txs)
}

// Parse dates helper
func parseDates(ctx *gin.Context) (int64, int64) {
	from, _ := strconv.ParseInt(ctx.Query("from"), 10, 64)
	to, _ := strconv.ParseInt(ctx.Query("to"), 10, 64)

	// Fix seconds vs milliseconds
	if from > 0 && from < 100000000000 {
		from = from * 1000
	}
	if to > 0 && to < 100000000000 {
		to = to * 1000
	}

	return from, to
}