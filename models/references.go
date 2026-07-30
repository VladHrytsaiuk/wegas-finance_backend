package models

// Category - категорії транзакцій
type Category struct {
	Base
	ParentID string `json:"parent_id"`
	FamilyID string `json:"family_id" gorm:"index"`

	// System templates have an empty FamilyID and a stable SystemKey. Family
	// copies keep GlobalTemplateID until the user customizes them.
	SystemKey        string  `json:"system_key" gorm:"index"`
	GlobalTemplateID *string `json:"global_template_id" gorm:"index"`
	IsArchived       bool    `json:"is_archived" gorm:"default:false"`

	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
	Type  string `json:"type"` // 'expense', 'income'

	// Computed only for platform-admin catalog responses; not persisted.
	UsageCount int64 `json:"usage_count" gorm:"->;-:migration"`
}

// ExchangeRate - курси валют
type ExchangeRate struct {
	CurrencyCode string  `gorm:"primaryKey" json:"currency_code"`
	Rate         float64 `json:"rate"` // Відносно базової валюти (зазвичай USD або UAH)
	UpdatedAt    int64   `json:"updated_at"`
}
