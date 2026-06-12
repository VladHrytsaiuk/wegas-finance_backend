package services

import (
	"bytes"
	"mime/multipart"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
)

func TestImportService_ProcessFile(t *testing.T) {
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)

	service := NewImportService(db)

	familyID := "fam-import"
	accountID := "acc-import"
	db.Create(&models.Family{Base: models.Base{ID: familyID}})
	db.Create(&models.Account{Base: models.Base{ID: accountID}, FamilyID: familyID, Currency: "UAH"})

	t.Run("Process Monobank CSV", func(t *testing.T) {
		csvData := "Дата,Опис,MCC,Сума\n01.01.2024 12:00:00,АТБ,5411,-100.50"
		
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "mono.csv")
		part.Write([]byte(csvData))
		writer.Close()

		reader := multipart.NewReader(body, writer.Boundary())
		form, _ := reader.ReadForm(1024)
		fileHeader := form.File["file"][0]

		result, err := service.ProcessFile(fileHeader, accountID, "monobank")
		assert.NoError(t, err)
		assert.Len(t, result.Transactions, 1)
		assert.Equal(t, int64(10050), result.Transactions[0].Amount)
		assert.Equal(t, "АТБ", result.Transactions[0].CounterpartyName)
	})

	t.Run("Duplicate check", func(t *testing.T) {
		// Pre-create transaction
		db.Create(&models.Transaction{
			Base:      models.Base{ID: "tx-existing"},
			AccountID: accountID,
			Amount:    10050,
			Type:      "expense",
			Date:      1704103200000, // 01.01.2024 10:00:00 UTC (approx)
		})

		// 01.01.2024 12:00:00 (local) might be different depending on timezone, 
		// but checkMatchStatus uses exact time +- 10 mins.
		// Let's use a known timestamp for both.
	})
}
