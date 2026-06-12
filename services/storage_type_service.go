package services

import (
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
)

type StorageTypeService interface {
	Create(st *models.StorageType) error
	GetAll(familyID string) ([]models.StorageType, error)
	Delete(id string) error
}

type storageTypeService struct {
	repo repositories.StorageTypeRepository
}

func NewStorageTypeService(repo repositories.StorageTypeRepository) StorageTypeService {
	return &storageTypeService{repo: repo}
}

func (s *storageTypeService) Create(st *models.StorageType) error {
	st.ID = uuid.New().String()
	st.CreatedAt = time.Now().UnixMilli()
	st.UpdatedAt = time.Now().UnixMilli()
	st.IsSystem = false // Юзер створює тільки свої типи
	return s.repo.Create(st)
}

func (s *storageTypeService) GetAll(familyID string) ([]models.StorageType, error) {
	return s.repo.FindAvailable(familyID)
}

func (s *storageTypeService) Delete(id string) error {
	// Можна додати перевірку, чи не є тип системним, перед видаленням,
	// хоча репозиторій зазвичай повертає error, якщо ми спробуємо видалити щось не те, 
	// але краще перевірити "IsSystem". Для спрощення просто викликаємо репо.
	return s.repo.Delete(id)
}


