package utils

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	secretKey := "test-secret-key"
	userID := "user-123"
	familyID := "family-456"
	roleID := "role-789"

	tokenString, err := GenerateToken(userID, familyID, roleID, secretKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims["user_id"])
	assert.Equal(t, familyID, claims["family_id"])
	assert.Equal(t, roleID, claims["role_id"])
	assert.NotNil(t, claims["exp"])
}

func TestGenerateToken_DifferentSecrets(t *testing.T) {
	userID := "user-123"
	familyID := "family-456"
	roleID := "role-789"

	tokenString, err := GenerateToken(userID, familyID, roleID, "secret1")
	assert.NoError(t, err)

	// Try to validate with wrong secret
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret2"), nil
	})

	assert.Error(t, err)
	assert.False(t, token.Valid)
}
