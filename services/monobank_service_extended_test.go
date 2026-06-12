package services

import (
	"testing"
	"time"
	"net/http"
	"net/http/httptest"
	"fmt"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMonobankService_Extended(t *testing.T) {
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)

	mockTxService := new(MockTransactionService)
	mockAccRepo := new(MockAccountRepository)
	mockClock := utils.NewMockClock(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))
	
	svc := NewMonobankService(db, mockTxService, mockAccRepo, mockClock).(*monobankService)
	svc.SkipRateLimit = true

	familyID := "fam-mono"
	userID := "user-mono"
	token := "valid-token"
	encryptedToken, _ := utils.Encrypt(token)

	db.Create(&models.Family{Base: models.Base{ID: familyID}})
	db.Create(&models.User{Base: models.Base{ID: userID}, FamilyID: familyID})

	t.Run("GetUserData - no connection", func(t *testing.T) {
		accounts, mappings, err := svc.GetUserData("non-existent")
		assert.Error(t, err)
		assert.Nil(t, accounts)
		assert.Nil(t, mappings)
	})

	t.Run("SaveSettings - with internal account creation", func(t *testing.T) {
		conn := models.BankConnection{
			ID:       "conn-1",
			UserID:   userID,
			FamilyID: familyID,
			Provider: "monobank",
			Token:    encryptedToken,
			IsActive: true,
		}
		db.Create(&conn)

		mockAccRepo.On("Create", mock.Anything).Return(nil).Once()

		settings := []models.BankAccountMapping{
			{
				ExternalID: "ext-1",
				Name:       "New Account",
				Currency:   "UAH",
				IsEnabled:  true,
				RawData:    `{"type":"black","maskedPan":["444455******6666"]}`,
			},
		}

		err := svc.SaveSettings(userID, familyID, settings)
		assert.NoError(t, err)
		
		var savedMapping models.BankAccountMapping
		err = db.Where("connection_id = ?", "conn-1").First(&savedMapping).Error
		assert.NoError(t, err)
		assert.NotEmpty(t, savedMapping.InternalAccountID)
		assert.Equal(t, "6666", savedMapping.CardNumber)
		
		mockAccRepo.AssertExpectations(t)
	})

	t.Run("GetUserData - with data", func(t *testing.T) {
		var mapping models.BankAccountMapping
		db.Where("external_id = ?", "ext-1").First(&mapping)
		
		db.Create(&models.Account{
			Base: models.Base{ID: mapping.InternalAccountID},
			Balance: 1234,
			Currency: "UAH",
		})

		accounts, mappings, err := svc.GetUserData(userID)
		assert.NoError(t, err)
		assert.Len(t, accounts, 1)
		assert.Len(t, mappings, 1)
		assert.Equal(t, int64(1234), accounts[0].Balance)
	})

	t.Run("RefreshClientInfo", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"clientId":"c1","accounts":[{"id":"ext-1","balance":999}]}`)
		}))
		defer server.Close()
		svc.baseURL = server.URL

		accounts, mappings, err := svc.RefreshClientInfo(userID)
		assert.NoError(t, err)
		assert.Len(t, accounts, 1)
		assert.Equal(t, int64(999), accounts[0].Balance)
		assert.Len(t, mappings, 1)
	})

	t.Run("Disconnect", func(t *testing.T) {
		err := svc.Disconnect(userID)
		assert.NoError(t, err)
		
		var count int64
		db.Model(&models.BankConnection{}).Where("user_id = ?", userID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("GlobalResyncCounterparties", func(t *testing.T) {
		cp := models.Counterparty{Base: models.Base{ID: "cp-1"}, FamilyID: familyID, Name: "Сільпо"}
		db.Create(&cp)
		
		tx := models.Transaction{
			Base: models.Base{ID: "tx-resync"},
			FamilyID: familyID,
			ExternalID: "mono-123",
			Note: "SILPO",
			CounterpartyID: "wrong-id",
		}
		db.Create(&tx)

		count, err := svc.GlobalResyncCounterparties()
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
		
		var updatedTx models.Transaction
		db.First(&updatedTx, "id = ?", "tx-resync")
		assert.Equal(t, "cp-1", updatedTx.CounterpartyID)
	})
}
