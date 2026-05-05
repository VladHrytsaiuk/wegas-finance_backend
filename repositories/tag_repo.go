package repositories

import (
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type TagRepository interface {
	Create(tag *models.Tag) error
	GetAll(familyID string) ([]models.Tag, error)
	Delete(id string, familyID string) error
	GetByID(id string, familyID string) (*models.Tag, error)
}

type tagRepo struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepo{db: db}
}

func (r *tagRepo) Create(tag *models.Tag) error {
	return r.db.Create(tag).Error
}

func (r *tagRepo) GetAll(familyID string) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.Where("family_id = ? AND deleted_at IS NULL", familyID).Find(&tags).Error
	return tags, err
}

func (r *tagRepo) GetByID(id string, familyID string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.Where("id = ? AND family_id = ?", id, familyID).First(&tag).Error
	return &tag, err
}

func (r *tagRepo) Delete(id string, familyID string) error {
	return r.db.Model(&models.Tag{}).
		Where("id = ? AND family_id = ?", id, familyID).
		Update("deleted_at", time.Now().UnixMilli()).Error
}