package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
)

func TestUtilityService(t *testing.T) {
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)

	repo := repositories.NewUtilityRepository(db)
	txRepo := repositories.NewTransactionRepository(db)
	assetRepo := repositories.NewAssetRepository(db)
	service := NewUtilityService(repo, txRepo, assetRepo)

	user := &models.User{Base: models.Base{ID: "u-util"}, FamilyID: "fam-util"}
	db.Create(&models.Family{Base: models.Base{ID: "fam-util"}})

	t.Run("Full Cycle: Meter -> Reading -> Payment", func(t *testing.T) {
		// 1. Create Meter with Counterparty
		cp := &models.Counterparty{Base: models.Base{ID: "cp-yasno"}, FamilyID: "fam-util", Name: "Yasno"}
		db.Create(cp)
		
		meter := models.UtilityMeter{
			Name: "Electro",
			Type: "electricity",
			Unit: "kWh",
			Tariff: 2.64,
			CounterpartyID: &cp.ID,
			Currency: "UAH",
		}
		err := service.CreateMeter(meter, user)
		assert.NoError(t, err)
		
		meters, _ := service.GetMeters(user)
		meterID := meters[0].ID

		// 2. Create First Reading (Point Zero)
		read1 := models.UtilityReading{
			MeterID: meterID,
			Date: 1704103200000,
			Value: 0, // Set to 0 to avoid large debt for "initial" reading
		}
		err = service.CreateReading(read1, user)
		assert.NoError(t, err)

		// 3. Create Second Reading (Usage)
		read2 := models.UtilityReading{
			MeterID: meterID,
			Date: 1706781600000,
			Value: 100, // 100 kWh usage
		}
		err = service.CreateReading(read2, user)
		assert.NoError(t, err)

		// 4. Verify Debt Transaction Created
		// Cost = 100 * 2.64 * 100 = 26400 (cents)
		var txs []models.Transaction
		db.Where("type = ?", "debt_take").Find(&txs)
		assert.Len(t, txs, 1)
		assert.Equal(t, int64(26400), txs[0].Amount)

		// 5. Pay Reading
		readings, _ := service.GetReadings(user, meterID)
		readingID := readings[0].ID // The latest one
		
		acc := &models.Account{Base: models.Base{ID: "acc-pay"}, FamilyID: "fam-util", Currency: "UAH", Balance: 100000}
		db.Create(acc)

		err = service.PayReading(readingID, "acc-pay", user)
		assert.NoError(t, err)

		// 6. Verify Payment Transaction
		var payTxs []models.Transaction
		db.Where("type = ?", "debt_repay").Find(&payTxs)
		assert.Len(t, payTxs, 1)
		assert.Equal(t, int64(26400), payTxs[0].Amount)
	})
}
