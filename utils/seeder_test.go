package utils

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
)

func TestSeeders(t *testing.T) {
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)

	familyID := "test-family-id"

	t.Run("SeedFamilyDefaults", func(t *testing.T) {
		err := SeedFamilyDefaults(db, familyID)
		assert.NoError(t, err)

		var count int64
		db.Model(&models.Category{}).Where("family_id = ?", familyID).Count(&count)
		assert.Greater(t, count, int64(0))
	})

	t.Run("SeedDefaultCounterparties", func(t *testing.T) {
		err := SeedDefaultCounterparties(db, familyID)
		assert.NoError(t, err)

		var cpCount int64
		db.Model(&models.Counterparty{}).Where("family_id = ?", familyID).Count(&cpCount)
		assert.Greater(t, cpCount, int64(0))

		var catCount int64
		db.Model(&models.CounterpartyCategory{}).Where("family_id = ?", familyID).Count(&catCount)
		assert.Greater(t, catCount, int64(0))
	})

	t.Run("SeedSystemStorageTypes", func(t *testing.T) {
		err := SeedSystemStorageTypes(db)
		assert.NoError(t, err)

		var count int64
		db.Model(&models.StorageType{}).Where("family_id IS NULL AND is_system = ?", true).Count(&count)
		assert.Greater(t, count, int64(0))
	})
}
