package models

// Transaction - основна операція
type Transaction struct {
	Base
	FamilyID       string `json:"family_id" gorm:"index;index:idx_dashboard"`
	UserID         string `json:"user_id" gorm:"index"`
	AccountID      string `json:"account_id" gorm:"index"`
	CategoryID     string `json:"category_id" gorm:"index"`
	CounterpartyID string `json:"counterparty_id" gorm:"index"`
	
	// Зв'язок із зовнішніми інтеграціями
	ExternalID     string `json:"external_id" gorm:"index"`

	Type string `json:"type" gorm:"index:idx_dashboard"`
	// expense, income, transfer, transfer_out, transfer_in, 
	// loan_give, loan_repay, debt_take, debt_repay

	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Date     int64  `json:"date" gorm:"index"`
	Note     string `json:"note"`
	
	ReceiptImg string `json:"receipt_img"` // Основне фото (legacy або прев'ю)
	IsForgiveness bool `json:"is_forgiveness" gorm:"default:false"` // Прощення боргу

	// Зв'язки з модулями
	AssetID *string `json:"asset_id" gorm:"index"`
	Asset   *Asset  `json:"asset" gorm:"foreignKey:AssetID"`
	Mileage *int    `json:"mileage"` // Пробіг при транзакції (для авто)

	// Перекази (Transfer)
	TransferRelatedID *string      `json:"transfer_related_id" gorm:"index"`
	TransferRelated   *Transaction `json:"related_transaction" gorm:"foreignKey:TransferRelatedID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// Relations
	Account      Account            `json:"account" gorm:"foreignKey:AccountID"`
	Category     Category           `json:"category" gorm:"foreignKey:CategoryID"`
	Counterparty Counterparty       `json:"counterparty" gorm:"foreignKey:CounterpartyID"`
	User         User               `json:"user" gorm:"foreignKey:UserID"`
	
	Items        []TransactionItem  `json:"items" gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE;"`
	Photos       []TransactionPhoto `json:"photos" gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE;"`
	Tags         []Tag              `json:"tags" gorm:"many2many:transaction_tags;"`
}

// TransactionItem - деталізація чеку
type TransactionItem struct {
	Base
	TransactionID string `json:"transaction_id" gorm:"index"`
	CategoryID    *string `json:"category_id"` // Категорія конкретного товару

	Name         string `json:"name"`
	Quantity     int64  `json:"quantity"`
	PricePerUnit int64  `json:"price_per_unit"`
	TotalAmount  int64  `json:"total_amount"`
}

// TransactionPhoto - додаткові фото
type TransactionPhoto struct {
	Base
	TransactionID string `json:"transaction_id" gorm:"index"`
	Path          string `json:"path"`
}

// Tag - мітки
type Tag struct {
	Base
	FamilyID string `json:"family_id" gorm:"index"`
	Name     string `json:"name"`
	Color    string `json:"color"`
}

type TransactionTag struct {
	Base
	TransactionID string `json:"transaction_id" gorm:"index"`
	TagID         string `json:"tag_id" gorm:"index"`
}