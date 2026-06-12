package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
)

func TestCurrencyService(t *testing.T) {
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)

	service := NewCurrencyService(db)

	t.Run("Convert - Same currency", func(t *testing.T) {
		res, err := service.Convert(100, "UAH", "UAH")
		assert.NoError(t, err)
		assert.Equal(t, int64(100), res)
	})

	t.Run("Convert - With predefined rates", func(t *testing.T) {
		// Mock rates in DB
		db.Create(&models.ExchangeRate{CurrencyCode: "USD", Rate: 40.0})
		db.Create(&models.ExchangeRate{CurrencyCode: "EUR", Rate: 44.0})

		// Convert 100 USD to UAH
		// 100 * 40.0 / 1.0 = 4000
		res, err := service.Convert(10000, "USD", "UAH") // 100.00 represented as 10000
		assert.NoError(t, err)
		assert.Equal(t, int64(400000), res) // 4000.00

		// Convert 100 UAH to USD
		// 100 / 40.0 = 2.5
		res, err = service.Convert(10000, "UAH", "USD")
		assert.NoError(t, err)
		assert.Equal(t, int64(250), res) // 2.50
	})

	t.Run("GetAllRates", func(t *testing.T) {
		rates, err := service.GetAllRates()
		assert.NoError(t, err)
		assert.NotEmpty(t, rates)
	})
}
