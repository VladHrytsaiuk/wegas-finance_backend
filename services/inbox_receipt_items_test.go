package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestTransactionItemsFromReceiptIncludesDiscount(t *testing.T) {
	subtotal := int64(94089)
	discount := int64(779)
	source := models.ReceiptSource{
		Subtotal:      &subtotal,
		DiscountTotal: &discount,
		Items: []models.ReceiptSourceItem{
			{Name: "Товар 1", Quantity: 1, PricePerUnit: 50000, TotalAmount: 50000},
			{Name: "Товар 2", Quantity: 1, PricePerUnit: 44089, TotalAmount: 44089},
		},
	}

	items := transactionItemsFromReceipt(source, "tx-1", 1)
	if assert.Len(t, items, 3) {
		assert.Equal(t, "Знижка за чеком", items[2].Name)
		assert.Equal(t, int64(-779), items[2].TotalAmount)
	}

	total := int64(0)
	for _, item := range items {
		total += item.TotalAmount
	}
	assert.Equal(t, int64(93310), total)
}

func TestTransactionItemsFromReceiptDoesNotDuplicateDistributedDiscount(t *testing.T) {
	total := int64(93310)
	discount := int64(21000)
	source := models.ReceiptSource{
		Total:         &total,
		DiscountTotal: &discount,
		Items: []models.ReceiptSourceItem{
			{Name: "Товар 1", Quantity: 1, PricePerUnit: 50000, TotalAmount: 50000},
			{Name: "Товар 2", Quantity: 1, PricePerUnit: 43310, TotalAmount: 43310},
		},
	}

	items := transactionItemsFromReceipt(source, "tx-1", 1)
	assert.Len(t, items, 2)
	assert.Equal(t, int64(93310), items[0].TotalAmount+items[1].TotalAmount)
}
