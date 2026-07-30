package models

import "time"

// SystemSetting - налаштування системи (наприклад, maintenance_mode)
type SystemSetting struct {
	Key       string    `json:"key" gorm:"primaryKey"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}
