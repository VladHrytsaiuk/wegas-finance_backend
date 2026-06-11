package models

import (
	"time"
	"gorm.io/gorm"
)

// Base - спільні поля для всіх сутностей
// Використовуємо int64 для часу (Unix timestamp), як у твоєму коді.
type Base struct {
	ID              string  `gorm:"primaryKey" json:"id"`
	CreatedAt       int64   `json:"created_at"`
	UpdatedAt       int64   `json:"updated_at"`
	DeletedAt       *int64  `json:"deleted_at" gorm:"index"` // Індекс для soft delete
	IsSynced        bool    `json:"is_synced" gorm:"index;default:false"`

	// Local-First fields
	ServerVersion   int64   `json:"server_version" gorm:"index;default:1"` // Інкрементальна версія на сервері
	ClientUpdatedAt int64   `json:"client_updated_at"`                    // Час останньої зміни на клієнті
	}

	// BeforeUpdate - GORM Hook, який спрацьовує перед кожним оновленням
	func (b *Base) BeforeUpdate(tx *gorm.DB) error {
	// Використовуємо UnixNano для ServerVersion, щоб мати монотонно зростаючий маркер для синхронізації.
	// Це дозволяє клієнту запитати "дай мені все, що змінилося після версії X".
	b.ServerVersion = time.Now().UnixNano()
	b.UpdatedAt = time.Now().UnixMilli()
	return nil
	}

	// BeforeCreate - GORM Hook для нових записів
	func (b *Base) BeforeCreate(tx *gorm.DB) error {
	now := time.Now().UnixMilli()
	if b.CreatedAt == 0 {
		b.CreatedAt = now
	}
	b.UpdatedAt = now
	b.ServerVersion = time.Now().UnixNano()
	return nil
	}