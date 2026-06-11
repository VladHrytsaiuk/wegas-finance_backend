package models

import "time"

// FamilyJoinCode - тимчасовий код для приєднання до сім'ї
type FamilyJoinCode struct {
	Base
	FamilyID  string    `json:"family_id" gorm:"index"`
	RoleID    string    `json:"role_id"`
	Code      string    `json:"code" gorm:"uniqueIndex"`
	ExpiresAt time.Time `json:"expires_at"`
}
