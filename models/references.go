package models

// Category - категорії транзакцій
type Category struct {
	Base
	ParentID string `json:"parent_id"`
	FamilyID string `json:"family_id" gorm:"index"`
	
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	Type     string `json:"type"` // 'expense', 'income'
}

// ExchangeRate - курси валют
type ExchangeRate struct {
	CurrencyCode string  `gorm:"primaryKey" json:"currency_code"`
	Rate         float64 `json:"rate"` // Відносно базової валюти (зазвичай USD або UAH)
	UpdatedAt    int64   `json:"updated_at"`
}