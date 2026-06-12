package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGoalController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockGoalService)
	controller := NewGoalController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "u-1", "f-1")
		c.Set("familyID", "f-1")
		c.Next()
	})

	r.POST("/goals", controller.Create)
	r.GET("/goals", controller.GetAll)
	r.GET("/goals/:id", controller.GetOne)

	t.Run("Create Success", func(t *testing.T) {
		mockService.On("Create", mock.MatchedBy(func(g *models.Goal) bool {
			return g.Name == "New Goal"
		}), "u-1").Return(nil).Once()

		body := map[string]string{"name": "New Goal"}
		w := PerformRequest(r, "POST", "/goals", body)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("GetAll", func(t *testing.T) {
		mockService.On("GetAll", "f-1", "u-1").Return([]models.Goal{{Name: "Goal 1"}}, nil).Once()

		w := PerformRequest(r, "GET", "/goals", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var goals []models.Goal
		json.Unmarshal(w.Body.Bytes(), &goals)
		assert.Len(t, goals, 1)
		mockService.AssertExpectations(t)
	})

	t.Run("GetOne Success", func(t *testing.T) {
		mockService.On("GetOne", "g-1", "u-1").Return(&models.Goal{Base: models.Base{ID: "g-1"}}, nil).Once()

		w := PerformRequest(r, "GET", "/goals/g-1", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
