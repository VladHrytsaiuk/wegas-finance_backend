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
	"github.com/stretchr/testify/mock"
)

func TestDashboardControllerExtended(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetStats - Success", func(t *testing.T) {
		mockService := new(services.MockStatsService)
		controller := NewDashboardController(mockService, nil)
		user := &models.User{Base: models.Base{ID: "u-1"}, BaseCurrency: "USD"}
		
		mockService.On("GetDashboardData", user, "USD", int64(1000000), int64(2000000), []string{"acc1"}).
			Return(&services.DashboardData{TotalBalance: 5000}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("GET", "/api/dashboard/stats?from=1000&to=2000&account_ids=acc1", nil)
		
		controller.GetStats(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp services.DashboardData
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, int64(5000), resp.TotalBalance)
	})

	t.Run("GetTopStats - Success", func(t *testing.T) {
		mockService := new(services.MockStatsService)
		controller := NewDashboardController(mockService, nil)
		user := &models.User{Base: models.Base{ID: "u-1"}, BaseCurrency: "USD"}
		
		mockService.On("GetTopStats", user, "expense", "category", "USD", mock.Anything, mock.Anything, mock.Anything).
			Return([]repositories.TopStat{{Name: "Food", Total: 100}}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("GET", "/api/dashboard/top?type=expense&entity=category", nil)
		
		controller.GetTopStats(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []repositories.TopStat
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Food", resp[0].Name)
	})

	t.Run("GetTopStats - Missing Params", func(t *testing.T) {
		controller := NewDashboardController(nil, nil)
		user := &models.User{Base: models.Base{ID: "u-1"}}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("GET", "/api/dashboard/top", nil)
		
		controller.GetTopStats(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetTrend - Custom Currency", func(t *testing.T) {
		mockService := new(services.MockStatsService)
		controller := NewDashboardController(mockService, nil)
		user := &models.User{Base: models.Base{ID: "u-1"}, BaseCurrency: "UAH"}
		
		mockService.On("GetTrendStats", user, "income", "EUR", mock.Anything, mock.Anything, mock.Anything).
			Return([]repositories.TrendStat{{Date: "2023-01-01", Total: 500}}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("GET", "/api/dashboard/trend?type=income&currency=EUR", nil)
		
		controller.GetTrend(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetRecent", func(t *testing.T) {
		mockService := new(services.MockStatsService)
		controller := NewDashboardController(mockService, nil)
		user := &models.User{Base: models.Base{ID: "u-1"}}
		
		mockService.On("GetRecentTransactions", user, []string{"acc1"}).
			Return([]models.Transaction{{Base: models.Base{ID: "tx1"}}}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("GET", "/api/dashboard/recent?accountIds=acc1", nil)
		
		controller.GetRecent(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []models.Transaction
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Len(t, resp, 1)
	})
}
