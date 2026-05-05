package models

// UtilityMeter - лічильник або послуга
type UtilityMeter struct {
	Base
	FamilyID string `json:"family_id" gorm:"index"`

	// Прив'язки
	AssetID        *string       `json:"asset_id" gorm:"index"`
	Asset          *Asset        `json:"asset" gorm:"foreignKey:AssetID"`
	CounterpartyID *string       `json:"counterparty_id" gorm:"index"`
	Counterparty   *Counterparty `json:"counterparty" gorm:"foreignKey:CounterpartyID"`

	Name            string `json:"name"`
	PersonalAccount string `json:"personal_account"`
	Type            string `json:"type"` // electricity, water, gas, internet...
	Unit            string `json:"unit"` // kW, m3, Gcal

	Tariff   float64 `json:"tariff"`
	Currency string  `json:"currency"`

	// Кеш останніх показників
	LastReadingDate  *int64   `json:"last_reading_date"`
	LastReadingValue *float64 `json:"last_reading_value"`
	
	// Хак для створення Asset "на льоту"
	NewAsset *CreateAssetOnFlyInput `json:"new_asset" gorm:"-"`
}

// UtilityReading - запис показника
type UtilityReading struct {
	Base
	MeterID string        `json:"meter_id" gorm:"index"`
	Meter   *UtilityMeter `json:"meter" gorm:"foreignKey:MeterID"`

	Date  int64   `json:"date"`
	Value float64 `json:"value"`

	Diff           float64 `json:"diff"`
	TariffAtDate   float64 `json:"tariff_at_date"`
	CalculatedCost int64   `json:"calculated_cost"`

	IsPaid bool `json:"is_paid"`

	// --- 👇 ПОВЕРНУВ TransactionID ---
	// Це поле використовується для зв'язку з транзакцією оплати
	TransactionID *string `json:"transaction_id" gorm:"index"`
	
	// Якщо у тебе логіка розділяє оплату і нарахування, залишаємо і це:
	PaymentTransactionID *string `json:"payment_transaction_id" gorm:"index"`
}