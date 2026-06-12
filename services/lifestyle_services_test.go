package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
)

func stringPtr(s string) *string {
	return &s
}

func TestTagService(t *testing.T) {
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)

	repo := repositories.NewTagRepository(db)
	service := NewTagService(repo)

	user := &models.User{Base: models.Base{ID: "u-1"}, FamilyID: "fam-1"}

	t.Run("Create and Get Tags", func(t *testing.T) {
		tag, err := service.Create("Coffee", "#ff0000", user)
		assert.NoError(t, err)
		assert.Equal(t, "Coffee", tag.Name)

		tags, err := service.GetAll("fam-1")
		assert.NoError(t, err)
		assert.Len(t, tags, 1)
	})

	t.Run("Delete Tag", func(t *testing.T) {
		tags, _ := service.GetAll("fam-1")
		tagID := tags[0].ID

		err := service.Delete(tagID, user)
		assert.NoError(t, err)

		tagsAfter, _ := service.GetAll("fam-1")
		assert.Len(t, tagsAfter, 0)
	})
}

func TestRoleService(t *testing.T) {
	db, _ := repositories.SetupTestDB()
	repo := repositories.NewRoleRepository(db)
	service := NewRoleService(repo)

	t.Run("Manage Roles", func(t *testing.T) {
		role, err := service.Create(CreateRoleInput{
			Name: "Manager",
			Description: "Can manage things",
		})
		assert.NoError(t, err)
		assert.Equal(t, "Manager", role.Name)

		roles, err := service.GetAll()
		assert.NoError(t, err)
		assert.Len(t, roles, 1)

		err = service.Delete(role.ID)
		assert.NoError(t, err)
	})
}

func TestStorageTypeService(t *testing.T) {
	db, _ := repositories.SetupTestDB()
	repo := repositories.NewStorageTypeRepository(db)
	service := NewStorageTypeService(repo)

	t.Run("Manage Storage Types", func(t *testing.T) {
		st := &models.StorageType{
			FamilyID: stringPtr("fam-1"),
			Name: "Vault",
			Slug: "vault",
		}
		err := service.Create(st)
		assert.NoError(t, err)

		types, err := service.GetAll("fam-1")
		assert.NoError(t, err)
		assert.NotEmpty(t, types)

		err = service.Delete(st.ID)
		assert.NoError(t, err)
	})
}
