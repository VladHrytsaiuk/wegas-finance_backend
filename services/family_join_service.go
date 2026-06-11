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
)

type FamilyJoinService interface {
	GenerateCode(familyID string) (string, error)
	JoinFamily(userID string, code string) (*models.Family, error)
}

type familyJoinService struct {
	repo       repositories.FamilyJoinRepository
	userRepo   repositories.UserRepository
	wsHub      *utils.WSHub
}

func NewFamilyJoinService(repo repositories.FamilyJoinRepository, userRepo repositories.UserRepository, wsHub *utils.WSHub) FamilyJoinService {
	return &familyJoinService{
		repo:     repo,
		userRepo: userRepo,
		wsHub:    wsHub,
	}
}

func (s *familyJoinService) GenerateCode(familyID string) (string, error) {
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

	if user.FamilyID != "" {
		return nil, errors.New("you are already in a family. Please leave your current family first")
	}

	joinCode, err := s.repo.GetCode(code)
	if err != nil {
		return nil, errors.New("invalid or expired code")
	}

	if time.Now().After(joinCode.ExpiresAt) {
		_ = s.repo.DeleteCode(code)
		return nil, errors.New("invalid or expired code")
	}

	// Оновлюємо FamilyID користувача
	if err := s.repo.UpdateUserFamily(userID, joinCode.FamilyID); err != nil {
		return nil, err
	}

	// Отримуємо інформацію про сім'ю, щоб повернути її
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
