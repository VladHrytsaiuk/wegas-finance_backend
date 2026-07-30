package controllers

import (
	"net/http"
	"sync"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/middlewares"
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type AdminOverviewController struct {
	db           *gorm.DB
	auditService services.AuditService
}

func NewAdminOverviewController(db *gorm.DB, auditService services.AuditService) *AdminOverviewController {
	return &AdminOverviewController{db: db, auditService: auditService}
}

type AdminOverviewStats struct {
	TotalUsers        int64 `json:"total_users"`
	TotalFamilies     int64 `json:"total_families"`
	TotalTransactions int64 `json:"total_transactions"`
	TotalInboxEntries int64 `json:"total_inbox_entries"`
	NewUsers7Days     int64 `json:"new_users_7_days"`
	InboxParseErrors  int64 `json:"inbox_parse_errors"`
	FailedMonobank    int64 `json:"failed_monobank"`
	MaintenanceMode   bool  `json:"maintenance_mode"`
}

var (
	cachedOverviewStats   AdminOverviewStats
	lastOverviewStatsTime time.Time
	overviewStatsMutex    sync.RWMutex
)

func (h *AdminOverviewController) GetStats(c *gin.Context) {
	overviewStatsMutex.RLock()
	if time.Since(lastOverviewStatsTime) < 60*time.Second {
		stats := cachedOverviewStats
		overviewStatsMutex.RUnlock()
		c.JSON(http.StatusOK, stats)
		return
	}
	overviewStatsMutex.RUnlock()

	var stats AdminOverviewStats
	var eg errgroup.Group

	// 1. Користувачі
	eg.Go(func() error {
		return h.db.Model(&models.User{}).Where("deleted_at IS NULL").Count(&stats.TotalUsers).Error
	})

	// 2. Сім'ї
	eg.Go(func() error {
		return h.db.Model(&models.Family{}).Where("deleted_at IS NULL").Count(&stats.TotalFamilies).Error
	})

	// 3. Транзакції
	eg.Go(func() error {
		return h.db.Model(&models.Transaction{}).Where("deleted_at IS NULL").Count(&stats.TotalTransactions).Error
	})

	// 4. Чеки
	eg.Go(func() error {
		return h.db.Model(&models.InboxEntry{}).Where("deleted_at IS NULL").Count(&stats.TotalInboxEntries).Error
	})

	// 5. Нові користувачі (за останні 7 днів)
	eg.Go(func() error {
		sevenDaysAgo := time.Now().AddDate(0, 0, -7).UnixMilli()
		return h.db.Model(&models.User{}).Where("created_at > ? AND deleted_at IS NULL", sevenDaysAgo).Count(&stats.NewUsers7Days).Error
	})

	// 6. Помилки парсингу
	eg.Go(func() error {
		return h.db.Model(&models.InboxEntry{}).Where("(status = ? OR reason != ?) AND deleted_at IS NULL", models.InboxEntryStatusNeedsReview, "").Count(&stats.InboxParseErrors).Error
	})

	// 7. Невдалі синхронізації Monobank (старіші за 24 години)
	eg.Go(func() error {
		oneDayAgo := time.Now().Add(-24 * time.Hour)
		return h.db.Model(&models.BankConnection{}).Where("last_sync < ? AND is_active = ? AND deleted_at IS NULL", oneDayAgo, true).Count(&stats.FailedMonobank).Error
	})

	// 8. Maintenance Mode
	eg.Go(func() error {
		var setting models.SystemSetting
		if err := h.db.Where("key = ?", "maintenance_mode").First(&setting).Error; err == nil {
			stats.MaintenanceMode = setting.Value == "true"
		}
		// Ігноруємо помилку (напр. RecordNotFound), якщо налаштування ще не створене
		return nil
	})

	if err := eg.Wait(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard statistics"})
		return
	}

	overviewStatsMutex.Lock()
	cachedOverviewStats = stats
	lastOverviewStatsTime = time.Now()
	overviewStatsMutex.Unlock()

	c.JSON(http.StatusOK, stats)
}

type toggleMaintenanceInput struct {
	Enabled bool `json:"enabled"`
}

func (h *AdminOverviewController) ToggleMaintenance(c *gin.Context) {
	adminID := c.GetString("userID")
	var input toggleMaintenanceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	valueStr := "false"
	if input.Enabled {
		valueStr = "true"
	}

	// Update or Create the system setting
	if err := h.db.Where("key = ?", "maintenance_mode").Assign(models.SystemSetting{
		Value: valueStr,
	}).FirstOrCreate(&models.SystemSetting{
		Key: "maintenance_mode",
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update maintenance mode"})
		return
	}

	middlewares.ForceUpdateMaintenanceCache(input.Enabled)
	h.auditService.LogAction(adminID, "toggle_maintenance", "system_setting", "maintenance_mode", input, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"maintenance_mode": input.Enabled})
}
