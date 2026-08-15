package database

import (
	"path/filepath"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRunDataMigrationsRunsEachMigrationOnce(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "finance.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)

	runs := 0
	migrations := []DataMigration{{
		ID: "test_001",
		Up: func(tx *gorm.DB) error {
			runs++
			return tx.Exec("CREATE TABLE IF NOT EXISTS migration_test_values (id INTEGER PRIMARY KEY)").Error
		},
	}}

	require.NoError(t, RunDataMigrations(db, dbPath, migrations))
	require.Equal(t, 1, runs)
	require.DirExists(t, filepath.Join(filepath.Dir(dbPath), "backups"))

	require.NoError(t, RunDataMigrations(db, dbPath, migrations))
	require.Equal(t, 1, runs)
}

func TestBackfillBatchImportAccountBalances(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Account{}, &models.Transaction{}))

	account := models.Account{
		Base:     models.Base{ID: uuid.NewString()},
		FamilyID: "family-1",
		Name:     "USD card",
		Currency: "USD",
		Balance:  0,
	}
	require.NoError(t, db.Create(&account).Error)
	require.NoError(t, db.Create(&models.Transaction{
		Base:      models.Base{ID: uuid.NewString()},
		FamilyID:  account.FamilyID,
		AccountID: account.ID,
		Type:      "expense",
		Amount:    1250,
	}).Error)
	require.NoError(t, db.Create(&models.Transaction{
		Base:      models.Base{ID: uuid.NewString()},
		FamilyID:  account.FamilyID,
		AccountID: account.ID,
		Type:      "income",
		Amount:    2000,
	}).Error)

	require.NoError(t, backfillBatchImportAccountBalances(db))

	var updatedAccount models.Account
	require.NoError(t, db.First(&updatedAccount, "id = ?", account.ID).Error)
	require.Equal(t, int64(750), updatedAccount.Balance)

	var transactions []models.Transaction
	require.NoError(t, db.Where("account_id = ?", account.ID).Find(&transactions).Error)
	require.Len(t, transactions, 2)
	for _, transaction := range transactions {
		require.Equal(t, "USD", transaction.Currency)
	}

	// Once the currency is backfilled, a repeat run cannot change the balance again.
	require.NoError(t, backfillBatchImportAccountBalances(db))
	require.NoError(t, db.First(&updatedAccount, "id = ?", account.ID).Error)
	require.Equal(t, int64(750), updatedAccount.Balance)
}
