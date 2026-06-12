package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestStorageTypeRepository(t *testing.T) {
	db, _ := SetupTestDB()
	repo := NewStorageTypeRepository(db)
	familyID := "fam-st"

	t.Run("Create and FindAvailable", func(t *testing.T) {
		st := &models.StorageType{
			Base: models.Base{ID: "st-1"},
			FamilyID: &familyID,
			Name: "Safe",
		}
		err := repo.Create(st)
		assert.NoError(t, err)

		// Create a system type
		db.Create(&models.StorageType{
			Base: models.Base{ID: "sys-1"},
			IsSystem: true,
			Name: "Bank",
		})

		types, err := repo.FindAvailable(familyID)
		assert.NoError(t, err)
		assert.Len(t, types, 2)
	})

	t.Run("FindBySlug", func(t *testing.T) {
		db.Create(&models.StorageType{
			Base: models.Base{ID: "slug-1"},
			Slug: "cash",
			Name: "Cash",
		})
		
		st, err := repo.FindBySlug("cash")
		assert.NoError(t, err)
		assert.Equal(t, "Cash", st.Name)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete("st-1")
		assert.NoError(t, err)

		types, _ := repo.FindAvailable(familyID)
		// Only system type left
		assert.Len(t, types, 1)
		var count int64
		db.Model(&models.StorageType{}).Where("id = ?", "st-1").Count(&count)
		assert.Equal(t, int64(0), count)
	})
}
