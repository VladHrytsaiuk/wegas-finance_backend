package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptionDecryption(t *testing.T) {
	// Set a custom secret key for testing
	os.Setenv("APP_SECRET_KEY", "thisis32byteslongsecretkeytestok")
	defer os.Unsetenv("APP_SECRET_KEY")

	tests := []struct {
		name string
		text string
	}{
		{
			name: "Simple text",
			text: "hello world",
		},
		{
			name: "Empty string",
			text: "",
		},
		{
			name: "Long text",
			text: "This is a much longer text to ensure that encryption and decryption work correctly with more data than just a few characters.",
		},
		{
			name: "Special characters",
			text: "!@#$%^&*()_+|~-=`{}[]:;'<>?,./",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(tt.text)
			assert.NoError(t, err)
			assert.NotEmpty(t, encrypted)
			assert.NotEqual(t, tt.text, encrypted)

			decrypted, err := Decrypt(encrypted)
			assert.NoError(t, err)
			assert.Equal(t, tt.text, decrypted)
		})
	}
}

func TestDecrypt_InvalidInput(t *testing.T) {
	os.Setenv("APP_SECRET_KEY", "thisis32byteslongsecretkeytestok")
	defer os.Unsetenv("APP_SECRET_KEY")

	tests := []struct {
		name       string
		cryptoText string
	}{
		{
			name:       "Invalid base64",
			cryptoText: "invalid-base64-content!",
		},
		{
			name:       "Too short",
			cryptoText: "YWJj", // "abc" in base64, too short for GCM nonce
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := Decrypt(tt.cryptoText)
			assert.Error(t, err)
			assert.Empty(t, decrypted)
		})
	}
}
