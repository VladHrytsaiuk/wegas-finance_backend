package services

import (
	"testing"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestMonobankService_ProcessWebhook(t *testing.T) {
	setup := func() (*gorm.DB, *MockTransactionService, MonobankService) {
		db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		db.AutoMigrate(
			&models.BankAccountMapping{},
			&models.BankConnection{},
			&models.User{},
			&models.Transaction{},
			&models.Category{},
			&models.Account{},
		)
		mockTxService := new(MockTransactionService)
		mockAccRepo := new(MockAccountRepository)
		mockClock := utils.NewMockClock(time.Now())
		service := NewMonobankService(db, mockTxService, mockAccRepo, mockClock)
		return db, mockTxService, service
	}

	t.Run("ProcessWebhook - success", func(t *testing.T) {
		db, mockTxService, service := setup()
		
		// Setup data
		user := models.User{Base: models.Base{ID: "user-1"}, FamilyID: "family-1", Name: "Test User"}
		db.Create(&user)
		conn := models.BankConnection{ID: "conn-1", UserID: "user-1", FamilyID: "family-1", Provider: "monobank"}
		db.Create(&conn)
		mapping := models.BankAccountMapping{ID: "mapping-1", ConnectionID: "conn-1", ExternalID: "mono-acc-1", InternalAccountID: "internal-acc-1", IsEnabled: true}
		db.Create(&mapping)

		payload := MonoWebhookPayload{Type: "StatementItem"}
		payload.Data.Account = "mono-acc-1"
		payload.Data.StatementItem = MonoTransaction{
			ID: "mono-tx-1", Amount: -10000, Time: time.Now().Unix(), Description: "Silpo", Mcc: 5411,
		}

		mockTxService.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("tx-1", nil).Once()

		err := service.ProcessWebhook(payload)

		assert.NoError(t, err)
		mockTxService.AssertExpectations(t)
	})

	t.Run("ProcessWebhook - duplicate tx", func(t *testing.T) {
		db, mockTxService, service := setup()

		// Setup data
		user := models.User{Base: models.Base{ID: "user-1"}, FamilyID: "family-1"}
		db.Create(&user)
		conn := models.BankConnection{ID: "conn-1", UserID: "user-1"}
		db.Create(&conn)
		mapping := models.BankAccountMapping{ID: "mapping-1", ConnectionID: "conn-1", ExternalID: "mono-acc-1", InternalAccountID: "internal-acc-1", IsEnabled: true}
		db.Create(&mapping)

		// Create existing tx
		db.Create(&models.Transaction{
			Base:       models.Base{ID: "existing-tx"},
			ExternalID: "mono-tx-2",
			AccountID:  "internal-acc-1",
		})

		payload := MonoWebhookPayload{Type: "StatementItem"}
		payload.Data.Account = "mono-acc-1"
		payload.Data.StatementItem = MonoTransaction{ID: "mono-tx-2"}

		err := service.ProcessWebhook(payload)

		assert.NoError(t, err)
		mockTxService.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("ProcessWebhook - mapping not found", func(t *testing.T) {
		_, _, service := setup()
		payload := MonoWebhookPayload{
			Type: "StatementItem",
		}
		payload.Data.Account = "unknown-acc"

		err := service.ProcessWebhook(payload)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account mapping not found")
	})
}
