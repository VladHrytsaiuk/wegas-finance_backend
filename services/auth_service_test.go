package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
)

func TestAuthService(t *testing.T) {
	secretKey := "test-secret"
	inviteCode := "invite-123"

	t.Run("Register Success", func(t *testing.T) {
		db, _ := repositories.SetupTestDB()
		userRepo := repositories.NewUserRepository(db)
		service := NewAuthService(userRepo, secretKey, inviteCode)

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
		service := NewAuthService(userRepo, secretKey, inviteCode)

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
		service := NewAuthService(userRepo, secretKey, inviteCode)

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
		service := NewAuthService(userRepo, secretKey, inviteCode)

		loginInput := LoginInput{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}
		_, err := service.Login(loginInput)
		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
	})
}
