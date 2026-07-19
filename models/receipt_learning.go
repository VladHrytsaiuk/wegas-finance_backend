package models

// ReceiptMerchantPreference stores only user-confirmed choices for a merchant.
type ReceiptMerchantPreference struct {
	Base
	FamilyID    string `gorm:"uniqueIndex:idx_receipt_merchant_preference"`
	UserID      string `gorm:"uniqueIndex:idx_receipt_merchant_preference"`
	MerchantKey string `gorm:"uniqueIndex:idx_receipt_merchant_preference"`

	CounterpartyID *string `gorm:"index"`
	CategoryID     *string `gorm:"index"`
	Confirmations  int     `gorm:"default:0"`
}

// ReceiptItemCategoryPreference stores user-confirmed categories for a product.
type ReceiptItemCategoryPreference struct {
	Base
	FamilyID    string `gorm:"uniqueIndex:idx_receipt_item_preference"`
	UserID      string `gorm:"uniqueIndex:idx_receipt_item_preference"`
	MerchantKey string `gorm:"uniqueIndex:idx_receipt_item_preference"`
	ItemKey     string `gorm:"uniqueIndex:idx_receipt_item_preference"`
	CategoryID  string `gorm:"uniqueIndex:idx_receipt_item_preference"`

	Confirmations int `gorm:"default:0"`
}
