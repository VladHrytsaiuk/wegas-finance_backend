package services

import (
	"testing"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
)

func TestStatsServiceExtended(t *testing.T) {
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)

	repo := repositories.NewStatsRepository(db)
	currSvc := NewCurrencyService(db)
	service := NewStatsService(repo, currSvc)

	familyID := "fam-stats"
	userID := "u-admin"
	childID := "u-child"
	
	db.Create(&models.Family{Base: models.Base{ID: familyID}})
	db.Create(&models.User{Base: models.Base{ID: userID}, FamilyID: familyID, RoleID: "admin"})
	db.Create(&models.User{Base: models.Base{ID: childID}, FamilyID: familyID, RoleID: "child"})

	// Seed exchange rates
	db.Create(&models.ExchangeRate{CurrencyCode: "USD", Rate: 40.0})
	db.Create(&models.ExchangeRate{CurrencyCode: "EUR", Rate: 43.0})
	
	// Create accounts
	uahAccount := models.Account{
		Base:     models.Base{ID: "acc-uah"},
		Name:     "UAH Account",
		Currency: "UAH",
		Balance:  1000,
		FamilyID: familyID,
		UserID:   userID,
	}
	usdAccount := models.Account{
		Base:     models.Base{ID: "acc-usd"},
		Name:     "USD Account",
		Currency: "USD",
		Balance:  100,
		FamilyID: familyID,
		UserID:   userID,
	}
	db.Create(&uahAccount)
	db.Create(&usdAccount)

	category := models.Category{
		Base:     models.Base{ID: "cat-1"},
		Name:     "Food",
		FamilyID: familyID,
	}
	db.Create(&category)

	now := time.Now().UnixMilli()

	// Transactions
	txs := []models.Transaction{
		{
			Base:      models.Base{ID: "tx-uah"},
			Amount:    400,
			Type:      "expense",
			Date:      now,
			FamilyID:  familyID,
			UserID:    userID,
			AccountID: "acc-uah",
			CategoryID: "cat-1",
		},
		{
			Base:      models.Base{ID: "tx-usd"},
			Amount:    10, // 10 USD * 40 = 400 UAH
			Type:      "expense",
			Date:      now,
			FamilyID:  familyID,
			UserID:    userID,
			AccountID: "acc-usd",
			CategoryID: "cat-1",
		},
		{
			Base:      models.Base{ID: "tx-child"},
			Amount:    100,
			Type:      "expense",
			Date:      now,
			FamilyID:  familyID,
			UserID:    childID,
			AccountID: "acc-uah",
			CategoryID: "cat-1",
		},
	}
	for _, tx := range txs {
		db.Create(&tx)
	}

	t.Run("GetDashboardData - Currency Conversion", func(t *testing.T) {
		adminUser := &models.User{Base: models.Base{ID: userID}, FamilyID: familyID, RoleID: "admin"}
		// Total balance: 1000 UAH + 100 USD * 40 = 5000 UAH
		// Total expense: 400 UAH + 10 USD * 40 + 100 UAH = 900 UAH
		data, err := service.GetDashboardData(adminUser, "UAH", now-1000, now+1000, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(5000), data.TotalBalance)
		assert.Equal(t, int64(900), data.TotalExpense)
	})

	t.Run("GetDashboardData - Role Access (Child)", func(t *testing.T) {
		childUser := &models.User{Base: models.Base{ID: childID}, FamilyID: familyID, RoleID: "child"}
		// Child can only see their own transactions
		// Balance check in GetBalances also filters by userID if child
		data, err := service.GetDashboardData(childUser, "UAH", now-1000, now+1000, nil)
		assert.NoError(t, err)
		// Account also has userID, so GetBalances for child should return only their accounts?
		// Wait, GetBalances in stats_repo.go:
		/*
		if restrictUserID != "" {
			query = query.Where("user_id = ?", restrictUserID)
		}
		*/
		// In our test, only uahAccount and usdAccount have userID = userID (u-admin).
		// So child sees 0 accounts and 0 balance.
		assert.Equal(t, int64(0), data.TotalBalance)
		assert.Equal(t, int64(100), data.TotalExpense) // Child transaction
	})

	t.Run("GetTopStats - Aggregation and Conversion", func(t *testing.T) {
		adminUser := &models.User{Base: models.Base{ID: userID}, FamilyID: familyID, RoleID: "admin"}
		results, err := service.GetTopStats(adminUser, "expense", "category", "UAH", now-1000, now+1000, nil)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Food", results[0].Name)
		assert.Equal(t, int64(900), results[0].Total)
	})

	t.Run("GetTrendStats", func(t *testing.T) {
		adminUser := &models.User{Base: models.Base{ID: userID}, FamilyID: familyID, RoleID: "admin"}
		results, err := service.GetTrendStats(adminUser, "expense", "UAH", now-1000, now+1000, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, int64(900), results[0].Total)
	})

	t.Run("GetRecentTransactions", func(t *testing.T) {
		adminUser := &models.User{Base: models.Base{ID: userID}, FamilyID: familyID, RoleID: "admin"}
		results, err := service.GetRecentTransactions(adminUser, nil)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
	})
}
