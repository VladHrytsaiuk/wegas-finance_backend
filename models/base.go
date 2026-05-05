package models

// Base - спільні поля для всіх сутностей
// Використовуємо int64 для часу (Unix timestamp), як у твоєму коді.
type Base struct {
	ID        string  `gorm:"primaryKey" json:"id"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
	DeletedAt *int64  `json:"deleted_at" gorm:"index"` // Індекс для soft delete
	IsSynced  bool    `json:"is_synced" gorm:"index;default:false"`
}