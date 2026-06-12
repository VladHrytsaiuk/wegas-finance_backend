package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDashboardController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetStats", func(t *testing.T) {
		mockService := new(services.MockStatsService)
		controller := NewDashboardController(mockService, nil)
		
		user := &models.User{Base: models.Base{ID: "u-1"}, BaseCurrency: "USD"}
		
		mockService.On("GetDashboardData", user, "USD", int64(0), int64(0), []string{}).Return(&services.DashboardData{
			TotalBalance: 1000,
		}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("GET", "/api/dashboard/stats", nil)
		
		controller.GetStats(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp services.DashboardData
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, int64(1000), resp.TotalBalance)
	})

	t.Run("GetTrend", func(t *testing.T) {
		mockService := new(services.MockStatsService)
		controller := NewDashboardController(mockService, nil)
		
		user := &models.User{Base: models.Base{ID: "u-1"}, BaseCurrency: "USD"}
		
		mockService.On("GetTrendStats", user, "expense", "USD", int64(0), int64(0), []string{}).Return([]repositories.TrendStat{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("GET", "/api/dashboard/trend?type=expense", nil)
		
		controller.GetTrend(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
