package models

// Family - коренева сутність
type Family struct {
	Base
	Name string `json:"name"`
}

// Role - права доступу
type Role struct {
	Base
	Name           string `json:"name"`
	Description    string `json:"description"`
	CanManageUsers bool   `json:"can_manage_users"`
	CanEditSchema  bool   `json:"can_edit_schema"`
}

// User - користувач системи
type User struct {
	Base
	RoleID   string `json:"role_id" gorm:"index"`
	FamilyID string `json:"family_id" gorm:"index"`

	Name         string `json:"name"`
	Email        string `json:"email" gorm:"uniqueIndex"` // Додав unique
	PasswordHash string `json:"-"`
	AvatarURL    string `json:"avatar_url"`

	// Налаштування користувача
	BaseCurrency         string `json:"base_currency" gorm:"default:'UAH'"`
	Language             string `json:"language" gorm:"default:'uk'"`
	Theme                string `json:"theme" gorm:"default:'light'"`
	RequireReceiptReview bool   `json:"require_receipt_review" gorm:"default:true"`
	MobileAccountsOrder  string `json:"mobile_accounts_order" gorm:"type:text"`
	IsPlatformAdmin      bool   `json:"is_platform_admin" gorm:"default:false"`
	IsActive             bool   `json:"is_active" gorm:"default:true"`
	SessionVersion       int    `json:"session_version" gorm:"default:1"`

	// Security flags (calculated)
	HasPassword bool `json:"has_password" gorm:"-"`
	HasPin      bool `json:"has_pin" gorm:"-"`
	HasPasskeys bool `json:"has_passkeys" gorm:"-"`

	// PIN Authentication
	PinHash           string `json:"-"`
	FailedPinAttempts int    `json:"-" gorm:"default:0"`
	PinLockedUntil    int64  `json:"-" gorm:"default:0"`
	LastPinAttemptAt  int64  `json:"-" gorm:"default:0"`

	// Relations
	Role   Role   `json:"role" gorm:"foreignKey:RoleID"`
	Family Family `json:"family" gorm:"foreignKey:FamilyID"`
}
