package repositories

import (
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *models.Category) error
	GetAll(familyID string) ([]models.Category, error)
	GetByID(id string, familyID string) (*models.Category, error)
	Update(category *models.Category) error
	Delete(id string, familyID string) error
}

type categoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepo) GetAll(familyID string) ([]models.Category, error) {
	var categories []models.Category
	// Показуємо тільки ті, що не видалені (Soft Delete)
	err := r.db.Where("family_id = ? AND deleted_at IS NULL", familyID).Find(&categories).Error
	return categories, err
}

func (r *categoryRepo) GetByID(id string, familyID string) (*models.Category, error) {
	var category models.Category
	// Шукаємо по ID та FamilyID для безпеки
	err := r.db.Where("id = ? AND family_id = ?", id, familyID).First(&category).Error
	return &category, err
}

func (r *categoryRepo) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

func (r *categoryRepo) Delete(id string, familyID string) error {
	// Soft Delete: ставимо поточний час у deleted_at
	return r.db.Model(&models.Category{}).
		Where("id = ? AND family_id = ?", id, familyID).
		Update("deleted_at", time.Now().UnixMilli()).Error
}