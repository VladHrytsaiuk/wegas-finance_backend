package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type ExportController struct {
	service services.ExportService
}

func NewExportController(s services.ExportService) *ExportController {
	return &ExportController{service: s}
}

// ExportTransactions обробляє запит на експорт
func (c *ExportController) ExportTransactions(ctx *gin.Context) {
	// 1. 🔥 Отримуємо повного юзера для перевірки прав (RoleID)
	currentUser := ctx.MustGet("user").(*models.User)

	var filter models.ExportFilterDTO

	// 2. Біндинг параметрів
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters"})
		return
	}

	// 3. Ручна обробка масивів (fallback)
	if len(filter.AccountIDs) == 0 && ctx.Query("accountIds") != "" {
		filter.AccountIDs = strings.Split(ctx.Query("accountIds"), ",")
	}
	if len(filter.CategoryIDs) == 0 && ctx.Query("categoryIds") != "" {
		filter.CategoryIDs = strings.Split(ctx.Query("categoryIds"), ",")
	}
	if len(filter.UserIDs) == 0 && ctx.Query("userIds") != "" {
		filter.UserIDs = strings.Split(ctx.Query("userIds"), ",")
	}
	if len(filter.CounterpartyIDs) == 0 && ctx.Query("counterpartyIds") != "" {
		filter.CounterpartyIDs = strings.Split(ctx.Query("counterpartyIds"), ",")
	}

	// 4. Отримання даних (Передаємо currentUser!)
	data, err := c.service.GetTransactions(currentUser, filter)
	if err != nil {
		fmt.Println("EXPORT ERROR:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}