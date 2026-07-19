package services

import (
	"testing"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoreTransactionCandidate(t *testing.T) {
	receiptDate := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC).UnixMilli()
	total := int64(93310)
	entry := &models.InboxEntry{
		Total:      &total,
		Currency:   "UAH",
		OccurredAt: &receiptDate,
		ReceiptSource: models.ReceiptSource{
			Merchant: "Сільпо",
		},
	}
	transaction := models.Transaction{
		Amount:   total,
		Currency: "UAH",
		Date:     receiptDate + int64(time.Hour/time.Millisecond),
		Note:     "Сільпо Київ",
	}

	candidate := scoreTransactionCandidate(entry, transaction)
	assert.Equal(t, 90, candidate.Score)
	assert.Equal(t, []string{"рахунок", "сума", "валюта", "час", "контрагент"}, candidate.MatchedBy)
}

func TestFindTransactionCandidatesExcludesLinkedTransactions(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	familyID := uuid.NewString()
	user := &models.User{Base: models.Base{ID: uuid.NewString()}, FamilyID: familyID}
	accountID := uuid.NewString()
	total := int64(93310)
	receiptDate := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC).UnixMilli()

	receipt := models.ReceiptSource{
		Base:     models.Base{ID: uuid.NewString()},
		FamilyID: familyID,
		UserID:   user.ID,
		Merchant: "Сільпо",
		Total:    &total,
		Currency: "UAH",
	}
	require.NoError(t, db.Create(&receipt).Error)

	entry := models.InboxEntry{
		Base:              models.Base{ID: uuid.NewString()},
		FamilyID:          familyID,
		UserID:            user.ID,
		ReceiptSourceID:   receipt.ID,
		SelectedAccountID: &accountID,
		Total:             &total,
		Currency:          "UAH",
		OccurredAt:        &receiptDate,
	}
	require.NoError(t, db.Create(&entry).Error)

	matchingTransaction := models.Transaction{
		Base:       models.Base{ID: uuid.NewString()},
		FamilyID:   familyID,
		UserID:     user.ID,
		AccountID:  accountID,
		Amount:     total,
		Currency:   "UAH",
		Date:       receiptDate + int64(time.Hour/time.Millisecond),
		ExternalID: "mono-match",
		Type:       "expense",
	}
	linkedTransaction := models.Transaction{
		Base:       models.Base{ID: uuid.NewString()},
		FamilyID:   familyID,
		UserID:     user.ID,
		AccountID:  accountID,
		Amount:     total,
		Currency:   "UAH",
		Date:       receiptDate,
		ExternalID: "mono-linked",
		Type:       "expense",
	}
	require.NoError(t, db.Create(&matchingTransaction).Error)
	require.NoError(t, db.Create(&linkedTransaction).Error)

	linkedSource := models.ReceiptSource{
		Base:                models.Base{ID: uuid.NewString()},
		FamilyID:            familyID,
		UserID:              user.ID,
		LinkedTransactionID: &linkedTransaction.ID,
	}
	require.NoError(t, db.Create(&linkedSource).Error)

	service := NewInboxService(repositories.NewInboxRepository(db), db)
	candidates, err := service.FindTransactionCandidates(entry.ID, user)
	require.NoError(t, err)
	if assert.Len(t, candidates, 1) {
		assert.Equal(t, matchingTransaction.ID, candidates[0].TransactionID)
	}
}

func TestLinkRejectsTransactionThatDoesNotMatchReceipt(t *testing.T) {
	tests := []struct {
		name               string
		transactionAmount  int64
		transactionAccount string
	}{
		{
			name:               "different account",
			transactionAmount:  93310,
			transactionAccount: uuid.NewString(),
		},
		{
			name:               "different amount",
			transactionAmount:  93311,
			transactionAccount: "selected-account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := repositories.SetupTestDB()
			require.NoError(t, err)

			familyID := uuid.NewString()
			user := &models.User{Base: models.Base{ID: uuid.NewString()}, FamilyID: familyID}
			selectedAccountID := "selected-account"
			total := int64(93310)

			receipt := models.ReceiptSource{
				Base:     models.Base{ID: uuid.NewString()},
				FamilyID: familyID,
				UserID:   user.ID,
				Total:    &total,
			}
			require.NoError(t, db.Create(&receipt).Error)

			entry := models.InboxEntry{
				Base:              models.Base{ID: uuid.NewString()},
				FamilyID:          familyID,
				UserID:            user.ID,
				ReceiptSourceID:   receipt.ID,
				SelectedAccountID: &selectedAccountID,
				Total:             &total,
			}
			require.NoError(t, db.Create(&entry).Error)

			transaction := models.Transaction{
				Base:      models.Base{ID: uuid.NewString()},
				FamilyID:  familyID,
				UserID:    user.ID,
				AccountID: tt.transactionAccount,
				Amount:    tt.transactionAmount,
				Type:      "expense",
			}
			require.NoError(t, db.Create(&transaction).Error)

			service := NewInboxService(repositories.NewInboxRepository(db), db)
			_, err = service.Link(entry.ID, transaction.ID, false, user)
			require.Error(t, err)

			var source models.ReceiptSource
			require.NoError(t, db.First(&source, "id = ?", receipt.ID).Error)
			assert.Nil(t, source.LinkedTransactionID)
		})
	}
}
