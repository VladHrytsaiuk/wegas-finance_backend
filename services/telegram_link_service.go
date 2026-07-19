package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const telegramLinkTokenTTL = 15 * time.Minute

type TelegramLinkStatus struct {
	IsLinked         bool   `json:"is_linked"`
	BotUsername      string `json:"bot_username,omitempty"`
	TelegramUserID   *int64 `json:"telegram_user_id,omitempty"`
	TelegramChatID   *int64 `json:"telegram_chat_id,omitempty"`
	TelegramUsername string `json:"telegram_username,omitempty"`
	TelegramFirstName string `json:"telegram_first_name,omitempty"`
	LinkedAt         *int64 `json:"linked_at,omitempty"`
}

type TelegramLinkTokenResponse struct {
	Token      string `json:"token"`
	DeepLink   string `json:"deep_link"`
	ExpiresAt  int64  `json:"expires_at"`
	BotUsername string `json:"bot_username"`
}

type TelegramLinkCompleteInput struct {
	Token          string `json:"token"`
	TelegramUserID int64  `json:"telegram_user_id"`
	TelegramChatID int64  `json:"telegram_chat_id"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
}

type TelegramLinkService interface {
	GetStatus(user *models.User) (*TelegramLinkStatus, error)
	CreateLinkToken(user *models.User) (*TelegramLinkTokenResponse, error)
	CompleteLink(input TelegramLinkCompleteInput) (*TelegramLinkStatus, error)
	RevokeLink(user *models.User) error
}

type telegramLinkService struct {
	repo        repositories.TelegramLinkRepository
	db          *gorm.DB
	botUsername string
}

func NewTelegramLinkService(repo repositories.TelegramLinkRepository, db *gorm.DB, botUsername string) TelegramLinkService {
	return &telegramLinkService{
		repo:        repo,
		db:          db,
		botUsername: botUsername,
	}
}

func (s *telegramLinkService) GetStatus(user *models.User) (*TelegramLinkStatus, error) {
	link, err := s.repo.GetActiveLinkByUserID(user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &TelegramLinkStatus{
				IsLinked:    false,
				BotUsername: s.botUsername,
			}, nil
		}
		return nil, err
	}
	return mapTelegramLinkStatus(link, s.botUsername), nil
}

func (s *telegramLinkService) CreateLinkToken(user *models.User) (*TelegramLinkTokenResponse, error) {
	if s.botUsername == "" {
		return nil, errors.New("telegram bot username is not configured")
	}

	now := time.Now().UnixMilli()
	rawToken, err := generateTelegramLinkToken()
	if err != nil {
		return nil, err
	}
	expiresAt := now + telegramLinkTokenTTL.Milliseconds()

	token := &models.TelegramLinkToken{
		Base:      models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
		UserID:    user.ID,
		FamilyID:  user.FamilyID,
		TokenHash: hashTelegramLinkToken(rawToken),
		ExpiresAt: expiresAt,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		repo := repositories.NewTelegramLinkRepository(tx)
		if err := repo.InvalidateUnusedTokens(user.ID, now); err != nil {
			return err
		}
		return repo.CreateToken(token)
	}); err != nil {
		return nil, err
	}

	return &TelegramLinkTokenResponse{
		Token:       rawToken,
		DeepLink:    fmt.Sprintf("https://t.me/%s?start=%s", s.botUsername, rawToken),
		ExpiresAt:   expiresAt,
		BotUsername: s.botUsername,
	}, nil
}

func (s *telegramLinkService) CompleteLink(input TelegramLinkCompleteInput) (*TelegramLinkStatus, error) {
	if input.Token == "" {
		return nil, errors.New("token is required")
	}
	if input.TelegramUserID == 0 || input.TelegramChatID == 0 {
		return nil, errors.New("telegram identifiers are required")
	}

	now := time.Now().UnixMilli()
	tokenHash := hashTelegramLinkToken(input.Token)

	var result *TelegramLinkStatus
	err := s.db.Transaction(func(tx *gorm.DB) error {
		repo := repositories.NewTelegramLinkRepository(tx)

		token, err := repo.GetValidTokenByHash(tokenHash, now)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("invalid or expired telegram link token")
			}
			return err
		}

		existingByTelegram, err := repo.GetActiveLinkByTelegramUserID(input.TelegramUserID)
		if err == nil && existingByTelegram.UserID != token.UserID {
			return errors.New("telegram account is already linked to another user")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		activeLink, err := repo.GetActiveLinkByUserID(token.UserID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if activeLink != nil && activeLink.ID != "" {
			activeLink.TelegramUserID = input.TelegramUserID
			activeLink.TelegramChatID = input.TelegramChatID
			activeLink.Username = input.Username
			activeLink.FirstName = input.FirstName
			activeLink.LinkedAt = now
			activeLink.IsActive = true
			activeLink.RevokedAt = nil
			if err := repo.UpdateLink(activeLink); err != nil {
				return err
			}
			result = mapTelegramLinkStatus(activeLink, s.botUsername)
		} else {
			link := &models.TelegramLink{
				Base:           models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
				UserID:         token.UserID,
				FamilyID:       token.FamilyID,
				IsActive:       true,
				LinkedAt:       now,
				TelegramUserID: input.TelegramUserID,
				TelegramChatID: input.TelegramChatID,
				Username:       input.Username,
				FirstName:      input.FirstName,
			}
			if err := repo.CreateLink(link); err != nil {
				return err
			}
			result = mapTelegramLinkStatus(link, s.botUsername)
		}

		return repo.MarkTokenUsed(token.ID, now)
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *telegramLinkService) RevokeLink(user *models.User) error {
	link, err := s.repo.GetActiveLinkByUserID(user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	now := time.Now().UnixMilli()
	return s.db.Transaction(func(tx *gorm.DB) error {
		repo := repositories.NewTelegramLinkRepository(tx)
		if err := repo.DeactivateLink(link.ID, now); err != nil {
			return err
		}
		return repo.InvalidateUnusedTokens(user.ID, now)
	})
}

func mapTelegramLinkStatus(link *models.TelegramLink, botUsername string) *TelegramLinkStatus {
	userID := link.TelegramUserID
	chatID := link.TelegramChatID
	linkedAt := link.LinkedAt
	return &TelegramLinkStatus{
		IsLinked:          true,
		BotUsername:       botUsername,
		TelegramUserID:    &userID,
		TelegramChatID:    &chatID,
		TelegramUsername:  link.Username,
		TelegramFirstName: link.FirstName,
		LinkedAt:          &linkedAt,
	}
}

func generateTelegramLinkToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

func hashTelegramLinkToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
