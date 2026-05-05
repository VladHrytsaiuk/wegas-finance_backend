package repositories

import (
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByEmail(email string) (*models.User, error)
	GetByID(id string) (*models.User, error)
	Update(user *models.User) error
	GetFamilyMembers(familyID string) ([]models.User, error)
	Delete(id string) error

	CreateFamily(family *models.Family) error
	GetFamilyByID(id string) (*models.Family, error)
	GetDB() *gorm.DB
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// === USER ===

func (r *userRepo) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepo) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	return &user, err
}

func (r *userRepo) GetByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	return &user, err
}

func (r *userRepo) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepo) GetFamilyMembers(familyID string) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("family_id = ? AND deleted_at IS NULL", familyID).
		Order("created_at asc").
		Find(&users).Error
	return users, err
}

func (r *userRepo) Delete(id string) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now().UnixMilli()).Error
}

// === FAMILY ===

func (r *userRepo) CreateFamily(family *models.Family) error {
	return r.db.Create(family).Error
}

func (r *userRepo) GetFamilyByID(id string) (*models.Family, error) {
	var family models.Family
	err := r.db.Where("id = ?", id).First(&family).Error
	return &family, err
}

func (r *userRepo) GetDB() *gorm.DB {
	return r.db
}