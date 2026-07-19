package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAccountRepo_Integration(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	repo := NewAccountRepository(db)

	familyID := uuid.NewString()
	userID := uuid.NewString()

	t.Run("TestCreateAccount", func(t *testing.T) {
		account := &models.Account{
			Base:     models.Base{ID: uuid.NewString()},
			FamilyID: familyID,
			UserID:   userID,
			Name:     "Test Account",
			Type:     "cash",
			Currency: "USD",
			Balance:  100,
		}

		err := repo.Create(account)
		assert.NoError(t, err)

		var saved models.Account
		db.First(&saved, "id = ?", account.ID)
		assert.Equal(t, "Test Account", saved.Name)
		assert.Equal(t, int64(100), saved.Balance)
	})

	t.Run("TestGetAllByFamilyID", func(t *testing.T) {
		accounts, err := repo.GetAllByFamilyID(familyID)
		assert.NoError(t, err)
		assert.NotEmpty(t, accounts)
		assert.Equal(t, familyID, accounts[0].FamilyID)
	})

	t.Run("TestGetByID", func(t *testing.T) {
		// Create another one
		id := uuid.NewString()
		db.Create(&models.Account{
			Base:     models.Base{ID: id},
			FamilyID: familyID,
			Name:     "By ID",
		})

		acc, err := repo.GetByID(id)
		assert.NoError(t, err)
		assert.Equal(t, "By ID", acc.Name)
	})

	t.Run("TestUpdateAccount", func(t *testing.T) {
		id := uuid.NewString()
		account := &models.Account{
			Base:     models.Base{ID: id},
			FamilyID: familyID,
			Name:     "Original Name",
			Balance:  500,
		}
		db.Create(account)

		account.Name = "Updated Name"
		account.Balance = 600
		err := repo.Update(account)
		assert.NoError(t, err)

		var updated models.Account
		db.First(&updated, "id = ?", id)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, int64(600), updated.Balance)
	})

	t.Run("TestUpdateAccountCardNumbers", func(t *testing.T) {
		id := uuid.NewString()
		account := &models.Account{
			Base:        models.Base{ID: id},
			FamilyID:    familyID,
			Name:        "Payment card",
			Type:        "card",
			CardNumber:  "1234",
			CardNumbers: []string{"1234", "5678"},
		}
		assert.NoError(t, repo.Create(account))

		account.CardNumbers = []string{"1234", "5678", "9012"}
		assert.NoError(t, repo.Update(account))

		accounts, err := repo.GetAllByFamilyID(familyID)
		assert.NoError(t, err)

		var saved *models.Account
		for index := range accounts {
			if accounts[index].ID == id {
				saved = &accounts[index]
				break
			}
		}
		if assert.NotNil(t, saved) {
			assert.Equal(t, []string{"1234", "5678", "9012"}, saved.CardNumbers)
		}
	})

	t.Run("TestDeleteAccount", func(t *testing.T) {
		id := uuid.NewString()
		db.Create(&models.Account{
			Base:     models.Base{ID: id},
			FamilyID: familyID,
			Name:     "To Delete",
		})

		err := repo.Delete(id)
		assert.NoError(t, err)

		var deleted models.Account
		err = db.First(&deleted, "id = ?", id).Error
		// It's a soft delete by updating deleted_at in repo.Delete
		assert.NoError(t, err)
		assert.NotNil(t, deleted.DeletedAt)
	})
}
