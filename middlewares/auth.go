package middlewares

import (
	"net/http"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/database"
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// AuthMiddleware приймає secretKey як аргумент!
func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""
		authHeader := c.GetHeader("Authorization")

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// Якщо в хедері немає, шукаємо в Query (для WebSocket)
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is missing"})
			return
		}

		// Використовуємо переданий secretKey
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		var user models.User

		// Використовуємо database.DB (ініціалізований в main.go)
		if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				return
			}
			// Якщо база заблокована або інша помилка - 500
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
			return
		}

		c.Set("user", &user)
		c.Set("userID", user.ID)
		c.Set("familyID", user.FamilyID)
		c.Set("roleID", user.RoleID)

		c.Next()
	}
}