package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestTagRepository(t *testing.T) {
	db, _ := SetupTestDB()
	repo := NewTagRepository(db)
	familyID := "fam-tag"

	t.Run("Create and GetAll", func(t *testing.T) {
		tag := &models.Tag{
			Base: models.Base{ID: "tag-1"},
			FamilyID: familyID,
			Name: "Gift",
		}
		err := repo.Create(tag)
		assert.NoError(t, err)

		tags, err := repo.GetAll(familyID)
		assert.NoError(t, err)
		assert.Len(t, tags, 1)
	})

	t.Run("GetByID", func(t *testing.T) {
		tag, err := repo.GetByID("tag-1", familyID)
		assert.NoError(t, err)
		assert.Equal(t, "Gift", tag.Name)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete("tag-1", familyID)
		assert.NoError(t, err)

		tags, _ := repo.GetAll(familyID)
		assert.Len(t, tags, 0)
	})
}
