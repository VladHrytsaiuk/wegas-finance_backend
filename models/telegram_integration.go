package models

type TelegramLink struct {
	Base
	UserID    string `json:"user_id" gorm:"index"`
	FamilyID  string `json:"family_id" gorm:"index"`
	IsActive  bool   `json:"is_active" gorm:"index;default:true"`
	LinkedAt  int64  `json:"linked_at"`

	TelegramUserID int64  `json:"telegram_user_id" gorm:"index"`
	TelegramChatID int64  `json:"telegram_chat_id"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	RevokedAt      *int64 `json:"revoked_at"`
}

type TelegramLinkToken struct {
	Base
	UserID    string `json:"user_id" gorm:"index"`
	FamilyID  string `json:"family_id" gorm:"index"`
	TokenHash string `json:"-" gorm:"uniqueIndex"`
	ExpiresAt int64  `json:"expires_at" gorm:"index"`
	UsedAt    *int64 `json:"used_at"`
}
