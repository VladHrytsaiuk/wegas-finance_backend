package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken створює JWT токен
func GenerateToken(userID, familyID, roleID string, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   userID,
		"family_id": familyID, // Важливо для швидкого доступу
		"role_id":   roleID,   // Важливо для перевірки прав
		"exp":       time.Now().Add(time.Hour * 72).Unix(), // 3 дні
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}