package models

import (
	"time"

	"gorm.io/gorm"
)

// BankConnection
type BankConnection struct {
	// 👇 ПРИБРАЛИ default:gen_random_uuid()
	ID        string         `gorm:"primaryKey" json:"id"` 
	UserID    string         `gorm:"not null" json:"user_id"`
	FamilyID  string         `gorm:"not null" json:"family_id"`
	
	Provider  string         `gorm:"type:varchar(20);default:'monobank'" json:"provider"`
	Token     string         `gorm:"type:text;not null" json:"-"`
	
	
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	LastSync  *time.Time     `json:"last_sync"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BankAccountMapping
type BankAccountMapping struct {
	// 👇 ПРИБРАЛИ default:gen_random_uuid()
	ID                string `gorm:"primaryKey" json:"id"`
	ConnectionID      string `gorm:"not null" json:"connection_id"`
	
	ExternalID        string `gorm:"not null" json:"external_id"`
	InternalAccountID string `gorm:"type:uuid" json:"internal_account_id"`
	
	Name              string `json:"name"`
	RawData           string `gorm:"type:jsonb" json:"raw_data"`
	
	IsEnabled         bool   `gorm:"default:false" json:"is_enabled"`

	Currency          string `json:"currency"`
	SyncFrom          int64  `json:"sync_from"`
	CardNumber string `json:"card_number"`
	PaymentSystem  string `json:"payment_system"`
	BankName      string `json:"bank_name"`
	CardType    string `json:"card_type"`
}