package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	assert.Equal(t, "test_value", getEnv("TEST_KEY", "default"))
	assert.Equal(t, "default", getEnv("NON_EXISTENT", "default"))
}

func TestLoadConfig(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("JWT_SECRET", "super-secret")
	os.Setenv("DB_PATH", "test.db")
	defer os.Unsetenv("SERVER_PORT")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("DB_PATH")

	cfg := LoadConfig()

	assert.Equal(t, "9090", cfg.ServerPort)
	assert.Equal(t, "super-secret", cfg.SecretKey)
	assert.Equal(t, "test.db", cfg.DBPath)
}
