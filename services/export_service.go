package services

import (
	"errors"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
)

type ExportService interface {
	GetTransactions(user *models.User, filter models.ExportFilterDTO) ([]models.Transaction, error)
	GetBackup(user *models.User) (*models.BackupDTO, error)
}

type exportService struct {
	repo repositories.ExportRepository
}

func NewExportService(repo repositories.ExportRepository) ExportService {
	return &exportService{repo: repo}
}

func (s *exportService) GetTransactions(user *models.User, filter models.ExportFilterDTO) ([]models.Transaction, error) {
	// 1. Валідація дат
	if filter.From < 100000000000 {
		filter.From *= 1000
	}
	if filter.To < 100000000000 {
		filter.To *= 1000
	}

	// 2. 🔥 ЗАХИСТ: Якщо це дитина, вона бачить ТІЛЬКИ свої транзакції
	if user.RoleID == "child" {
		// Ми перезаписуємо UserIDs, ігноруючи те, що прийшло з фронтенду
		filter.UserIDs = []string{user.ID}
	}

	// 3. Виклик репозиторія
	return s.repo.GetTransactionsForExport(user.FamilyID, filter)
}

func (s *exportService) GetBackup(user *models.User) (*models.BackupDTO, error) {
	if user.RoleID == "child" {
		return nil, errors.New("access denied: backups are available only for parents or admins")
	}
	return s.repo.GetBackupData(user.FamilyID, user.ID, false)
}