package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestUtilityRepo(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	repo := NewUtilityRepository(db)
	familyID := "fam-util"

	t.Run("Create and Get Meter", func(t *testing.T) {
		meter := &models.UtilityMeter{
			Base:     models.Base{ID: "m-1"},
			FamilyID: familyID,
			Name:     "Electricity",
			Type:     "electricity",
			Unit:     "kWh",
			Tariff:   2.64,
		}
		err := repo.CreateMeter(meter)
		assert.NoError(t, err)

		meters, err := repo.GetMeters(familyID)
		assert.NoError(t, err)
		assert.Len(t, meters, 1)
		assert.Equal(t, "Electricity", meters[0].Name)
	})

	t.Run("Create and Get Readings", func(t *testing.T) {
		read1 := &models.UtilityReading{
			Base:    models.Base{ID: "r-1"},
			MeterID: "m-1",
			Date:    1700000000000,
			Value:   100,
		}
		repo.CreateReading(read1)

		read2 := &models.UtilityReading{
			Base:    models.Base{ID: "r-2"},
			MeterID: "m-1",
			Date:    1710000000000,
			Value:   150,
		}
		repo.CreateReading(read2)

		prev, _ := repo.GetPreviousReading("m-1", 1710000000000)
		assert.NotNil(t, prev)
		assert.Equal(t, float64(150), prev.Value) // Previous includes current date in repo logic

		// Previous for a date between 1 and 2
		prevMid, _ := repo.GetPreviousReading("m-1", 1705000000000)
		assert.Equal(t, float64(100), prevMid.Value)
	})
}
