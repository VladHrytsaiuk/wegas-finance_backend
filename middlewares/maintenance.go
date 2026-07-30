package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/database"
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/gin-gonic/gin"
)

var (
	maintenanceModeCache bool
	lastMaintenanceCheck time.Time
	maintenanceMutex     sync.RWMutex
)

// IsMaintenanceMode returns the current cached value of maintenance mode.
func IsMaintenanceMode() bool {
	maintenanceMutex.RLock()
	if !lastMaintenanceCheck.IsZero() && time.Since(lastMaintenanceCheck) < 10*time.Second {
		val := maintenanceModeCache
		maintenanceMutex.RUnlock()
		return val
	}
	maintenanceMutex.RUnlock()

	var setting models.SystemSetting
	err := database.DB.Where("key = ?", "maintenance_mode").First(&setting).Error

	maintenanceMutex.Lock()
	defer maintenanceMutex.Unlock()
	if err == nil {
		maintenanceModeCache = setting.Value == "true"
	} else {
		maintenanceModeCache = false
	}
	lastMaintenanceCheck = time.Now()

	return maintenanceModeCache
}

// ForceUpdateMaintenanceCache updates the cache immediately.
func ForceUpdateMaintenanceCache(enabled bool) {
	maintenanceMutex.Lock()
	defer maintenanceMutex.Unlock()
	maintenanceModeCache = enabled
	lastMaintenanceCheck = time.Now()
}

// MaintenanceMiddleware blocks requests if maintenance mode is active, unless user is platform admin.
func MaintenanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isMaintenance := IsMaintenanceMode()

		if isMaintenance {
			// Check if the current user is a platform admin
			if userObj, exists := c.Get("user"); exists {
				if user, ok := userObj.(*models.User); ok && user.IsPlatformAdmin {
					// Admins bypass maintenance mode
					c.Next()
					return
				}
			}

			// If not admin or not authenticated, return 503
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "System is currently under maintenance. Please try again later.",
			})
			return
		}

		c.Next()
	}
}
