package controllers

import (
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockExportService)
	controller := NewExportController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Next()
	})
	r.GET("/export/transactions", controller.ExportTransactions)

	t.Run("Export Success", func(t *testing.T) {
		mockService.On("GetTransactions", mock.Anything, mock.Anything).Return([]models.Transaction{
			{Amount: 100},
		}, nil).Once()

		w := PerformRequest(r, "GET", "/export/transactions?from=123&to=456", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	r.GET("/export/backup", controller.ExportBackup)

	t.Run("Backup Success", func(t *testing.T) {
		mockService.On("GetBackup", mock.Anything).Return(&models.BackupDTO{
			Transactions: []models.Transaction{{Amount: 100}},
		}, nil).Once()

		w := PerformRequest(r, "GET", "/export/backup", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "attachment; filename=backup.json", w.Header().Get("Content-Disposition"))
		mockService.AssertExpectations(t)
	})
}
