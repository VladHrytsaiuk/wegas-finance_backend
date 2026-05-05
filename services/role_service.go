package services

import (
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
)

type CreateRoleInput struct {
	Name           string
	Description    string
	CanManageUsers bool
	CanEditSchema  bool
}

type RoleService interface {
	Create(input CreateRoleInput) (*models.Role, error)
	GetAll() ([]models.Role, error)
	Delete(id string) error
}

type roleService struct {
	repo repositories.RoleRepository
}

func NewRoleService(repo repositories.RoleRepository) RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) Create(input CreateRoleInput) (*models.Role, error) {
	role := &models.Role{
		Base: models.Base{
			ID:        uuid.NewString(),
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
			IsSynced:  true,
		},
		Name:           input.Name,
		Description:    input.Description,
		CanManageUsers: input.CanManageUsers,
		CanEditSchema:  input.CanEditSchema,
	}

	if err := s.repo.Create(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) GetAll() ([]models.Role, error) {
	return s.repo.GetAll()
}

func (s *roleService) Delete(id string) error {
	return s.repo.Delete(id)
}