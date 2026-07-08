package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(CORSMiddleware([]string{
			"https://wegas-finance.vercel.app",
			"http://localhost:5173",
		}))
		r.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		r.OPTIONS("/ping", func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})
		return r
	}

	t.Run("allows configured origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		resp := httptest.NewRecorder()

		newRouter().ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "http://localhost:5173", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", resp.Header().Get("Vary"))
	})

	t.Run("rejects unknown origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Origin", "https://evil.example.com")
		resp := httptest.NewRecorder()

		newRouter().ServeHTTP(resp, req)

		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("handles preflight for configured origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
		req.Header.Set("Origin", "https://wegas-finance.vercel.app")
		resp := httptest.NewRecorder()

		newRouter().ServeHTTP(resp, req)

		assert.Equal(t, http.StatusNoContent, resp.Code)
		assert.Equal(t, "https://wegas-finance.vercel.app", resp.Header().Get("Access-Control-Allow-Origin"))
	})
}
