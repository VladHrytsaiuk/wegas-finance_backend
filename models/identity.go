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
	RoleID       string `json:"role_id" gorm:"index"`
	FamilyID     string `json:"family_id" gorm:"index"`
	
	Name         string `json:"name"`
	Email        string `json:"email" gorm:"uniqueIndex"` // Додав unique
	PasswordHash string `json:"password_hash"`
	AvatarURL    string `json:"avatar_url"`
	
	// Налаштування користувача
	BaseCurrency string `json:"base_currency" gorm:"default:'UAH'"`
	Language     string `json:"language" gorm:"default:'uk'"`
	Theme        string `json:"theme" gorm:"default:'light'"`

	// Relations
	Role   Role   `json:"role" gorm:"foreignKey:RoleID"`
	Family Family `json:"family" gorm:"foreignKey:FamilyID"`
}