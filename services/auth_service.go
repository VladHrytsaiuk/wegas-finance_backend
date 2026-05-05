package services

import (
	"errors"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegisterInput struct {
	Name       string
	Email      string
	Password   string
	InviteCode string
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type AuthService interface {
	Register(input RegisterInput) (*LoginResponse, error)
	Login(input LoginInput) (*LoginResponse, error)
}

type authService struct {
	userRepo   repositories.UserRepository
	secretKey  string
	inviteCode string
}

func NewAuthService(userRepo repositories.UserRepository, secretKey, inviteCode string) AuthService {
	return &authService{
		userRepo:   userRepo,
		secretKey:  secretKey,
		inviteCode: inviteCode,
	}
}

func (s *authService) Register(input RegisterInput) (*LoginResponse, error) {
	// 1. Перевірка Інвайт-коду
	if input.InviteCode != s.inviteCode {
		return nil, errors.New("invalid invite code")
	}

	// 2. Перевірка email
	if _, err := s.userRepo.GetByEmail(input.Email); err == nil {
		return nil, errors.New("email already taken")
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	var user models.User
	var family models.Family
	now := time.Now().UnixMilli()

	// 3. Транзакція: Сім'я + Адмін
	err = s.userRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		// A. Створюємо сім'ю
		family = models.Family{
			Base: models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
			Name: input.Name + "'s Family",
		}
		if err := tx.Create(&family).Error; err != nil {
			return err
		}

		// B. Створюємо юзера
		user = models.User{
			Base:         models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
			FamilyID:     family.ID,
			RoleID:       "admin",
			Name:         input.Name,
			Email:        input.Email,
			PasswordHash: hashedPassword,
			BaseCurrency: "UAH",
			Language:     "uk",
			Theme:        "light",
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 4. СІДІНГ (НАПОВНЕННЯ ДАНИМИ)
	db := s.userRepo.GetDB()
	
	// А. Створюємо категорії (Продукти, Транспорт...)
	if err := utils.SeedFamilyDefaults(db, family.ID); err != nil {
		// Логуємо помилку, але не зупиняємо реєстрацію (не критично)
		// log.Println("Error seeding categories:", err)
	}

	// Б. 🔥 ДОДАВ ЦЕЙ ВИКЛИК: Створюємо контрагентів (Сільпо, АТБ...)
	if err := utils.SeedDefaultCounterparties(db, family.ID); err != nil {
		// log.Println("Error seeding counterparties:", err)
	}

	// 5. Генерація токена
	token, err := utils.GenerateToken(user.ID, user.FamilyID, user.RoleID, s.secretKey)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token, User: user}, nil
}

func (s *authService) Login(input LoginInput) (*LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(input.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !utils.CheckPassword(input.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	token, err := utils.GenerateToken(user.ID, user.FamilyID, user.RoleID, s.secretKey)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token, User: *user}, nil
}