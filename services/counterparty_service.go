package services

import (
	"errors"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
)

// Ліміт 35KB.
// Цього вистачить для Base64 рядка (який представляє файл ~25KB)
// І тим більше вистачить для звичайної назви файлу типу "atb.svg".
const MaxLogoLength = 35 * 1024

type CpCategoryInput struct {
	Name  string
	Type  string
	Icon  string
	Color string
}

type CounterpartyInput struct {
	Name       string
	Type       string
	CategoryID *string
	Icon       string
	Logo       string // Тут може бути ім'я файлу ("atb.svg") АБО Base64 рядок
}

type CounterpartyService interface {
	// --- Categories ---
	CreateCategory(input CpCategoryInput, user *models.User) (*models.CounterpartyCategory, error)
	GetCategories(user *models.User) ([]models.CounterpartyCategory, error)
	UpdateCategory(id string, input CpCategoryInput, user *models.User) (*models.CounterpartyCategory, error)
	GetCategoryByID(id string, user *models.User) (*models.CounterpartyCategory, error)

	// --- Counterparties ---
	Create(input CounterpartyInput, user *models.User) (*models.Counterparty, error)
	GetAll(user *models.User) ([]models.Counterparty, error)
	Update(id string, input CounterpartyInput, user *models.User) (*models.Counterparty, error)
	Delete(id string, user *models.User) error
	GetByID(id string, user *models.User) (*models.Counterparty, error)
}

type cpService struct {
	repo repositories.CounterpartyRepository
}

func NewCounterpartyService(repo repositories.CounterpartyRepository) CounterpartyService {
	return &cpService{repo: repo}
}

var ErrCpAccessDenied = errors.New("access denied: only parents can manage counterparties")
var ErrLogoTooLarge = errors.New("logo file is too large (max 20KB)")

// === HELPER ===
func validateLogo(logo string) error {
	// Перевіряємо довжину рядка.
	// Якщо це ім'я файлу (коротке) - пройде.
	// Якщо це Base64 (довге) - перевіримо, чи не перевищує ліміт.
	if len(logo) > MaxLogoLength {
		return ErrLogoTooLarge
	}
	return nil
}

// === CATEGORIES IMPLEMENTATION ===

func (s *cpService) CreateCategory(input CpCategoryInput, user *models.User) (*models.CounterpartyCategory, error) {
	if user.RoleID == "child" {
		return nil, ErrCpAccessDenied
	}

	cat := &models.CounterpartyCategory{
		Base: models.Base{
			ID:        uuid.NewString(),
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
			IsSynced:  true,
		},
		FamilyID: user.FamilyID,
		Name:     input.Name,
		Type:     input.Type,
		Icon:     input.Icon,
		Color:    input.Color,
	}
	if err := s.repo.CreateCategory(cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *cpService) GetCategories(user *models.User) ([]models.CounterpartyCategory, error) {
	return s.repo.GetCategories(user.FamilyID)
}

func (s *cpService) GetCategoryByID(id string, user *models.User) (*models.CounterpartyCategory, error) {
	return s.repo.GetCategoryByID(id, user.FamilyID)
}

func (s *cpService) UpdateCategory(id string, input CpCategoryInput, user *models.User) (*models.CounterpartyCategory, error) {
	if user.RoleID == "child" {
		return nil, ErrCpAccessDenied
	}

	cat, err := s.repo.GetCategoryByID(id, user.FamilyID)
	if err != nil {
		return nil, errors.New("category not found")
	}

	cat.Name = input.Name
	cat.Icon = input.Icon
	cat.Color = input.Color
	cat.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.UpdateCategory(cat); err != nil {
		return nil, err
	}
	return cat, nil
}

// === COUNTERPARTIES IMPLEMENTATION ===

func (s *cpService) Create(input CounterpartyInput, user *models.User) (*models.Counterparty, error) {
	if user.RoleID == "child" {
		return nil, ErrCpAccessDenied
	}

	// Валідація логотипа
	if err := validateLogo(input.Logo); err != nil {
		return nil, err
	}

	cp := &models.Counterparty{
		Base: models.Base{
			ID:        uuid.NewString(),
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
			IsSynced:  true,
		},
		FamilyID:   user.FamilyID,
		Name:       input.Name,
		Type:       input.Type,
		CategoryID: input.CategoryID,
		Icon:       input.Icon,
		Logo:       input.Logo,
	}

	if err := s.repo.Create(cp); err != nil {
		return nil, err
	}

	return s.repo.GetByID(cp.ID, user.FamilyID)
}

func (s *cpService) GetAll(user *models.User) ([]models.Counterparty, error) {
	return s.repo.GetAll(user.FamilyID)
}

func (s *cpService) GetByID(id string, user *models.User) (*models.Counterparty, error) {
	return s.repo.GetByID(id, user.FamilyID)
}

func (s *cpService) Update(id string, input CounterpartyInput, user *models.User) (*models.Counterparty, error) {
	if user.RoleID == "child" {
		return nil, ErrCpAccessDenied
	}

	// Валідація логотипа
	if err := validateLogo(input.Logo); err != nil {
		return nil, err
	}

	cp, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return nil, errors.New("counterparty not found")
	}

	cp.Name = input.Name
	cp.Type = input.Type
	cp.CategoryID = input.CategoryID
	cp.Icon = input.Icon
	cp.Logo = input.Logo
	cp.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.Update(cp); err != nil {
		return nil, err
	}

	return s.repo.GetByID(cp.ID, user.FamilyID)
}

func (s *cpService) Delete(id string, user *models.User) error {
	if user.RoleID == "child" {
		return ErrCpAccessDenied
	}

	_, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return errors.New("counterparty not found")
	}
	return s.repo.Delete(id, user.FamilyID)
}