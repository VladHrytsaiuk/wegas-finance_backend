package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FamilyJoinService interface {
	GenerateCode(familyID string, roleID string) (string, error)
	JoinFamily(userID string, code string) (*models.Family, error)
}

type familyJoinService struct {
	repo     repositories.FamilyJoinRepository
	userRepo repositories.UserRepository
	wsHub    *utils.WSHub
	db       *gorm.DB
}

func NewFamilyJoinService(repo repositories.FamilyJoinRepository, userRepo repositories.UserRepository, wsHub *utils.WSHub, db *gorm.DB) FamilyJoinService {
	return &familyJoinService{
		repo:     repo,
		userRepo: userRepo,
		wsHub:    wsHub,
		db:       db,
	}
}

func (s *familyJoinService) GenerateCode(familyID string, roleID string) (string, error) {
	code, err := s.generateRandomSixDigits()
	if err != nil {
		return "", err
	}

	joinCode := &models.FamilyJoinCode{
		Base: models.Base{
			ID:        uuid.NewString(),
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
		},
		FamilyID:  familyID,
		RoleID:    roleID,
		Code:      code,
		ExpiresAt: time.Now().Add(2 * time.Minute),
	}

	if err := s.repo.CreateCode(joinCode); err != nil {
		return "", err
	}

	return code, nil
}

func (s *familyJoinService) JoinFamily(userID string, code string) (*models.Family, error) {
	// Перевірка, чи користувач вже є в сім'ї
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	var oldFamilyID string
	if user.FamilyID != "" {
		count, err := s.userRepo.CountFamilyMembers(user.FamilyID)
		if err != nil {
			return nil, err
		}

		if count > 1 {
			return nil, errors.New("you are already in a family with other members. Please leave it first")
		}
		
		// Запам'ятовуємо ID старої сім'ї для подальшого видалення
		oldFamilyID = user.FamilyID
	}

	joinCode, err := s.repo.GetCode(code)
	if err != nil {
		return nil, errors.New("invalid or expired code")
	}

	if time.Now().After(joinCode.ExpiresAt) {
		_ = s.repo.DeleteCode(code)
		return nil, errors.New("invalid or expired code")
	}

	// Виконуємо в транзакції для надійності міграції даних
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Оновлюємо FamilyID та RoleID користувача
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"family_id": joinCode.FamilyID,
			"role_id":   joinCode.RoleID,
		}).Error; err != nil {
			return err
		}

		// 2. Мігруємо дані користувача (оновлюємо family_id)
		
		// А) Таблиці з user_id
		tablesWithUserID := []string{
			"accounts",
			"transactions",
			"goals",
			"assets",
			"shopping_lists",
			"wishlist_groups",
			"wishlist_items",
			"medical_records",
			"bank_connections",
		}

		for _, table := range tablesWithUserID {
			if err := tx.Table(table).Where("user_id = ?", userID).Update("family_id", joinCode.FamilyID).Error; err != nil {
				return err
			}
		}

		// Б) Таблиці без user_id (мігруємо все, що належало старій персональній сім'ї)
		if oldFamilyID != "" {
			tablesWithFamilyIDOnly := []string{
				"counterparties",
				"counterparty_categories",
				"tags",
				"utility_meters",
				"categories",
			}
			for _, table := range tablesWithFamilyIDOnly {
				if err := tx.Table(table).Where("family_id = ?", oldFamilyID).Update("family_id", joinCode.FamilyID).Error; err != nil {
					return err
				}
			}

			// 3. Видаляємо стару порожню сім'ю
			if err := tx.Where("id = ?", oldFamilyID).Delete(&models.Family{}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Отримуємо інформацію про нову сім'ю
	family, err := s.userRepo.GetFamilyByID(joinCode.FamilyID)
	if err != nil {
		return nil, err
	}

	// Видаляємо код після успішного використання
	_ = s.repo.DeleteCode(code)

	// Сповіщення адміна та інших членів сім'ї через WebSocket
	s.wsHub.BroadcastToFamily(joinCode.FamilyID, map[string]interface{}{
		"type":    "member_joined",
		"user_id": userID,
		"message": "A new member has joined your family!",
	})

	return family, nil
}

func (s *familyJoinService) generateRandomSixDigits() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
