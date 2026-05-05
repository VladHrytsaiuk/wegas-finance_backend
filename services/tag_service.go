package services

import (
	"errors"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
)

type TagService interface {
	Create(name string, color string, user *models.User) (*models.Tag, error)
	GetAll(familyID string) ([]models.Tag, error)
	Delete(id string, user *models.User) error
}

type tagService struct {
	repo repositories.TagRepository
}

func NewTagService(repo repositories.TagRepository) TagService {
	return &tagService{repo: repo}
}

func (s *tagService) Create(name string, color string, user *models.User) (*models.Tag, error) {
	// ✅ БЕЗ ОБМЕЖЕНЬ: Дитина теж може створювати теги
	tag := &models.Tag{
		Base: models.Base{
			ID:        uuid.NewString(),
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
			IsSynced:  true,
		},
		FamilyID: user.FamilyID,
		Name:     name,
		Color:    color,
	}

	if err := s.repo.Create(tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *tagService) GetAll(familyID string) ([]models.Tag, error) {
	return s.repo.GetAll(familyID)
}

func (s *tagService) Delete(id string, user *models.User) error {
	// ✅ БЕЗ ОБМЕЖЕНЬ: Дитина може видаляти теги
	// (Можна додати логіку, що видаляти можна тільки теги, які не використовуються іншими, 
	// але поки що залишимо просто перевірку на існування)
	
	_, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return errors.New("tag not found")
	}
	return s.repo.Delete(id, user.FamilyID)
}