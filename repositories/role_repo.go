package repositories

import (
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(role *models.Role) error
	GetAll() ([]models.Role, error)
	Delete(id string) error
}

type roleRepo struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepo{db: db}
}

func (r *roleRepo) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepo) GetAll() ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Where("deleted_at IS NULL").Find(&roles).Error
	return roles, err
}

func (r *roleRepo) Delete(id string) error {
	return r.db.Model(&models.Role{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now().UnixMilli()).Error
}