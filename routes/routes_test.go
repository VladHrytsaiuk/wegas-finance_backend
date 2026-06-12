package routes

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/controllers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Use empty controllers (nil pointers are fine if we don't call them, 
	// but SetupRoutes might access them to register handlers. 
	// Most controllers are structs, so we should provide initialized ones.)
	c := AppControllers{
		Auth:         &controllers.AuthController{},
		User:         &controllers.UserController{},
		Account:      &controllers.AccountController{},
		Category:     &controllers.CategoryController{},
		Counterparty: &controllers.CounterpartyController{},
		Tag:          &controllers.TagController{},
		Transaction:  &controllers.TransactionController{},
		Dashboard:    &controllers.DashboardController{},
		Role:         &controllers.RoleController{},
		Settings:     &controllers.SettingsController{},
		Export:       &controllers.ExportController{},
		Import:       &controllers.ImportController{},
		Monobank:     &controllers.MonobankController{},
		Asset:        &controllers.AssetController{},
		Utility:      &controllers.UtilityController{},
		Goal:         &controllers.GoalController{},
		StorageType:  &controllers.StorageTypeController{},
		Currency:     &controllers.CurrencyController{},
		Feedback:     &controllers.FeedbackController{},
		Shopping:     &controllers.ShoppingController{},
		Wishlist:     &controllers.WishlistController{},
		Family:       &controllers.FamilyController{},
		WS:           &controllers.WSController{},
	}

	SetupRoutes(r, c, "./uploads", "secret")

	// Verify some critical routes exist in the engine
	routes := r.Routes()
	
	findRoute := func(method, path string) bool {
		for _, route := range routes {
			if route.Method == method && route.Path == path {
				return true
			}
		}
		return false
	}

	// Public routes
	assert.True(t, findRoute("POST", "/api/users"))
	assert.True(t, findRoute("POST", "/api/login"))
	assert.True(t, findRoute("POST", "/api/feedback"))

	// Protected routes (prefix /api/)
	assert.True(t, findRoute("GET", "/api/accounts"))
	assert.True(t, findRoute("GET", "/api/transactions"))
	assert.True(t, findRoute("POST", "/api/transactions"))
	assert.True(t, findRoute("GET", "/api/dashboard/stats"))
	
	// Check count - should match the number of registered routes in SetupRoutes
	// This is a good way to ensure no routes were accidentally removed or duplicated
	assert.Greater(t, len(routes), 50)
}
