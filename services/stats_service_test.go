package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
)

func TestStatsService(t *testing.T) {
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)

	repo := repositories.NewStatsRepository(db)
	currSvc := NewCurrencyService(db)
	service := NewStatsService(repo, currSvc)

	familyID := "fam-stats"
	db.Create(&models.Family{Base: models.Base{ID: familyID}})
	
	user := &models.User{Base: models.Base{ID: "u-stats"}, FamilyID: familyID, RoleID: "admin"}
	
	// Create some data
	account := &models.Account{
		Base: models.Base{ID: "acc-stats"},
		FamilyID: familyID,
		UserID: "u-stats",
		Currency: "UAH",
		Balance: 1000,
	}
	db.Create(account)

	db.Create(&models.Transaction{
		Base: models.Base{ID: "tx-1"},
		FamilyID: familyID,
		UserID: "u-stats",
		AccountID: "acc-stats",
		Type: "expense",
		Amount: 500,
		Date: 1700000000000,
	})

	t.Run("GetDashboardData", func(t *testing.T) {
		data, err := service.GetDashboardData(user, "UAH", 0, 2000000000000, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(1000), data.TotalBalance)
		assert.Equal(t, int64(500), data.TotalExpense)
	})

	t.Run("GetTrendStats", func(t *testing.T) {
		trends, err := service.GetTrendStats(user, "expense", "UAH", 0, 2000000000000, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, trends)
		assert.Equal(t, int64(500), trends[0].Total)
	})
}
