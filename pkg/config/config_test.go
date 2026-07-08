package config

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	assert.Equal(t, "test_value", getEnv("TEST_KEY", "default"))
	assert.Equal(t, "default", getEnv("NON_EXISTENT", "default"))
}

func TestMustGetEnv(t *testing.T) {
	os.Setenv("REQUIRED_KEY", "required_value")
	defer os.Unsetenv("REQUIRED_KEY")

	assert.Equal(t, "required_value", mustGetEnv("REQUIRED_KEY"))
}

func TestSanitizePath(t *testing.T) {
	assert.Equal(t, "./uploads", sanitizePath("", "./uploads"))

	if runtime.GOOS == "windows" {
		assert.Equal(t, `C:\WegasFinance\uploads`, sanitizePath(`C:\WegasFinance\uploads`, "./uploads"))
		return
	}

	assert.Equal(t, "./uploads", sanitizePath(`C:\WegasFinance\uploads`, "./uploads"))
	assert.Equal(t, "./uploads", sanitizePath("C:/WegasFinance/uploads", "./uploads"))
	assert.Equal(t, "/Users/test/uploads", sanitizePath("/Users/test/uploads", "./uploads"))
}

func TestBuildAllowedOrigins(t *testing.T) {
	t.Run("uses explicit env origins", func(t *testing.T) {
		origins := buildAllowedOrigins(
			"https://wegas-finance.vercel.app, http://192.168.0.10:5173/path , invalid",
			"http://localhost:3000",
		)

		assert.Equal(t, []string{
			"https://wegas-finance.vercel.app",
			"http://192.168.0.10:5173",
		}, origins)
	})

	t.Run("falls back to defaults", func(t *testing.T) {
		origins := buildAllowedOrigins("", "https://wegas-finance.vercel.app/app")

		assert.Contains(t, origins, "https://wegas-finance.vercel.app")
		assert.Contains(t, origins, "http://localhost:5173")
		assert.Contains(t, origins, "http://127.0.0.1:5173")
	})
}

func TestLoadConfig(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("JWT_SECRET", "super-secret")
	os.Setenv("DB_PATH", "test.db")
	os.Setenv("UPLOADS_DIR", "./test-uploads")
	os.Setenv("REGISTRATION_CODE", "invite-code")
	os.Setenv("APP_URL", "https://wegas-finance.vercel.app/app")
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://wegas-finance.vercel.app,http://192.168.0.10:5173")
	defer os.Unsetenv("SERVER_PORT")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("DB_PATH")
	defer os.Unsetenv("UPLOADS_DIR")
	defer os.Unsetenv("REGISTRATION_CODE")
	defer os.Unsetenv("APP_URL")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	cfg := LoadConfig()

	assert.Equal(t, "9090", cfg.ServerPort)
	assert.Equal(t, "super-secret", cfg.SecretKey)
	assert.Equal(t, "test.db", cfg.DBPath)
	assert.Equal(t, "./test-uploads", cfg.UploadsDir)
	assert.Equal(t, "invite-code", cfg.RegistrationCode)
	assert.Equal(t, "https://wegas-finance.vercel.app/app", cfg.AppURL)
	assert.Equal(t, []string{
		"https://wegas-finance.vercel.app",
		"http://192.168.0.10:5173",
	}, cfg.AllowedOrigins)
}
