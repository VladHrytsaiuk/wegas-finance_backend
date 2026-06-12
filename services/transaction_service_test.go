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

func TestTransactionService_Other(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	mockAssetRepo := new(MockAssetRepository)
	service := NewTransactionService(nil, mockRepo, nil, mockAssetRepo, nil, utils.NewMockClock(time.Now()))

	user := &models.User{Base: models.Base{ID: "u-1"}, FamilyID: "f-1"}

	t.Run("Predict Category", func(t *testing.T) {
		mockRepo.On("GetPredictedCategory", "f-1", "Milk").Return("cat-1", nil).Once()
		res, err := service.PredictCategory("Milk", user)
		assert.NoError(t, err)
		assert.Equal(t, "cat-1", res)
	})

	t.Run("Delete Tx", func(t *testing.T) {
		tx := &models.Transaction{Base: models.Base{ID: "tx-1"}, UserID: "u-1"}
		// GetByID called once in Delete and once in DeleteReceipt
		mockRepo.On("GetByID", "tx-1", "f-1").Return(tx, nil).Times(2)
		mockRepo.On("DeleteAllPhotos", "tx-1").Return(nil).Once()
		mockRepo.On("Delete", tx, mock.Anything).Return(nil).Once()
		
		err := service.Delete("tx-1", user)
		assert.NoError(t, err)
	})

	t.Run("Update with Mileage Ratchet", func(t *testing.T) {
		txID := "tx-1"
		assetID := "asset-1"
		input := CreateTransactionInput{
			Amount: 100,
			AssetID: &assetID,
			Mileage: func(i int) *int { return &i }(1500),
		}
		
		mockRepo.On("GetByID", txID, "f-1").Return(&models.Transaction{UserID: "u-1"}, nil).Once()
		mockRepo.On("Update", txID, "f-1", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		
		// Asset currently has 1000 mileage
		mockAssetRepo.On("GetByID", assetID, "f-1").Return(&models.Asset{Base: models.Base{ID: assetID}, Mileage: 1000}, nil).Once()
		mockAssetRepo.On("UpdateServiceData", assetID, 1500, mock.Anything).Return(nil).Once()

		err := service.Update(txID, input, user)
		assert.NoError(t, err)
		mockAssetRepo.AssertExpectations(t)
	})
}
