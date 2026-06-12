package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/database"
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secretKey := "test-secret-key"

	// Setup test DB
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)
	database.DB = db

	// Create a test user
	testUser := models.User{
		Base:     models.Base{ID: "test-user-id"},
		FamilyID: "test-family-id",
		RoleID:   "admin",
		Email:    "test@example.com",
	}
	db.Create(&testUser)

	t.Run("Missing token", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		r.Use(AuthMiddleware(secretKey))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization token is missing")
	})

	t.Run("Invalid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		r.Use(AuthMiddleware(secretKey))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired token")
	})

	t.Run("Valid token", func(t *testing.T) {
		token, err := utils.GenerateToken(testUser.ID, testUser.FamilyID, testUser.RoleID, secretKey)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		r.Use(AuthMiddleware(secretKey))
		r.GET("/test", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			familyID, _ := c.Get("familyID")
			assert.Equal(t, testUser.ID, userID)
			assert.Equal(t, testUser.FamilyID, familyID)
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Token in query param (WebSocket support)", func(t *testing.T) {
		token, err := utils.GenerateToken(testUser.ID, testUser.FamilyID, testUser.RoleID, secretKey)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		r.Use(AuthMiddleware(secretKey))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test?token="+token, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("User not found in DB", func(t *testing.T) {
		token, err := utils.GenerateToken("non-existent-user", "family-id", "role-id", secretKey)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		r.Use(AuthMiddleware(secretKey))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})
}
