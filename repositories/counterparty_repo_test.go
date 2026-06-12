package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestCounterpartyRepository(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	repo := NewCounterpartyRepository(db)
	familyID := "fam-1"

	t.Run("Create and Get Category", func(t *testing.T) {
		cat := &models.CounterpartyCategory{
			Base:     models.Base{ID: "cat-1"},
			FamilyID: familyID,
			Name:     "Shops",
		}
		err := repo.CreateCategory(cat)
		assert.NoError(t, err)

		saved, err := repo.GetCategoryByID("cat-1", familyID)
		assert.NoError(t, err)
		assert.Equal(t, "Shops", saved.Name)
	})

	t.Run("Create and Get Counterparty", func(t *testing.T) {
		cp := &models.Counterparty{
			Base:     models.Base{ID: "cp-1"},
			FamilyID: familyID,
			Name:     "Silpo",
			Balances: []models.CounterpartyBalance{
				{Currency: "UAH", Balance: 0},
			},
		}
		err := repo.Create(cp)
		assert.NoError(t, err)

		saved, err := repo.GetByID("cp-1", familyID)
		assert.NoError(t, err)
		assert.Equal(t, "Silpo", saved.Name)
		assert.Len(t, saved.Balances, 1)
	})

	t.Run("GetByName Case Insensitive", func(t *testing.T) {
		// Existing 'Silpo' was created in previous test
		_, err := repo.GetByName("Silpo", familyID)
		assert.NoError(t, err)

		// Create another one with lower case to test matching
		cpLow := &models.Counterparty{
			Base:     models.Base{ID: "cp-low"},
			FamilyID: familyID,
			Name:     "atb",
		}
		repo.Create(cpLow)

		_, err = repo.GetByName("atb", familyID)
		assert.NoError(t, err)

		_, err = repo.GetByName("ATB", familyID)
		assert.NoError(t, err)
	})

	t.Run("Delete with active debt should fail", func(t *testing.T) {
		cp := &models.Counterparty{
			Base:     models.Base{ID: "cp-debt"},
			FamilyID: familyID,
			Name:     "Debtor",
		}
		db.Create(cp)
		db.Create(&models.CounterpartyBalance{
			CounterpartyID: "cp-debt",
			Currency:       "UAH",
			Balance:        50000, // 500.00
		})

		err := repo.Delete("cp-debt", familyID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "active debt")
	})

	t.Run("Delete without debt should succeed", func(t *testing.T) {
		cp := &models.Counterparty{
			Base:     models.Base{ID: "cp-clean"},
			FamilyID: familyID,
			Name:     "Clean",
		}
		db.Create(cp)

		err := repo.Delete("cp-clean", familyID)
		assert.NoError(t, err)

		// Verify soft delete
		var saved models.Counterparty
		db.Unscoped().First(&saved, "id = ?", "cp-clean")
		assert.NotNil(t, saved.DeletedAt)
	})
}
