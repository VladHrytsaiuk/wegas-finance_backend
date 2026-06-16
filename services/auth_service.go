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
	Token        string      `json:"token"` // Legacy field
	AccessToken  string      `json:"access_token,omitempty"`
	RefreshToken string      `json:"refresh_token,omitempty"`
	User         models.User `json:"user"`
}

type AuthService interface {
	Register(input RegisterInput) (*LoginResponse, error)
	Login(input LoginInput) (*LoginResponse, error)
	SetPIN(userID, pin string) error
	LoginWithPIN(email, pin string) (*LoginResponse, error)
}

type authService struct {
	userRepo   repositories.UserRepository
	jwtService JWTService // 🔥 Added JWTService
	secretKey  string
	inviteCode string
}

func NewAuthService(userRepo repositories.UserRepository, jwtService JWTService, secretKey, inviteCode string) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtService: jwtService,
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
		// log.Println("Error seeding categories:", err)
	}

	// Б. 🔥 ДОДАВ ЦЕЙ ВИКЛИК: Створюємо контрагентів (Сільпо, АТБ...)
	if err := utils.SeedDefaultCounterparties(db, family.ID); err != nil {
		// log.Println("Error seeding counterparties:", err)
	}

	// 5. Генерація токена
	accessToken, _ := s.jwtService.GenerateAccessToken(user.ID, user.FamilyID, user.RoleID)
	refreshToken, _ := s.jwtService.GenerateRefreshToken(user.ID, user.FamilyID, user.RoleID)

	return &LoginResponse{
		Token:        accessToken,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *authService) Login(input LoginInput) (*LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(input.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !utils.CheckPassword(input.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	accessToken, _ := s.jwtService.GenerateAccessToken(user.ID, user.FamilyID, user.RoleID)
	refreshToken, _ := s.jwtService.GenerateRefreshToken(user.ID, user.FamilyID, user.RoleID)

	return &LoginResponse{
		Token:        accessToken,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *authService) SetPIN(userID, pin string) error {
	if len(pin) != 4 {
		return errors.New("PIN must be exactly 4 digits")
	}

	hashedPin, err := utils.HashPassword(pin) // We can reuse bcrypt for PIN
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.PinHash = hashedPin
	return s.userRepo.Update(user)
}

func (s *authService) LoginWithPIN(email, pin string) (*LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.PinHash == "" {
		return nil, errors.New("PIN login not set up for this user")
	}

	now := time.Now().Unix()

	// 1. Брутфорс: Перевірка блокування (наприклад, 15 хв після 5 спроб)
	if user.PinLockedUntil > now {
		return nil, errors.New("PIN login temporarily locked")
	}

	// 2. Брутфорс: Перевірка інтервалу (2 рази на 5 секунд)
	if now-user.LastPinAttemptAt < 2 {
		return nil, errors.New("too many attempts, wait a second")
	}

	user.LastPinAttemptAt = now

	// 3. Перевірка ПІН
	if !utils.CheckPassword(pin, user.PinHash) {
		user.FailedPinAttempts++
		if user.FailedPinAttempts >= 5 {
			user.PinLockedUntil = now + 900 // 15 хвилин
			user.FailedPinAttempts = 0     // Скидаємо після блокування
		}
		s.userRepo.Update(user)
		return nil, errors.New("invalid PIN")
	}

	// Успішний вхід - скидаємо лічильники
	user.FailedPinAttempts = 0
	user.PinLockedUntil = 0
	s.userRepo.Update(user)

	accessToken, _ := s.jwtService.GenerateAccessToken(user.ID, user.FamilyID, user.RoleID)
	refreshToken, _ := s.jwtService.GenerateRefreshToken(user.ID, user.FamilyID, user.RoleID)

	return &LoginResponse{
		Token:        accessToken,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}
