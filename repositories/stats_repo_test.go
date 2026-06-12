package repositories

import (
	"testing"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestStatsRepository(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	repo := NewStatsRepository(db)

	familyID := "family-1"
	userID := "user-1"
	currency := "UAH"

	// Create test data
	account := models.Account{
		Base:     models.Base{ID: "account-1"},
		Name:     "Test Account",
		Currency: currency,
		FamilyID: familyID,
		UserID:   userID,
	}
	db.Create(&account)

	category := models.Category{
		Base:     models.Base{ID: "cat-1"},
		Name:     "Food",
		Color:    "#FF0000",
		FamilyID: familyID,
	}
	db.Create(&category)

	counterparty := models.Counterparty{
		Base:     models.Base{ID: "cp-1"},
		Name:     "Supermarket",
		FamilyID: familyID,
	}
	db.Create(&counterparty)

	tag := models.Tag{
		Base:     models.Base{ID: "tag-1"},
		Name:     "Essential",
		Color:    "#00FF00",
		FamilyID: familyID,
	}
	db.Create(&tag)

	now := time.Now().UnixMilli()
	
	transactions := []models.Transaction{
		{
			Base:           models.Base{ID: "tx-1"},
			Amount:         100,
			Type:           "expense",
			Date:           now,
			FamilyID:       familyID,
			UserID:         userID,
			AccountID:      account.ID,
			CategoryID:     category.ID,
			CounterpartyID: counterparty.ID,
		},
		{
			Base:           models.Base{ID: "tx-2"},
			Amount:         200,
			Type:           "expense",
			Date:           now - 86400000, // Yesterday
			FamilyID:       familyID,
			UserID:         userID,
			AccountID:      account.ID,
			CategoryID:     category.ID,
			CounterpartyID: counterparty.ID,
		},
	}
	for _, tx := range transactions {
		db.Create(&tx)
	}
	
	// Add tag to tx-1
	db.Create(&models.TransactionTag{
		Base:          models.Base{ID: "tx-tag-1"},
		TransactionID: "tx-1",
		TagID:         tag.ID,
	})

	t.Run("GetBalances", func(t *testing.T) {
		balances, err := repo.GetBalances(familyID, "", nil)
		assert.NoError(t, err)
		assert.Len(t, balances, 1)
		assert.Equal(t, account.ID, balances[0].ID)
	})

	t.Run("GetTopFlow - Category", func(t *testing.T) {
		results, err := repo.GetTopFlow(familyID, "", "expense", "category", now-2*86400000, now+86400000, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, "Food", results[0].Name)
		assert.Equal(t, int64(300), results[0].Total)
	})

	t.Run("GetTopFlow - Counterparty", func(t *testing.T) {
		results, err := repo.GetTopFlow(familyID, "", "expense", "counterparty", now-2*86400000, now+86400000, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, "Supermarket", results[0].Name)
		assert.Equal(t, int64(300), results[0].Total)
	})

	t.Run("GetTopFlow - Tag", func(t *testing.T) {
		results, err := repo.GetTopFlow(familyID, "", "expense", "tag", now-2*86400000, now+86400000, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, "Essential", results[0].Name)
		assert.Equal(t, int64(100), results[0].Total)
	})

	t.Run("GetTrend", func(t *testing.T) {
		results, err := repo.GetTrend(familyID, "", "expense", now-2*86400000, now+86400000, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("GetRecentTransactions", func(t *testing.T) {
		results, err := repo.GetRecentTransactions(familyID, "", 10, nil)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("GetTotalSumByCurrency", func(t *testing.T) {
		results, err := repo.GetTotalSumByCurrency(familyID, "", "expense", now-2*86400000, now+86400000, nil)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, currency, results[0].Currency)
		assert.Equal(t, int64(300), results[0].Total)
	})
}
