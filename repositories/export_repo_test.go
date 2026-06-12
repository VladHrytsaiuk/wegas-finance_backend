package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestExportRepository(t *testing.T) {
	db, _ := SetupTestDB()
	repo := NewExportRepository(db)

	familyID := "fam-export"
	db.Create(&models.Family{Base: models.Base{ID: familyID}})
	db.Create(&models.Transaction{
		Base: models.Base{ID: "tx-1"},
		FamilyID: familyID,
		Amount: 100,
		Date: 1000,
	})

	t.Run("GetTransactionsForExport", func(t *testing.T) {
		filter := models.ExportFilterDTO{
			From: 500,
			To: 1500,
		}
		txs, err := repo.GetTransactionsForExport(familyID, filter)
		assert.NoError(t, err)
		assert.Len(t, txs, 1)
	})
}
