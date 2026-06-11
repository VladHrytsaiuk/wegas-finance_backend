package models

import (
	"time"

	"gorm.io/gorm"
)

// Base - спільні поля для всіх сутностей
type Base struct {
	ID              string `gorm:"primaryKey" json:"id"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
	DeletedAt       *int64 `json:"deleted_at" gorm:"index"`
	IsSynced        bool   `json:"is_synced" gorm:"index;default:false"`
	ServerVersion   int64  `json:"server_version" gorm:"index;default:1"`
	ClientUpdatedAt int64  `json:"client_updated_at"`
}

// BeforeUpdate - GORM Hook, який спрацьовує перед кожним оновленням
func (b *Base) BeforeUpdate(tx *gorm.DB) error {
	nowNano := time.Now().UnixNano()
	nowMilli := time.Now().UnixMilli()
	b.ServerVersion = nowNano
	b.UpdatedAt = nowMilli

	// Ensure fields are updated even if a map is used in Updates()
	tx.Statement.SetColumn("server_version", nowNano)
	tx.Statement.SetColumn("updated_at", nowMilli)
	return nil
}

// BeforeCreate - GORM Hook для нових записів
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	nowNano := time.Now().UnixNano()
	nowMilli := time.Now().UnixMilli()

	if b.CreatedAt == 0 {
		b.CreatedAt = nowMilli
	}
	b.UpdatedAt = nowMilli
	b.ServerVersion = nowNano

	// Ensure fields are set for map-based creations if any (though usually structs are used)
	tx.Statement.SetColumn("server_version", nowNano)
	tx.Statement.SetColumn("updated_at", nowMilli)
	tx.Statement.SetColumn("created_at", b.CreatedAt)
	return nil
}
