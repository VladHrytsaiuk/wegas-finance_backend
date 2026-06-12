package services

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
)

type MedicalService interface {
	CreateRecord(user *models.User, record *models.MedicalRecord) error
	GetAllRecords(user *models.User) ([]models.MedicalRecord, error)
	GetRecord(user *models.User, id string) (*models.MedicalRecord, error)
	UpdateRecord(user *models.User, record *models.MedicalRecord) error
	DeleteRecord(user *models.User, id string) error
}

type medicalService struct {
	repo repositories.MedicalRepository
}

func NewMedicalService(repo repositories.MedicalRepository) MedicalService {
	return &medicalService{repo: repo}
}

func (s *medicalService) resolveAccess(user *models.User) (string, string) {
	familyID := user.FamilyID
	userID := ""
	if user.RoleID == "child" {
		userID = user.ID
	}
	return familyID, userID
}

func (s *medicalService) CreateRecord(user *models.User, record *models.MedicalRecord) error {
	record.FamilyID = user.FamilyID
	record.UserID = user.ID
	return s.repo.Create(record)
}

func (s *medicalService) GetAllRecords(user *models.User) ([]models.MedicalRecord, error) {
	familyID, userID := s.resolveAccess(user)
	return s.repo.GetAll(familyID, userID)
}

func (s *medicalService) GetRecord(user *models.User, id string) (*models.MedicalRecord, error) {
	return s.repo.GetByID(id, user.FamilyID)
}

func (s *medicalService) UpdateRecord(user *models.User, record *models.MedicalRecord) error {
	// Security check
	existing, err := s.repo.GetByID(record.ID, user.FamilyID)
	if err != nil {
		return err
	}
	if user.RoleID == "child" && existing.UserID != user.ID {
		return ErrForbidden // Or some custom error
	}
	
	record.FamilyID = user.FamilyID
	return s.repo.Update(record)
}

func (s *medicalService) DeleteRecord(user *models.User, id string) error {
	// Security check
	existing, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return err
	}
	if user.RoleID == "child" && existing.UserID != user.ID {
		return ErrForbidden
	}
	return s.repo.Delete(id, user.FamilyID)
}
