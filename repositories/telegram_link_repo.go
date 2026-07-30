package repositories

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type TelegramLinkRepository interface {
	CreateToken(token *models.TelegramLinkToken) error
	GetValidTokenByHash(tokenHash string, now int64) (*models.TelegramLinkToken, error)
	MarkTokenUsed(id string, usedAt int64) error
	InvalidateUnusedTokens(userID string, now int64) error

	GetActiveLinkByUserID(userID string) (*models.TelegramLink, error)
	GetActiveLinkByTelegramUserID(telegramUserID int64) (*models.TelegramLink, error)
	CreateLink(link *models.TelegramLink) error
	UpdateLink(link *models.TelegramLink) error
	DeactivateLink(id string, revokedAt int64) error
}

type telegramLinkRepo struct {
	db *gorm.DB
}

func NewTelegramLinkRepository(db *gorm.DB) TelegramLinkRepository {
	return &telegramLinkRepo{db: db}
}

func (r *telegramLinkRepo) CreateToken(token *models.TelegramLinkToken) error {
	return r.db.Create(token).Error
}

func (r *telegramLinkRepo) GetValidTokenByHash(tokenHash string, now int64) (*models.TelegramLinkToken, error) {
	var token models.TelegramLinkToken
	err := r.db.
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ? AND deleted_at IS NULL", tokenHash, now).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *telegramLinkRepo) MarkTokenUsed(id string, usedAt int64) error {
	return r.db.Model(&models.TelegramLinkToken{}).
		Where("id = ? AND used_at IS NULL", id).
		Updates(map[string]interface{}{
			"used_at":     usedAt,
			"updated_at":  usedAt,
			"is_synced":   true,
		}).Error
}

func (r *telegramLinkRepo) InvalidateUnusedTokens(userID string, now int64) error {
	return r.db.Model(&models.TelegramLinkToken{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Updates(map[string]interface{}{
			"used_at":    now,
			"updated_at": now,
			"is_synced":  true,
		}).Error
}

func (r *telegramLinkRepo) GetActiveLinkByUserID(userID string) (*models.TelegramLink, error) {
	var link models.TelegramLink
	err := r.db.
		Where("user_id = ? AND is_active = ? AND deleted_at IS NULL", userID, true).
		First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *telegramLinkRepo) GetActiveLinkByTelegramUserID(telegramUserID int64) (*models.TelegramLink, error) {
	var link models.TelegramLink
	err := r.db.
		Where("telegram_user_id = ? AND is_active = ? AND deleted_at IS NULL", telegramUserID, true).
		First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *telegramLinkRepo) CreateLink(link *models.TelegramLink) error {
	return r.db.Create(link).Error
}

func (r *telegramLinkRepo) UpdateLink(link *models.TelegramLink) error {
	return r.db.Save(link).Error
}

func (r *telegramLinkRepo) DeactivateLink(id string, revokedAt int64) error {
	return r.db.Model(&models.TelegramLink{}).
		Where("id = ? AND is_active = ?", id, true).
		Updates(map[string]interface{}{
			"is_active":   false,
			"revoked_at":  revokedAt,
			"updated_at":  revokedAt,
			"is_synced":   true,
		}).Error
}
