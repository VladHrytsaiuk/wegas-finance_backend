package models

// Counterparty - люди, магазини, компанії
type Counterparty struct {
	Base
	FamilyID string `json:"family_id" gorm:"index"`
	ParentID string `json:"parent_id"`
	IsGroup  bool   `json:"is_group"`

	SystemKey        string  `json:"system_key" gorm:"index"`
	GlobalTemplateID *string `json:"global_template_id" gorm:"index"`
	IsArchived       bool    `json:"is_archived" gorm:"default:false"`

	Name string `json:"name"`
	Type string `json:"type"` // 'shop', 'person', 'other'

	CategoryID *string               `json:"category_id"`
	Category   *CounterpartyCategory `json:"category" gorm:"foreignKey:CategoryID"`

	Icon string `json:"icon"`
	Logo string `json:"logo"`

	// Computed only for platform-admin catalog responses; not persisted.
	UsageCount int64 `json:"usage_count" gorm:"->;-:migration"`

	// Баланси (борги)
	Balances []CounterpartyBalance `json:"balances" gorm:"foreignKey:CounterpartyID"`
}

// CounterpartyBalance - денормалізований баланс по валютах
type CounterpartyBalance struct {
	CounterpartyID string `gorm:"primaryKey" json:"counterparty_id"`
	Currency       string `gorm:"primaryKey" json:"currency"`
	Balance        int64  `json:"balance"` // + нам винні, - ми винні
}

// CounterpartyCategory - категорії контрагентів
type CounterpartyCategory struct {
	Base
	FamilyID         string  `json:"family_id" gorm:"index"`
	SystemKey        string  `json:"system_key" gorm:"index"`
	GlobalTemplateID *string `json:"global_template_id" gorm:"index"`
	IsArchived       bool    `json:"is_archived" gorm:"default:false"`
	Name             string  `json:"name"`
	Type             string  `json:"type"`
	Icon             string  `json:"icon"`
	Color            string  `json:"color"`
	UsageCount       int64   `json:"usage_count" gorm:"->;-:migration"`
}
