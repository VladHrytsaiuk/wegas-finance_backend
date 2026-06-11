package repositories

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type FamilyJoinRepository interface {
	CreateCode(code *models.FamilyJoinCode) error
	GetCode(code string) (*models.FamilyJoinCode, error)
	DeleteCode(code string) error
	UpdateUserFamily(userID string, familyID string, roleID string) error
}

type familyJoinRepo struct {
	db *gorm.DB
}

func NewFamilyJoinRepository(db *gorm.DB) FamilyJoinRepository {
	return &familyJoinRepo{db: db}
}

func (r *familyJoinRepo) CreateCode(code *models.FamilyJoinCode) error {
	// Спочатку видалимо старі коди для цієї сім'ї, щоб не накопичувались
	r.db.Where("family_id = ?", code.FamilyID).Delete(&models.FamilyJoinCode{})
	return r.db.Create(code).Error
}

func (r *familyJoinRepo) GetCode(code string) (*models.FamilyJoinCode, error) {
	var joinCode models.FamilyJoinCode
	err := r.db.Where("code = ?", code).First(&joinCode).Error
	return &joinCode, err
}

func (r *familyJoinRepo) DeleteCode(code string) error {
	return r.db.Where("code = ?", code).Delete(&models.FamilyJoinCode{}).Error
}

func (r *familyJoinRepo) UpdateUserFamily(userID string, familyID string, roleID string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"family_id": familyID,
		"role_id":   roleID,
	}).Error
}
