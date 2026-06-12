package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestFamilyJoinService(t *testing.T) {
	t.Run("Generate and Join", func(t *testing.T) {
		db, _ := repositories.SetupTestDB()
		userRepo := repositories.NewUserRepository(db)
		joinRepo := repositories.NewFamilyJoinRepository(db)
		wsHub := utils.NewWSHub()
		service := NewFamilyJoinService(joinRepo, userRepo, wsHub, db)

		// Create family and admin
		family := &models.Family{Base: models.Base{ID: "fam-1"}, Name: "New Family"}
		db.Create(family)
		
		admin := &models.User{
			Base: models.Base{ID: "admin-1"},
			FamilyID: "fam-1",
			RoleID: "admin",
			Email: "admin@test.com",
		}
		db.Create(admin)

		// Create user joining
		user := &models.User{
			Base: models.Base{ID: "user-1"},
			FamilyID: "fam-old",
			RoleID: "admin",
			Email: "user@test.com",
		}
		db.Create(&models.Family{Base: models.Base{ID: "fam-old"}})
		db.Create(user)

		code, err := service.GenerateCode("fam-1", "member")
		assert.NoError(t, err)
		assert.Len(t, code, 6)

		joinedFamily, err := service.JoinFamily("user-1", code)
		assert.NoError(t, err)
		assert.Equal(t, "fam-1", joinedFamily.ID)

		// Check user updated
		var updatedUser models.User
		db.First(&updatedUser, "id = ?", "user-1")
		assert.Equal(t, "fam-1", updatedUser.FamilyID)
		assert.Equal(t, "member", updatedUser.RoleID)
	})

	t.Run("Join with invalid code", func(t *testing.T) {
		db, _ := repositories.SetupTestDB()
		userRepo := repositories.NewUserRepository(db)
		joinRepo := repositories.NewFamilyJoinRepository(db)
		wsHub := utils.NewWSHub()
		service := NewFamilyJoinService(joinRepo, userRepo, wsHub, db)

		user := &models.User{
			Base: models.Base{ID: "user-1"},
			Email: "user@test.com",
		}
		db.Create(user)

		_, err := service.JoinFamily("user-1", "000000")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired code")
	})
}
