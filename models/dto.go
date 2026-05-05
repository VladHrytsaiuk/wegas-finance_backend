package models

// Settings - налаштування програми (JSON config)
type Settings struct {
	BaseCurrency string `json:"base_currency"`
	Language     string `json:"language"`
	Theme        string `json:"theme"`
}
// DefaultSettings повертає стандартні налаштування
// (Це функція-конструктор, тому вона має жити поруч зі структурою)
func DefaultSettings() Settings {
	return Settings{
		BaseCurrency: "UAH",    // Стандартна валюта
		Language:     "uk",     // Стандартна мова
		Theme:        "light",  // Стандартна тема
	}
}

// CreateAssetOnFlyInput - DTO для швидкого створення
type CreateAssetOnFlyInput struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	SerialNumber string `json:"serial_number"`
	WarrantyEnd  int64  `json:"warranty_end"`
	Note         string `json:"note"`
	
	Price        int64  `json:"price"`
	PurchaseDate int64  `json:"purchase_date"`

	VINCode         string  `json:"vin_code"`
	Mileage         int     `json:"mileage"`
	InsuranceExpiry int64   `json:"insurance_expiry"`
	LastServiceDate int64   `json:"last_service_date"`
	Address         string  `json:"address"`
	Area            float64 `json:"area"`
}

// ExportFilterDTO - для фільтрації експорту
type ExportFilterDTO struct {
	From            int64    `form:"from"`
	To              int64    `form:"to"`
	AccountIDs      []string `form:"accountIds[]"`
	CategoryIDs     []string `form:"categoryIds[]"`
	UserIDs         []string `form:"userIds[]"`
	CounterpartyIDs []string `form:"counterpartyIds[]"`
	Types           []string `form:"type[]"`
}

// UtilityStatRaw - для вибірки статистики SQL
type UtilityStatRaw struct {
	Month string  `json:"month"`
	Type  string  `json:"type"`
	Total float64 `json:"total"`
}

type UtilityStatGlobalDTO struct {
	Month string             `json:"month"`
	Data  map[string]float64 `json:"data"`
}

// UtilityStatMeterDTO - статистика одного лічильника (для графіків)
type UtilityStatMeterDTO struct {
	Month            string  `json:"month"`
	TotalConsumption float64 `json:"total_consumption"` // Спожито одиниць (кВт, м3)
	TotalCost        float64 `json:"total_cost"`        // Гроші (копійки)
	AvgTariff        float64 `json:"avg_tariff"`        // Середній тариф за місяць
}

// --- DTO для НОТАТОК (Shopping List) ---
type CreateShoppingListRequest struct {
	Title      string `json:"title" binding:"required"`
	Color      string `json:"color"`
	Visibility string `json:"visibility"`
	HiddenFrom string `json:"hidden_from"`
}

type UpdateShoppingListRequest struct {
	Title      *string `json:"title"`
	Color      *string `json:"color"`
	Visibility *string `json:"visibility"`
	HiddenFrom *string `json:"hidden_from"`
}

// --- DTO для ПУНКТІВ (Shopping Item) ---
type CreateShoppingItemRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateShoppingItemRequest struct {
	Name     *string `json:"name"`
	IsBought *bool   `json:"is_bought"`
}


type CreateWishlistGroupRequest struct {
	Name       string `json:"name" binding:"required"`
	Color      string `json:"color"`
	Icon       string `json:"icon"`
	Visibility string `json:"visibility"`
	HiddenFrom string `json:"hidden_from"`
}

type UpdateWishlistGroupRequest struct {
	Name       string `json:"name"`
	Color      string `json:"color"`
	Icon       string `json:"icon"`
	Visibility string `json:"visibility"`
	HiddenFrom string `json:"hidden_from"`
}

// CreateWishlistRequest - DTO для створення
type CreateWishlistRequest struct {
	GroupID string `json:"group_id"`
	Name       string `json:"name" binding:"required"`
	URL        string `json:"url"`
	Price      int64  `json:"price"`
	Currency   string `json:"currency"`
	Priority   int    `json:"priority"`
	Visibility string `json:"visibility"`
	HiddenFrom string `json:"hidden_from"`
}

// UpdateWishlistRequest - DTO для оновлення
type UpdateWishlistRequest struct {
	GroupID *string `json:"group_id"`
	Name       *string `json:"name"`
	URL        *string `json:"url"`
	Price      *int64  `json:"price"`
	Currency   *string `json:"currency"`
	Priority   *int    `json:"priority"`
	Status     *string `json:"status"`
	Visibility *string `json:"visibility"`
	HiddenFrom *string `json:"hidden_from"`
	GoalID     *string `json:"goal_id"`
}