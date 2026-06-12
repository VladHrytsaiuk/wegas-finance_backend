package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestAssetRepository(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	repo := NewAssetRepository(db)
	familyID := "fam-assets"
	userID := "u-assets"

	t.Run("Create and Get Asset", func(t *testing.T) {
		asset := &models.Asset{
			Base:     models.Base{ID: "asset-1"},
			FamilyID: familyID,
			UserID:   userID,
			Name:     "Laptop",
			Type:     "electronics",
			Price:    50000,
		}
		err := repo.Create(asset)
		assert.NoError(t, err)

		saved, err := repo.GetByID("asset-1", familyID)
		assert.NoError(t, err)
		assert.Equal(t, "Laptop", saved.Name)
	})

	t.Run("GetAll", func(t *testing.T) {
		assets, err := repo.GetAll(familyID)
		assert.NoError(t, err)
		assert.Len(t, assets, 1)
	})

	t.Run("Update Asset", func(t *testing.T) {
		asset, _ := repo.GetByID("asset-1", familyID)
		asset.Name = "New Laptop"
		err := repo.Update(asset)
		assert.NoError(t, err)

		updated, _ := repo.GetByID("asset-1", familyID)
		assert.Equal(t, "New Laptop", updated.Name)
	})

	t.Run("Delete Asset", func(t *testing.T) {
		err := repo.Delete("asset-1", familyID)
		assert.NoError(t, err)

		_, err = repo.GetByID("asset-1", familyID)
		assert.Error(t, err)
	})
}
