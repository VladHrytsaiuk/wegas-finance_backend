package models

// Asset - майно
type Asset struct {
  Base
  FamilyID string `json:"family_id" gorm:"index"`
  UserID   string `json:"user_id"`

  Name string `json:"name"`
  Type string `json:"type"` // 'car', 'real_estate', 'appliance', 'furniture'

  // Фінанси
  Price        int64  `json:"price"`
  Currency     string `json:"currency"`
  CurrentPrice int64  `json:"current_price"`
  PurchaseDate int64  `json:"purchase_date"`

  // Деталі
  SerialNumber string `json:"serial_number"`
  WarrantyEnd  int64  `json:"warranty_end"`
  Note         string `json:"note"`

  // Продаж
  IsSold    bool   `json:"is_sold"`
  SoldDate  *int64 `json:"sold_date"`
  SoldPrice int64  `json:"sold_price"`

  // Медіа
  Photo  string       `json:"photo"`
  Photos []AssetPhoto `json:"photos" gorm:"foreignKey:AssetID;constraint:OnDelete:CASCADE;"`

  // 🔥 ДОДАНО: Документи (PDF та інші)
  Documents []AssetDocument `json:"documents" gorm:"foreignKey:AssetID;constraint:OnDelete:CASCADE;"`

  // Амортизація
  DepreciationType string `json:"depreciation_type"`
  EstimatedLife    int    `json:"estimated_life"`
  InitialValue     int64  `json:"initial_value"`

  // Автоспецифіка
  VINCode         string `json:"vin_code"`
  Mileage         int    `json:"mileage"`
  InitialMileage  int    `json:"initial_mileage"`
  InsuranceExpiry int64  `json:"insurance_expiry"`
  LastServiceDate int64  `json:"last_service_date"`

  // Нерухомість
  Address      string  `json:"address"`
  Area         float64 `json:"area"`
  CadastralNum string  `json:"cadastral_num"`
}

type AssetPhoto struct {
  Base
  AssetID string `json:"asset_id" gorm:"index"`
  Path    string `json:"path"`
}

// 🔥 НОВА МОДЕЛЬ ДЛЯ ДОКУМЕНТІВ
type AssetDocument struct {
  Base
  AssetID  string `json:"asset_id" gorm:"index"`
  Name     string `json:"name"`      // Оригінальна назва файлу (напр. "чек.pdf")
  Path     string `json:"path"`      // Шлях на сервері
  FileType string `json:"file_type"` // Розширення (напр. "pdf")
  Size     int64  `json:"size"`      // Розмір в байтах (опціонально, але корисно)
}