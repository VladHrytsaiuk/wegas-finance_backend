package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
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

		// Б) Таблиці без user_id з дедуплікацією (мігруємо все, що належало старій персональній сім'ї)
		if oldFamilyID != "" {
			// Карти для мапінгу ID (oldID -> newID) для збереження ієрархії
			categoryMap := make(map[string]string)
			counterpartyMap := make(map[string]string)

			// 1. Дедуплікація Категорій Транзакцій
			var oldCategories []models.Category
			if err := tx.Where("family_id = ? AND deleted_at IS NULL", oldFamilyID).Find(&oldCategories).Error; err == nil {
				for _, oldCat := range oldCategories {
					var existingCat models.Category
					// Шукаємо таку саму категорію (тип + назва) в новій сім'ї (Кирилиця-friendly)
					err := tx.Where("family_id = ? AND type = ? AND (name = ? OR name = ? OR name = ?)",
						joinCode.FamilyID, oldCat.Type, oldCat.Name, strings.ToLower(oldCat.Name), strings.ToUpper(oldCat.Name)).
						Where("deleted_at IS NULL").First(&existingCat).Error

					if err == nil {
						// Знайшли дублікат
						categoryMap[oldCat.ID] = existingCat.ID
						tx.Table("transactions").Where("category_id = ?", oldCat.ID).Update("category_id", existingCat.ID)
						tx.Table("transaction_items").Where("category_id = ?", oldCat.ID).Update("category_id", existingCat.ID)
						tx.Delete(&oldCat)
					} else {
						// Не знайшли - просто переносимо
						categoryMap[oldCat.ID] = oldCat.ID
						tx.Model(&oldCat).Update("family_id", joinCode.FamilyID)
					}
				}
				// Оновлюємо ParentID для категорій, що переїхали
				for _, oldCat := range oldCategories {
					if oldCat.ParentID != "" {
						if newParentID, ok := categoryMap[oldCat.ParentID]; ok {
							tx.Model(&models.Category{}).Where("id = ?", categoryMap[oldCat.ID]).Update("parent_id", newParentID)
						}
					}
				}
			}

			// 2. Дедуплікація Категорій Контрагентів
			var oldCPCats []models.CounterpartyCategory
			if err := tx.Where("family_id = ? AND deleted_at IS NULL", oldFamilyID).Find(&oldCPCats).Error; err == nil {
				for _, oldCPCat := range oldCPCats {
					var existingCPCat models.CounterpartyCategory
					err := tx.Where("family_id = ? AND (name = ? OR name = ? OR name = ?)",
						joinCode.FamilyID, oldCPCat.Name, strings.ToLower(oldCPCat.Name), strings.ToUpper(oldCPCat.Name)).
						Where("deleted_at IS NULL").First(&existingCPCat).Error

					if err == nil {
						tx.Table("counterparties").Where("category_id = ?", oldCPCat.ID).Update("category_id", existingCPCat.ID)
						tx.Delete(&oldCPCat)
					} else {
						tx.Model(&oldCPCat).Update("family_id", joinCode.FamilyID)
					}
				}
			}

			// 3. Дедуплікація Контрагентів
			var oldCPs []models.Counterparty
			if err := tx.Where("family_id = ? AND deleted_at IS NULL", oldFamilyID).Find(&oldCPs).Error; err == nil {
				for _, oldCP := range oldCPs {
					var existingCP models.Counterparty
					err := tx.Where("family_id = ? AND (name = ? OR name = ? OR name = ?)",
						joinCode.FamilyID, oldCP.Name, strings.ToLower(oldCP.Name), strings.ToUpper(oldCP.Name)).
						Where("deleted_at IS NULL").First(&existingCP).Error

					if err == nil {
						// Знайшли дублікат
						counterpartyMap[oldCP.ID] = existingCP.ID
						tx.Table("transactions").Where("counterparty_id = ?", oldCP.ID).Update("counterparty_id", existingCP.ID)
						tx.Table("utility_meters").Where("counterparty_id = ?", oldCP.ID).Update("counterparty_id", existingCP.ID)

						// Мержимо баланси (борги)
						var oldBalances []models.CounterpartyBalance
						if err := tx.Where("counterparty_id = ?", oldCP.ID).Find(&oldBalances).Error; err == nil {
							for _, b := range oldBalances {
								var exBalance models.CounterpartyBalance
								if err := tx.Where("counterparty_id = ? AND currency = ?", existingCP.ID, b.Currency).First(&exBalance).Error; err == nil {
									tx.Model(&exBalance).Update("balance", gorm.Expr("balance + ?", b.Balance))
									tx.Delete(&b)
								} else {
									tx.Model(&b).Update("counterparty_id", existingCP.ID)
								}
							}
						}
						tx.Delete(&oldCP)
					} else {
						// Не знайшли - переносимо
						counterpartyMap[oldCP.ID] = oldCP.ID
						tx.Model(&oldCP).Update("family_id", joinCode.FamilyID)
					}
				}
				// Оновлюємо ParentID для контрагентів
				for _, oldCP := range oldCPs {
					if oldCP.ParentID != "" {
						if newParentID, ok := counterpartyMap[oldCP.ParentID]; ok {
							tx.Model(&models.Counterparty{}).Where("id = ?", counterpartyMap[oldCP.ID]).Update("parent_id", newParentID)
						}
					}
				}
			}

			// 4. Дедуплікація Тегів
			var oldTags []models.Tag
			if err := tx.Where("family_id = ? AND deleted_at IS NULL", oldFamilyID).Find(&oldTags).Error; err == nil {
				for _, oldTag := range oldTags {
					var existingTag models.Tag
					err := tx.Where("family_id = ? AND (name = ? OR name = ? OR name = ?)",
						joinCode.FamilyID, oldTag.Name, strings.ToLower(oldTag.Name), strings.ToUpper(oldTag.Name)).
						Where("deleted_at IS NULL").First(&existingTag).Error

					if err == nil {
						tx.Table("transaction_tags").Where("tag_id = ?", oldTag.ID).Update("tag_id", existingTag.ID)
						tx.Delete(&oldTag)
					} else {
						tx.Model(&oldTag).Update("family_id", joinCode.FamilyID)
					}
				}
			}

			// 5. Решта (Utility Meters) - просто переносимо
			tx.Table("utility_meters").Where("family_id = ?", oldFamilyID).Update("family_id", joinCode.FamilyID)

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
