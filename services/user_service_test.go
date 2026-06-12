package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestUserService(t *testing.T) {
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	wsHub := utils.NewWSHub()
	service := NewUserService(userRepo, wsHub, db)

	t.Run("GetMe", func(t *testing.T) {
		user := &models.User{
			Base:  models.Base{ID: "user-1"},
			Name:  "Test User",
			Email: "test@me.com",
		}
		db.Create(user)

		me, err := service.GetMe("user-1")
		assert.NoError(t, err)
		assert.Equal(t, "Test User", me.Name)
	})

	t.Run("AddMember - Parent permission", func(t *testing.T) {
		parent := &models.User{
			Base:     models.Base{ID: "parent-1"},
			FamilyID: "fam-1",
			RoleID:   "admin",
		}
		
		input := CreateUserInput{
			Name:     "Child User",
			Email:    "child@test.com",
			Password: "password",
			RoleID:   "child",
		}

		newUser, err := service.AddMember(parent, input)
		assert.NoError(t, err)
		assert.Equal(t, "Child User", newUser.Name)
		assert.Equal(t, "fam-1", newUser.FamilyID)
	})

	t.Run("AddMember - Child permission denied", func(t *testing.T) {
		child := &models.User{
			Base:     models.Base{ID: "child-1"},
			FamilyID: "fam-1",
			RoleID:   "child",
		}
		
		input := CreateUserInput{
			Name:     "Another Child",
			Email:    "another@test.com",
			Password: "password",
		}

		_, err := service.AddMember(child, input)
		assert.Error(t, err)
		assert.Equal(t, ErrUserPermission, err)
	})

	t.Run("UpdateProfile", func(t *testing.T) {
		user := &models.User{
			Base:  models.Base{ID: "user-update"},
			Name:  "Old Name",
			Email: "old@email.com",
		}
		db.Create(user)

		updated, err := service.UpdateProfile("user-update", "New Name", "new@email.com")
		assert.NoError(t, err)
		assert.Equal(t, "New Name", updated.Name)
		assert.Equal(t, "new@email.com", updated.Email)
	})

	t.Run("ChangePassword", func(t *testing.T) {
		hashed, _ := utils.HashPassword("old-pass")
		user := &models.User{
			Base:         models.Base{ID: "user-pass"},
			PasswordHash: hashed,
		}
		db.Create(user)

		err := service.ChangePassword("user-pass", "old-pass", "new-pass")
		assert.NoError(t, err)

		// Verify
		var dbUser models.User
		db.First(&dbUser, "id = ?", "user-pass")
		assert.True(t, utils.CheckPassword("new-pass", dbUser.PasswordHash))
	})

	t.Run("LeaveFamily", func(t *testing.T) {
		familyID := "fam-leave"
		db.Create(&models.Family{Base: models.Base{ID: familyID}})
		
		user1 := &models.User{Base: models.Base{ID: "u1"}, FamilyID: familyID, Name: "User 1", Email: "u1@test.com"}
		user2 := &models.User{Base: models.Base{ID: "u2"}, FamilyID: familyID, Name: "User 2", Email: "u2@test.com"}
		db.Create(user1)
		db.Create(user2)

		err := service.LeaveFamily(user1)
		assert.NoError(t, err)

		// Check user 1 has new family
		var updatedU1 models.User
		db.First(&updatedU1, "id = ?", "u1")
		assert.NotEqual(t, familyID, updatedU1.FamilyID)
		assert.Equal(t, "admin", updatedU1.RoleID)
	})

	t.Run("DeleteMember and Resource Cloning", func(t *testing.T) {
		familyID := "fam-clone"
		assert.NoError(t, db.Create(&models.Family{Base: models.Base{ID: familyID}}).Error)
		
		parent := &models.User{
			Base: models.Base{ID: "p1"}, 
			FamilyID: familyID, 
			RoleID: "admin", 
			Name: "Parent",
			Email: "parent@clone.com",
		}
		child := &models.User{
			Base: models.Base{ID: "c1"}, 
			FamilyID: familyID, 
			RoleID: "child", 
			Name: "Child",
			Email: "child@clone.com",
		}
		assert.NoError(t, db.Create(parent).Error)
		assert.NoError(t, db.Create(child).Error)

		// Create shared resources
		cat := &models.Category{Base: models.Base{ID: "cat-1"}, FamilyID: familyID, Name: "Food"}
		cp := &models.Counterparty{Base: models.Base{ID: "cp-1"}, FamilyID: familyID, Name: "Store"}
		assert.NoError(t, db.Create(cat).Error)
		assert.NoError(t, db.Create(cp).Error)

		// Create child's transaction
		tx := &models.Transaction{
			Base: models.Base{ID: "tx-child"},
			FamilyID: familyID,
			UserID: "c1",
			CategoryID: "cat-1",
			CounterpartyID: "cp-1",
		}
		assert.NoError(t, db.Create(tx).Error)

		err := service.DeleteMember(parent, "c1")
		assert.NoError(t, err)

		// Check child has new family
		var updatedC1 models.User
		db.First(&updatedC1, "id = ?", "c1")
		newFamID := updatedC1.FamilyID
		assert.NotEqual(t, familyID, newFamID)

		// Check resources were cloned
		var clonedCat models.Category
		err = db.Where("family_id = ? AND name = ?", newFamID, "Food").First(&clonedCat).Error
		assert.NoError(t, err)

		// Check transaction was re-linked
		var updatedTx models.Transaction
		db.First(&updatedTx, "id = ?", "tx-child")
		assert.Equal(t, newFamID, updatedTx.FamilyID)
		assert.Equal(t, clonedCat.ID, updatedTx.CategoryID)
	})
}
