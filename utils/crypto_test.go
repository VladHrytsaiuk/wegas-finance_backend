package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrypto(t *testing.T) {
	password := "Secret123!"

	t.Run("Hash and Check Success", func(t *testing.T) {
		hash, err := HashPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)

		assert.True(t, CheckPassword(password, hash))
	})

	t.Run("Check Failure", func(t *testing.T) {
		hash, _ := HashPassword(password)
		assert.False(t, CheckPassword("WrongPassword", hash))
	})
}
