package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/gin-gonic/gin"
)

// GetTestGinContext creates a gin context for testing
func GetTestGinContext(w *httptest.ResponseRecorder) *gin.Context {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(w)
	return ctx
}

// SetupTestUser sets an authenticated user in the gin context
func SetupTestUser(ctx *gin.Context, userID, familyID string) {
	user := &models.User{
		Base: models.Base{
			ID: userID,
		},
		FamilyID: familyID,
		Name:     "Test User",
		Email:    "test@example.com",
		RoleID:   "admin",
	}
	ctx.Set("user", user)
	ctx.Set("userID", userID)
	ctx.Set("familyID", familyID)
}

// PerformRequest is a helper to perform a request for testing
func PerformRequest(r http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
