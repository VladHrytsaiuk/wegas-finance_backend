package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestCategoryRepository(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	repo := NewCategoryRepository(db)
	familyID := "fam-cats"

	t.Run("Create and Get Category", func(t *testing.T) {
		cat := &models.Category{
			Base:     models.Base{ID: "cat-1"},
			FamilyID: familyID,
			Name:     "Food",
			Type:     "expense",
		}
		err := repo.Create(cat)
		assert.NoError(t, err)

		saved, err := repo.GetByID("cat-1", familyID)
		assert.NoError(t, err)
		assert.Equal(t, "Food", saved.Name)
	})

	t.Run("GetAll", func(t *testing.T) {
		cats, err := repo.GetAll(familyID)
		assert.NoError(t, err)
		assert.Len(t, cats, 1)
	})

	t.Run("Update Category", func(t *testing.T) {
		cat, _ := repo.GetByID("cat-1", familyID)
		cat.Name = "Groceries"
		err := repo.Update(cat)
		assert.NoError(t, err)

		updated, _ := repo.GetByID("cat-1", familyID)
		assert.Equal(t, "Groceries", updated.Name)
	})

	t.Run("Delete Category", func(t *testing.T) {
		err := repo.Delete("cat-1", familyID)
		assert.NoError(t, err)

		// Verify soft delete
		var saved models.Category
		db.Unscoped().First(&saved, "id = ?", "cat-1")
		assert.NotNil(t, saved.DeletedAt)
	})
}
