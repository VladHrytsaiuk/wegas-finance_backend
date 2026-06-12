package repositories

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type StorageTypeRepository interface {
	Create(st *models.StorageType) error
	FindAvailable(familyID string) ([]models.StorageType, error)
	Delete(id string) error
	FindBySlug(slug string) (*models.StorageType, error)
}

type storageTypeRepository struct {
	db *gorm.DB
}

func NewStorageTypeRepository(db *gorm.DB) StorageTypeRepository {
	return &storageTypeRepository{db: db}
}

func (r *storageTypeRepository) Create(st *models.StorageType) error {
	return r.db.Create(st).Error
}

// FindAvailable повертає системні типи + типи конкретної сім'ї
func (r *storageTypeRepository) FindAvailable(familyID string) ([]models.StorageType, error) {
	var types []models.StorageType
	err := r.db.Where("family_id = ? OR is_system = ?", familyID, true).Find(&types).Error
	return types, err
}

func (r *storageTypeRepository) Delete(id string) error {
	return r.db.Delete(&models.StorageType{}, "id = ?", id).Error
}

func (r *storageTypeRepository) FindBySlug(slug string) (*models.StorageType, error) {
	var st models.StorageType
	err := r.db.Where("slug = ?", slug).First(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}