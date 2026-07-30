package middlewares

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/gin-gonic/gin"
)

// RequirePlatformAdmin protects operations that affect every WeGaS family.
// It intentionally does not use the family's admin/child role.
func RequirePlatformAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := c.Get("user")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		currentUser, ok := user.(*models.User)
		if !ok || !currentUser.IsPlatformAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Platform administrator access required"})
			return
		}

		c.Next()
	}
}
