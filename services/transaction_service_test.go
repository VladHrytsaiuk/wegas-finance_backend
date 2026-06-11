package services

import (
	"testing"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionService_Create(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	mockCPRepo := new(MockCounterpartyRepository)
	mockAssetRepo := new(MockAssetRepository)
	mockStorage := new(MockStorageService)
	fixedNow := time.Now()
	mockClock := utils.NewMockClock(fixedNow)

	service := NewTransactionService(nil, mockRepo, mockCPRepo, mockAssetRepo, mockStorage, mockClock)

	user := &models.User{
		Base:     models.Base{ID: "user-1"},
		FamilyID: "family-1",
	}

	t.Run("Create transfer - success", func(t *testing.T) {
		input := CreateTransactionInput{
			AccountID:       "acc-1",
			TargetAccountID: "acc-2",
			Amount:          1000,
			Type:            "transfer",
			Date:            fixedNow.UnixMilli(),
		}

		mockRepo.On("CreateTransfer", mock.Anything, mock.Anything).Return(nil).Once()

		txID, err := service.Create(input, nil, user)

		assert.NoError(t, err)
		assert.NotEmpty(t, txID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Create expense - success", func(t *testing.T) {
		input := CreateTransactionInput{
			AccountID: "acc-1",
			Amount:    500,
			Type:      "expense",
			Date:      fixedNow.UnixMilli(),
		}

		mockRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		txID, err := service.Create(input, nil, user)

		assert.NoError(t, err)
		assert.NotEmpty(t, txID)
		mockRepo.AssertCalled(t, "Create", mock.MatchedBy(func(tx *models.Transaction) bool {
			return tx.Amount == 500 && tx.Type == "expense" && tx.AccountID == "acc-1"
		}), mock.Anything, mock.Anything, mock.Anything)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Create expense with new asset - success", func(t *testing.T) {
		input := CreateTransactionInput{
			AccountID: "acc-1",
			Amount:    200000,
			Type:      "expense",
			Date:      fixedNow.UnixMilli(),
			NewAsset: &models.CreateAssetOnFlyInput{
				Name: "iPhone 15",
				Type: "electronics",
			},
		}

		mockRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(asset *models.Asset) bool {
			return asset != nil && asset.Name == "iPhone 15"
		})).Return(nil).Once()

		txID, err := service.Create(input, nil, user)

		assert.NoError(t, err)
		assert.NotEmpty(t, txID)
		mockRepo.AssertExpectations(t)
	})
}

func TestTransactionService_BatchCreate(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	mockCPRepo := new(MockCounterpartyRepository)
	fixedNow := time.Now()
	mockClock := utils.NewMockClock(fixedNow)

	service := NewTransactionService(nil, mockRepo, mockCPRepo, nil, nil, mockClock)

	user := &models.User{
		Base:     models.Base{ID: "user-1"},
		FamilyID: "family-1",
	}

	t.Run("Batch create - success", func(t *testing.T) {
		inputs := []CreateTransactionInput{
			{AccountID: "acc-1", Amount: 100, Type: "expense", Date: fixedNow.UnixMilli()},
			{AccountID: "acc-1", Amount: 200, Type: "income", Date: fixedNow.UnixMilli()},
		}

		mockRepo.On("BatchCreate", mock.MatchedBy(func(txs []models.Transaction) bool {
			return len(txs) == 2
		})).Return(2, nil).Once()

		count, err := service.BatchCreate(inputs, user)

		assert.NoError(t, err)
		assert.Equal(t, 2, count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Batch create with counterparty name - success", func(t *testing.T) {
		inputs := []CreateTransactionInput{
			{AccountID: "acc-1", Amount: 100, Type: "expense", CounterpartyName: "Silpo"},
		}

		mockCPRepo.On("GetByName", "Silpo", "family-1").Return(&models.Counterparty{Base: models.Base{ID: "cp-1"}}, nil).Once()
		mockRepo.On("BatchCreate", mock.MatchedBy(func(txs []models.Transaction) bool {
			return len(txs) == 1 && txs[0].CounterpartyID == "cp-1"
		})).Return(1, nil).Once()

		count, err := service.BatchCreate(inputs, user)

		assert.NoError(t, err)
		assert.Equal(t, 1, count)
		mockCPRepo.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
