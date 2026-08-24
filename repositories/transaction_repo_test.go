package repositories

import (
	"testing"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTransactionRepo_Integration(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	repo := NewTransactionRepository(db)

	// Setup initial data
	family := models.Family{
		Base: models.Base{ID: uuid.NewString()},
		Name: "Test Family",
	}
	db.Create(&family)

	user := models.User{
		Base:     models.Base{ID: uuid.NewString()},
		FamilyID: family.ID,
		Name:     "Test User",
		Email:    "test@example.com",
	}
	db.Create(&user)

	account1 := models.Account{
		Base:           models.Base{ID: uuid.NewString()},
		FamilyID:       family.ID,
		UserID:         user.ID,
		Name:           "Account 1",
		Currency:       "UAH",
		InitialBalance: 1000,
		Balance:        1000,
	}
	db.Create(&account1)

	account2 := models.Account{
		Base:           models.Base{ID: uuid.NewString()},
		FamilyID:       family.ID,
		UserID:         user.ID,
		Name:           "Account 2",
		Currency:       "UAH",
		InitialBalance: 500,
		Balance:        500,
	}
	db.Create(&account2)

	category := models.Category{
		Base:     models.Base{ID: uuid.NewString()},
		FamilyID: family.ID,
		Name:     "Food",
		Type:     "expense",
	}
	db.Create(&category)

	counterparty := models.Counterparty{
		Base:     models.Base{ID: uuid.NewString()},
		FamilyID: family.ID,
		Name:     "Supermarket",
	}
	db.Create(&counterparty)

	t.Run("TestCreateExpense", func(t *testing.T) {
		tx := &models.Transaction{
			Base:           models.Base{ID: uuid.NewString()},
			FamilyID:       family.ID,
			UserID:         user.ID,
			AccountID:      account1.ID,
			CategoryID:     category.ID,
			CounterpartyID: counterparty.ID,
			Type:           "expense",
			Amount:         200,
			Date:           time.Now().UnixMilli(),
		}

		err := repo.Create(tx, nil, nil, nil)
		assert.NoError(t, err)

		// Verify account balance
		var updatedAcc models.Account
		db.First(&updatedAcc, "id = ?", account1.ID)
		assert.Equal(t, int64(800), updatedAcc.Balance)

		// Verify ServerVersion updated
		assert.True(t, tx.ServerVersion > 0)

		// Verify counterparty balance
		var cpBalance models.CounterpartyBalance
		db.First(&cpBalance, "counterparty_id = ? AND currency = ?", counterparty.ID, "UAH")
		// expense doesn't affect counterparty debt by default unless it's a loan
		// wait, calculateDebtDelta("expense", 200) -> 0
		assert.Equal(t, int64(0), cpBalance.Balance)
	})

	t.Run("TestCreateTransfer", func(t *testing.T) {
		// Reset balances
		db.Model(&models.Account{}).Where("id = ?", account1.ID).Update("balance", 1000)
		db.Model(&models.Account{}).Where("id = ?", account2.ID).Update("balance", 500)

		txFrom := &models.Transaction{
			Base:      models.Base{ID: uuid.NewString()},
			FamilyID:  family.ID,
			UserID:    user.ID,
			AccountID: account1.ID,
			Type:      "transfer_out",
			Amount:    300,
			Date:      time.Now().UnixMilli(),
		}
		txTo := &models.Transaction{
			Base:      models.Base{ID: uuid.NewString()},
			FamilyID:  family.ID,
			UserID:    user.ID,
			AccountID: account2.ID,
			Type:      "transfer_in",
			Amount:    300,
			Date:      time.Now().UnixMilli(),
		}

		err := repo.CreateTransfer(txFrom, txTo)
		assert.NoError(t, err)

		var acc1, acc2 models.Account
		db.First(&acc1, "id = ?", account1.ID)
		db.First(&acc2, "id = ?", account2.ID)

		assert.Equal(t, int64(700), acc1.Balance)
		assert.Equal(t, int64(800), acc2.Balance)
	})

	t.Run("TestUpdateTransaction", func(t *testing.T) {
		// Create a transaction first
		tx := &models.Transaction{
			Base:      models.Base{ID: uuid.NewString()},
			FamilyID:  family.ID,
			UserID:    user.ID,
			AccountID: account1.ID,
			Type:      "expense",
			Amount:    100,
			Date:      time.Now().UnixMilli(),
		}
		db.Model(&models.Account{}).Where("id = ?", account1.ID).Update("balance", 1000)
		err := repo.Create(tx, nil, nil, nil)
		assert.NoError(t, err)

		// Update amount
		txUpdate := *tx
		oldVersion := tx.ServerVersion
		txUpdate.Amount = 150

		// Wait a bit to ensure UnixNano changes
		time.Sleep(2 * time.Millisecond)

		err = repo.Update(tx.ID, family.ID, &txUpdate, nil, nil)
		assert.NoError(t, err)

		var updatedTx models.Transaction
		db.First(&updatedTx, "id = ?", tx.ID)
		assert.True(t, updatedTx.ServerVersion > oldVersion, "ServerVersion should increase on update")

		var updatedAcc models.Account
		db.First(&updatedAcc, "id = ?", account1.ID)
		// 1000 - 100 (original) + 100 (revert) - 150 (new) = 850
		assert.Equal(t, int64(850), updatedAcc.Balance)

		// Update type to income
		txUpdate2 := txUpdate
		txUpdate2.Type = "income"
		txUpdate2.Amount = 50
		err = repo.Update(tx.ID, family.ID, &txUpdate2, nil, nil)
		assert.NoError(t, err)

		db.First(&updatedAcc, "id = ?", account1.ID)
		// 850 + 150 (revert expense) + 50 (income) = 1050
		assert.Equal(t, int64(1050), updatedAcc.Balance)
	})

	t.Run("TestUpdateSyncedTransactionAllowsOnlyCategoryAndCounterparty", func(t *testing.T) {
		syncedAccount := models.Account{
			Base:     models.Base{ID: uuid.NewString(), IsSynced: true},
			FamilyID: family.ID,
			UserID:   user.ID,
			Name:     "Synced account",
			Currency: "UAH",
			Balance:  900,
		}
		db.Create(&syncedAccount)

		secondCategory := models.Category{
			Base:     models.Base{ID: uuid.NewString()},
			FamilyID: family.ID,
			Name:     "Transport",
			Type:     "expense",
		}
		db.Create(&secondCategory)
		secondCounterparty := models.Counterparty{
			Base:     models.Base{ID: uuid.NewString()},
			FamilyID: family.ID,
			Name:     "Fuel station",
		}
		db.Create(&secondCounterparty)

		bankTx := &models.Transaction{
			Base:           models.Base{ID: uuid.NewString()},
			FamilyID:       family.ID,
			UserID:         user.ID,
			AccountID:      syncedAccount.ID,
			CategoryID:     category.ID,
			CounterpartyID: counterparty.ID,
			ExternalID:     "bank-tx-1",
			Type:           "expense",
			Amount:         100,
			Date:           time.Now().UnixMilli(),
			Note:           "from bank",
			Currency:       "UAH",
		}
		db.Create(bankTx)

		update := *bankTx
		update.CategoryID = secondCategory.ID
		update.CounterpartyID = secondCounterparty.ID
		update.Amount = 999
		update.Date = bankTx.Date + 1000
		update.Note = "must stay unchanged"
		update.Type = "income"

		err := repo.Update(bankTx.ID, family.ID, &update, nil, nil)
		assert.NoError(t, err)

		var saved models.Transaction
		db.First(&saved, "id = ?", bankTx.ID)
		assert.Equal(t, secondCategory.ID, saved.CategoryID)
		assert.Equal(t, secondCounterparty.ID, saved.CounterpartyID)
		assert.Equal(t, bankTx.Amount, saved.Amount)
		assert.Equal(t, bankTx.Date, saved.Date)
		assert.Equal(t, bankTx.Note, saved.Note)
		assert.Equal(t, bankTx.Type, saved.Type)

		var savedAccount models.Account
		db.First(&savedAccount, "id = ?", syncedAccount.ID)
		assert.Equal(t, int64(900), savedAccount.Balance)
	})

	t.Run("TestDeleteTransaction", func(t *testing.T) {
		db.Model(&models.Account{}).Where("id = ?", account1.ID).Update("balance", 1000)
		tx := &models.Transaction{
			Base:      models.Base{ID: uuid.NewString()},
			FamilyID:  family.ID,
			UserID:    user.ID,
			AccountID: account1.ID,
			Type:      "expense",
			Amount:    100,
			Date:      time.Now().UnixMilli(),
		}
		repo.Create(tx, nil, nil, nil)

		err := repo.Delete(tx, nil)
		assert.NoError(t, err)

		var updatedAcc models.Account
		db.First(&updatedAcc, "id = ?", account1.ID)
		assert.Equal(t, int64(1000), updatedAcc.Balance)

		var deletedTx models.Transaction
		db.Unscoped().First(&deletedTx, "id = ?", tx.ID)
		assert.NotNil(t, deletedTx.DeletedAt)
	})

	t.Run("TestDebtLogic", func(t *testing.T) {
		// debt_take - should decrease balance (increase debt)
		tx := &models.Transaction{
			Base:           models.Base{ID: uuid.NewString()},
			FamilyID:       family.ID,
			UserID:         user.ID,
			AccountID:      account1.ID,
			CounterpartyID: counterparty.ID,
			Type:           "debt_take",
			Amount:         500,
			Currency:       "UAH",
			Date:           time.Now().UnixMilli(),
		}
		repo.Create(tx, nil, nil, nil)

		var balance models.CounterpartyBalance
		db.First(&balance, "counterparty_id = ? AND currency = ?", counterparty.ID, "UAH")
		assert.Equal(t, int64(-500), balance.Balance)

		// debt_repay - should increase balance (decrease debt)
		repayTx := &models.Transaction{
			Base:           models.Base{ID: uuid.NewString()},
			FamilyID:       family.ID,
			UserID:         user.ID,
			AccountID:      account1.ID,
			CounterpartyID: counterparty.ID,
			Type:           "debt_repay",
			Amount:         200,
			Currency:       "UAH",
			Date:           time.Now().UnixMilli(),
		}
		repo.Create(repayTx, nil, nil, nil)

		db.First(&balance, "counterparty_id = ? AND currency = ?", counterparty.ID, "UAH")
		assert.Equal(t, int64(-300), balance.Balance)
	})

	t.Run("TestGetAllTransactions", func(t *testing.T) {
		filter := TransactionFilter{
			FamilyID: family.ID,
			Limit:    20,
		}
		txs, count, err := repo.GetAll(filter)
		assert.NoError(t, err)
		assert.True(t, count > 0)
		assert.NotEmpty(t, txs)
	})

	t.Run("TestGetByID", func(t *testing.T) {
		// Use the first tx we created
		var someTx models.Transaction
		db.First(&someTx)

		res, err := repo.GetByID(someTx.ID, family.ID)
		assert.NoError(t, err)
		assert.Equal(t, someTx.ID, res.ID)
	})

	t.Run("TestBatchCreate", func(t *testing.T) {
		var before models.Account
		db.First(&before, "id = ?", account1.ID)
		txs := []models.Transaction{
			{Base: models.Base{ID: uuid.NewString()}, FamilyID: family.ID, UserID: user.ID, AccountID: account1.ID, Amount: 10, Type: "expense"},
			{Base: models.Base{ID: uuid.NewString()}, FamilyID: family.ID, UserID: user.ID, AccountID: account1.ID, Amount: 20, Type: "income"},
		}
		count, err := repo.BatchCreate(txs)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Equal(t, "UAH", txs[0].Currency)
		assert.Equal(t, "UAH", txs[1].Currency)

		var after models.Account
		db.First(&after, "id = ?", account1.ID)
		assert.Equal(t, before.Balance+10, after.Balance)
	})

	t.Run("TestGetPredictedCategory", func(t *testing.T) {
		// Create some history
		cat1 := models.Category{Base: models.Base{ID: "cat-pred-1"}, FamilyID: family.ID, Name: "Groceries"}
		cat2 := models.Category{Base: models.Base{ID: "cat-pred-2"}, FamilyID: family.ID, Name: "Gas"}
		db.Create(&cat1)
		db.Create(&cat2)

		now := time.Now().UnixMilli()
		tx1ID := uuid.NewString()
		tx := models.Transaction{Base: models.Base{ID: tx1ID}, FamilyID: family.ID, UserID: user.ID, Date: now}
		db.Create(&tx)
		db.Create(&models.TransactionItem{Base: models.Base{ID: uuid.NewString()}, TransactionID: tx1ID, Name: "Milk", CategoryID: &cat1.ID})
		db.Create(&models.TransactionItem{Base: models.Base{ID: uuid.NewString()}, TransactionID: tx1ID, Name: "Bread", CategoryID: &cat1.ID})

		tx2ID := uuid.NewString()
		tx2 := models.Transaction{Base: models.Base{ID: tx2ID}, FamilyID: family.ID, UserID: user.ID, Date: now}
		db.Create(&tx2)
		db.Create(&models.TransactionItem{Base: models.Base{ID: uuid.NewString()}, TransactionID: tx2ID, Name: "Petrol", CategoryID: &cat2.ID})

		// Test prediction
		res, err := repo.GetPredictedCategory(family.ID, "Milk")
		assert.NoError(t, err)
		assert.Equal(t, "cat-pred-1", res)

		res, err = repo.GetPredictedCategory(family.ID, "Petro")
		assert.NoError(t, err)
		assert.Equal(t, "cat-pred-2", res)
	})

	t.Run("TestSearchTransactionsCyrillic", func(t *testing.T) {
		note := "Оренда квартири " + uuid.NewString()
		tx := &models.Transaction{
			Base:     models.Base{ID: uuid.NewString()},
			FamilyID: family.ID,
			UserID:   user.ID,
			Note:     note,
			Amount:   1000,
			Type:     "expense",
			Date:     time.Now().UnixMilli(),
		}
		db.Create(tx)

		// Search exact - MUST set Limit
		_, count, err := repo.GetAll(TransactionFilter{FamilyID: family.ID, Search: "Оренда", Limit: 100})
		assert.NoError(t, err)
		assert.True(t, count >= 1)
	})
}
