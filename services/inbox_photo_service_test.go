package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInboxServiceCreatePhotoRequiresSyncedAccount(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	user := &models.User{Base: models.Base{ID: uuid.NewString()}, FamilyID: uuid.NewString()}
	total := int64(12500)

	createAccount := func(isSynced bool) models.Account {
		account := models.Account{
			Base:     models.Base{ID: uuid.NewString(), IsSynced: isSynced},
			FamilyID: user.FamilyID,
			UserID:   user.ID,
			Name:     "Картка",
			Currency: "UAH",
		}
		require.NoError(t, db.Create(&account).Error)
		return account
	}

	t.Run("creates inbox entry for synced account", func(t *testing.T) {
		account := createAccount(true)
		service := NewInboxService(repositories.NewInboxRepository(db), db)
		entry, err := service.CreatePhoto(InboxCreateInput{
			SelectedAccountID: &account.ID,
			FilePath:          "/uploads/receipts/check.jpg",
			Total:             &total,
		}, user)
		require.NoError(t, err)
		assert.Equal(t, models.InboxEntryStatusNeedsLink, entry.Status)
		assert.Equal(t, "manual_photo_waiting_for_bank_transaction", entry.Reason)
		assert.Equal(t, models.ReceiptSourceTypePhoto, entry.SourceType)
		assert.Equal(t, account.ID, *entry.SelectedAccountID)
	})

	t.Run("rejects non-synced account", func(t *testing.T) {
		account := createAccount(false)
		service := NewInboxService(repositories.NewInboxRepository(db), db)
		_, err := service.CreatePhoto(InboxCreateInput{
			SelectedAccountID: &account.ID,
			FilePath:          "/uploads/receipts/check.jpg",
			Total:             &total,
		}, user)
		require.Error(t, err)
	})
}
