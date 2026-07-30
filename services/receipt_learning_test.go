package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReceiptLearningRequiresTwoManualConfirmations(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	user := &models.User{Base: models.Base{ID: uuid.NewString()}, FamilyID: uuid.NewString()}
	categoryID := uuid.NewString()
	counterpartyID := uuid.NewString()
	source := &models.ReceiptSource{
		Base:     models.Base{ID: uuid.NewString()},
		FamilyID: user.FamilyID,
		UserID:   user.ID,
		Merchant: "Сільпо",
		Items: []models.ReceiptSourceItem{{
			Base:        models.Base{ID: uuid.NewString()},
			Name:        "Молоко 2,5%",
			TotalAmount: 8990,
		}},
	}
	transaction := models.Transaction{CategoryID: categoryID, CounterpartyID: counterpartyID}
	items := []models.TransactionItem{{Name: "Молоко 2,5%", TotalAmount: 8990, CategoryID: &categoryID}}

	require.NoError(t, learnReceiptPreferences(db, user, source, transaction, items, 1))
	merchant, err := receiptPreferenceForMerchant(db, user, source.Merchant)
	require.NoError(t, err)
	assert.Nil(t, merchant)
	assert.Nil(t, receiptItemCategoryForName(db, user, source.Merchant, source.Items[0].Name))

	require.NoError(t, learnReceiptPreferences(db, user, source, transaction, items, 2))
	merchant, err = receiptPreferenceForMerchant(db, user, source.Merchant)
	require.NoError(t, err)
	if assert.NotNil(t, merchant) {
		assert.Equal(t, categoryID, *merchant.CategoryID)
		assert.Equal(t, counterpartyID, *merchant.CounterpartyID)
	}
	assert.Equal(t, categoryID, *receiptItemCategoryForName(db, user, source.Merchant, source.Items[0].Name))
}

func TestReceiptMerchantKeyNormalizesChainVariants(t *testing.T) {
	for input, expected := range map[string]string{
		`ТОВ "Сільпо-Фуд" № 123`: "сільпо",
		"СІЛЬПО Київ":            "сільпо",
		"АТБ-Маркет #45":         "атб",
		"Аврора 12":              "аврора",
	} {
		assert.Equal(t, expected, receiptMerchantKey(input), input)
	}
}
