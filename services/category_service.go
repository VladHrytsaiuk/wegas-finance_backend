package services

import (
	"errors"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
)

type CategoryInput struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	ParentID string `json:"parent_id"`
}

type CategoryService interface {
	Create(input CategoryInput, user *models.User) (*models.Category, error)
	GetAll(user *models.User) ([]models.Category, error)
	GetByID(id string, user *models.User) (*models.Category, error)
	Update(id string, input CategoryInput, user *models.User) (*models.Category, error)
	Delete(id string, user *models.User) error
}

type categoryService struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

// Помилка для перевикористання
var ErrAccessDenied = errors.New("access denied: only parents can manage categories")

func (s *categoryService) Create(input CategoryInput, user *models.User) (*models.Category, error) {
	// 🛑 ЗАХИСТ: Діти не можуть створювати категорії
	if user.RoleID == "child" {
		return nil, ErrAccessDenied
	}

	category := &models.Category{
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
		ParentID: input.ParentID,
	}

	if err := s.repo.Create(category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *categoryService) GetAll(user *models.User) ([]models.Category, error) {
	// ✅ ЧИТАННЯ: Дозволено всім членам сім'ї
	return s.repo.GetAll(user.FamilyID)
}

func (s *categoryService) GetByID(id string, user *models.User) (*models.Category, error) {
	// ✅ ЧИТАННЯ: Дозволено всім
	return s.repo.GetByID(id, user.FamilyID)
}

func (s *categoryService) Update(id string, input CategoryInput, user *models.User) (*models.Category, error) {
	// 🛑 ЗАХИСТ
	if user.RoleID == "child" {
		return nil, ErrAccessDenied
	}

	existing, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return nil, errors.New("category not found")
	}

	existing.Name = input.Name
	existing.Type = input.Type
	existing.Icon = input.Icon
	existing.Color = input.Color
	existing.ParentID = input.ParentID
	existing.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *categoryService) Delete(id string, user *models.User) error {
	// 🛑 ЗАХИСТ
	if user.RoleID == "child" {
		return ErrAccessDenied
	}

	_, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return errors.New("category not found")
	}
	return s.repo.Delete(id, user.FamilyID)
}