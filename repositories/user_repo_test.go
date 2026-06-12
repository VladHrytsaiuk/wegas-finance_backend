package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	repo := NewUserRepository(db)

	t.Run("Create and Get User", func(t *testing.T) {
		user := &models.User{
			Base:  models.Base{ID: "u-1"},
			Name:  "User One",
			Email: "u1@test.com",
		}
		err := repo.Create(user)
		assert.NoError(t, err)

		saved, err := repo.GetByEmail("u1@test.com")
		assert.NoError(t, err)
		assert.Equal(t, "User One", saved.Name)

		savedByID, err := repo.GetByID("u-1")
		assert.NoError(t, err)
		assert.Equal(t, "u1@test.com", savedByID.Email)
	})

	t.Run("Update User", func(t *testing.T) {
		user, _ := repo.GetByID("u-1")
		user.Name = "Updated Name"
		err := repo.Update(user)
		assert.NoError(t, err)

		updated, _ := repo.GetByID("u-1")
		assert.Equal(t, "Updated Name", updated.Name)
	})

	t.Run("Family Members", func(t *testing.T) {
		familyID := "fam-users"
		db.Create(&models.User{Base: models.Base{ID: "f-1"}, FamilyID: familyID, Email: "f1@test.com"})
		db.Create(&models.User{Base: models.Base{ID: "f-2"}, FamilyID: familyID, Email: "f2@test.com"})

		members, err := repo.GetFamilyMembers(familyID)
		assert.NoError(t, err)
		assert.Len(t, members, 2)

		count, err := repo.CountFamilyMembers(familyID)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("Soft Delete User", func(t *testing.T) {
		err := repo.Delete("u-1")
		assert.NoError(t, err)

		_, err = repo.GetByID("u-1")
		assert.Error(t, err) // Should not find deleted user
	})

	t.Run("Family operations", func(t *testing.T) {
		fam := &models.Family{Base: models.Base{ID: "new-fam"}, Name: "New Fam"}
		err := repo.CreateFamily(fam)
		assert.NoError(t, err)

		saved, err := repo.GetFamilyByID("new-fam")
		assert.NoError(t, err)
		assert.Equal(t, "New Fam", saved.Name)

		err = repo.DeleteFamily("new-fam")
		assert.NoError(t, err)
	})
}
