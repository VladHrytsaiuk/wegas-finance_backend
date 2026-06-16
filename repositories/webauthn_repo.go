package repositories

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type WebAuthnRepository interface {
	CreateCredential(cred *models.WebAuthnCredential) error
	GetCredentialsByUserID(userID string) ([]models.WebAuthnCredential, error)
	GetCredentialByID(credentialID []byte) (*models.WebAuthnCredential, error)
	UpdateCredential(cred *models.WebAuthnCredential) error
}

type webAuthnRepository struct {
	db *gorm.DB
}

func NewWebAuthnRepository(db *gorm.DB) WebAuthnRepository {
	return &webAuthnRepository{db: db}
}

func (r *webAuthnRepository) CreateCredential(cred *models.WebAuthnCredential) error {
	return r.db.Create(cred).Error
}

func (r *webAuthnRepository) GetCredentialsByUserID(userID string) ([]models.WebAuthnCredential, error) {
	var creds []models.WebAuthnCredential
	err := r.db.Where("user_id = ?", userID).Find(&creds).Error
	return creds, err
}

func (r *webAuthnRepository) GetCredentialByID(credentialID []byte) (*models.WebAuthnCredential, error) {
	var cred models.WebAuthnCredential
	err := r.db.Where("credential_id = ?", credentialID).First(&cred).Error
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *webAuthnRepository) UpdateCredential(cred *models.WebAuthnCredential) error {
	return r.db.Save(cred).Error
}
