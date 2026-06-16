package services

import (
	"testing"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
)

func TestAuthService(t *testing.T) {
	secretKey := "test-secret"
	inviteCode := "invite-123"
	jwtService := NewJWTService(secretKey)

	t.Run("Register Success", func(t *testing.T) {
		db, _ := repositories.SetupTestDB()
		userRepo := repositories.NewUserRepository(db)
		service := NewAuthService(userRepo, jwtService, secretKey, inviteCode)

		input := RegisterInput{
			Name:       "Test User",
			Email:      "test@example.com",
			Password:   "password123",
			InviteCode: inviteCode,
		}

		resp, err := service.Register(input)
		assert.NoError(t, err)
		assert.NotNil(t, resp.Token)
		assert.Equal(t, "test@example.com", resp.User.Email)
		assert.Equal(t, "admin", resp.User.RoleID)
	})

	t.Run("Register Invalid Invite Code", func(t *testing.T) {
		db, _ := repositories.SetupTestDB()
		userRepo := repositories.NewUserRepository(db)
		service := NewAuthService(userRepo, jwtService, secretKey, inviteCode)

		input := RegisterInput{
			Name:       "Test User",
			Email:      "test@example.com",
			Password:   "password123",
			InviteCode: "wrong-code",
		}

		_, err := service.Register(input)
		assert.Error(t, err)
		assert.Equal(t, "invalid invite code", err.Error())
	})

	t.Run("Login Success", func(t *testing.T) {
		db, _ := repositories.SetupTestDB()
		userRepo := repositories.NewUserRepository(db)
		service := NewAuthService(userRepo, jwtService, secretKey, inviteCode)

		// Register first
		regInput := RegisterInput{
			Name:       "Test User",
			Email:      "test@example.com",
			Password:   "password123",
			InviteCode: inviteCode,
		}
		service.Register(regInput)

		// Login
		loginInput := LoginInput{
			Email:    "test@example.com",
			Password: "password123",
		}
		resp, err := service.Login(loginInput)
		assert.NoError(t, err)
		assert.NotNil(t, resp.Token)
		assert.Equal(t, "test@example.com", resp.User.Email)
	})

	t.Run("Login Invalid Credentials", func(t *testing.T) {
		db, _ := repositories.SetupTestDB()
		userRepo := repositories.NewUserRepository(db)
		service := NewAuthService(userRepo, jwtService, secretKey, inviteCode)

		loginInput := LoginInput{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}
		_, err := service.Login(loginInput)
		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("PIN Logic", func(t *testing.T) {
		db, _ := repositories.SetupTestDB()
		userRepo := repositories.NewUserRepository(db)
		service := NewAuthService(userRepo, jwtService, secretKey, inviteCode)

		regInput := RegisterInput{
			Name: "PIN User", Email: "pin@test.com", Password: "password123", InviteCode: inviteCode,
		}
		regResp, _ := service.Register(regInput)

		// 1. Set PIN
		err := service.SetPIN(regResp.User.ID, "1234")
		assert.NoError(t, err)

		// 2. Login with valid PIN
		resp, err := service.LoginWithPIN("pin@test.com", "1234")
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)

		// Wait for rate limiter (2 seconds)
		time.Sleep(2 * time.Second)

		// 3. Login with invalid PIN
		_, err = service.LoginWithPIN("pin@test.com", "0000")
		assert.Error(t, err)
		assert.Equal(t, "invalid PIN", err.Error())
	})
}
