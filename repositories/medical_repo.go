package repositories

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type MedicalRepository interface {
	Create(record *models.MedicalRecord) error
	GetAll(familyID string, userID string) ([]models.MedicalRecord, error)
	GetByID(id string, familyID string) (*models.MedicalRecord, error)
	Update(record *models.MedicalRecord) error
	Delete(id string, familyID string) error
}

type medicalRepo struct {
	db *gorm.DB
}

func NewMedicalRepository(db *gorm.DB) MedicalRepository {
	return &medicalRepo{db: db}
}

func (r *medicalRepo) Create(record *models.MedicalRecord) error {
	return r.db.Create(record).Error
}

func (r *medicalRepo) GetAll(familyID string, userID string) ([]models.MedicalRecord, error) {
	var records []models.MedicalRecord
	query := r.db.Where("family_id = ?", familyID)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Preload("Files").Order("date desc").Find(&records).Error
	return records, err
}

func (r *medicalRepo) GetByID(id string, familyID string) (*models.MedicalRecord, error) {
	var record models.MedicalRecord
	err := r.db.Preload("Files").Where("id = ? AND family_id = ?", id, familyID).First(&record).Error
	return &record, err
}

func (r *medicalRepo) Update(record *models.MedicalRecord) error {
	return r.db.Save(record).Error
}

func (r *medicalRepo) Delete(id string, familyID string) error {
	return r.db.Where("id = ? AND family_id = ?", id, familyID).Delete(&models.MedicalRecord{}).Error
}
