package models

import "time"

// AuditLog - запис логу для дій адміністратора
type AuditLog struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	AdminID    string    `json:"admin_id" gorm:"index"`
	Action     string    `json:"action"` // create, update, delete, block, unblock
	EntityType string    `json:"entity_type"` // user, category, system_setting
	EntityID   string    `json:"entity_id"`
	Changes    string    `json:"changes"` // JSON representation of changes
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`

	Admin User `json:"admin" gorm:"foreignKey:AdminID"`
}
