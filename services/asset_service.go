package services

import (
	"errors"
	"mime/multipart"
	"path/filepath"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/google/uuid"
)

// Допоміжні структури для статистики
type AssetStats struct {
	TotalExpenses int64 `json:"total_expenses"`
	TotalIncome   int64 `json:"total_income"`
	TCO           int64 `json:"tco"`
	CurrentValue  int64 `json:"current_value"`
}

type AssetWithStats struct {
	*models.Asset
	Stats AssetStats `json:"stats"`
}

type AssetService interface {
	GetAll(user *models.User) ([]models.Asset, error)
	Create(input models.Asset, user *models.User) (string, error)
	GetByID(id string, user *models.User) (*AssetWithStats, error)
	Update(id string, input models.Asset, user *models.User) error
	UpdateMileage(id string, mileage int, user *models.User) error
	Delete(id string, user *models.User) error
	UploadPhoto(id string, file multipart.File, header *multipart.FileHeader, user *models.User) (string, error)
	RemovePhoto(id string, photoPath string, user *models.User) error
	UploadDocument(id string, file multipart.File, header *multipart.FileHeader, user *models.User) (*models.AssetDocument, error)
	RemoveDocument(id string, docID string, user *models.User) error
}

type assetService struct {
	repo    repositories.AssetRepository
	txRepo  repositories.TransactionRepository
	storage StorageService
	clock   utils.Clock
}

func NewAssetService(repo repositories.AssetRepository, txRepo repositories.TransactionRepository, storage StorageService, clock utils.Clock) AssetService {
	return &assetService{repo: repo, txRepo: txRepo, storage: storage, clock: clock}
}

func (s *assetService) Create(input models.Asset, user *models.User) (string, error) {
	input.ID = uuid.NewString()
	input.FamilyID = user.FamilyID
	input.CreatedAt = s.clock.NowUnixMilli()
	input.UserID = user.ID

	if input.Type == "car" {
		if input.InitialMileage == 0 && input.Mileage > 0 {
			input.InitialMileage = input.Mileage
		}
		if input.Mileage == 0 && input.InitialMileage > 0 {
			input.Mileage = input.InitialMileage
		}
	}

	return input.ID, s.repo.Create(&input)
}

func (s *assetService) GetAll(user *models.User) ([]models.Asset, error) {
	return s.repo.GetAll(user.FamilyID)
}

func (s *assetService) GetByID(id string, user *models.User) (*AssetWithStats, error) {
	asset, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return nil, err
	}

	txs, _, err := s.txRepo.GetAll(repositories.TransactionFilter{
		FamilyID: user.FamilyID,
		AssetID:  id,
		Limit:    1000,
	})

	var stats AssetStats
	if err == nil {
		for _, t := range txs {
			if t.Type == "expense" {
				if t.Date > asset.PurchaseDate+1000 {
					stats.TotalExpenses += t.Amount
				}
			} else if t.Type == "income" {
				stats.TotalIncome += t.Amount
			}
		}
	}

	stats.TCO = asset.Price + stats.TotalExpenses - stats.TotalIncome
	stats.CurrentValue = s.CalculateDepreciation(asset)

	return &AssetWithStats{Asset: asset, Stats: stats}, nil
}

func (s *assetService) Update(id string, input models.Asset, user *models.User) error {
	existing, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return err
	}

	existing.Name = input.Name
	existing.Type = input.Type
	existing.SerialNumber = input.SerialNumber
	existing.Price = input.Price
	existing.Currency = input.Currency
	existing.PurchaseDate = input.PurchaseDate
	existing.CurrentPrice = input.CurrentPrice
	existing.WarrantyEnd = input.WarrantyEnd
	existing.IsSold = input.IsSold
	existing.SoldDate = input.SoldDate
	existing.SoldPrice = input.SoldPrice
	existing.Note = input.Note
	existing.DepreciationType = input.DepreciationType
	existing.EstimatedLife = input.EstimatedLife
	existing.InitialValue = input.InitialValue
	existing.VINCode = input.VINCode

	existing.UpdatedAt = s.clock.NowUnixMilli()

	return s.repo.Update(existing)
}

func (s *assetService) UpdateMileage(id string, newMileage int, user *models.User) error {
	existing, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return err
	}
	existing.Mileage = newMileage
	existing.UpdatedAt = s.clock.NowUnixMilli()
	return s.repo.Update(existing)
}

func (s *assetService) Delete(id string, user *models.User) error {
	return s.repo.Delete(id, user.FamilyID)
}

func (s *assetService) UploadPhoto(assetID string, file multipart.File, header *multipart.FileHeader, user *models.User) (string, error) {
	asset, err := s.repo.GetByID(assetID, user.FamilyID)
	if err != nil {
		return "", errors.New("asset not found")
	}

	path, err := s.storage.SaveImage(header, "assets")
	if err != nil {
		return "", err
	}

	if asset.Photo == "" {
		_ = s.repo.UpdatePhoto(assetID, path)
	} else {
		_ = s.repo.AddPhotoToGallery(&models.AssetPhoto{
			Base:    models.Base{ID: uuid.NewString(), CreatedAt: s.clock.NowUnixMilli()},
			AssetID: assetID,
			Path:    path,
		})
	}

	return path, nil
}

func (s *assetService) RemovePhoto(assetID string, path string, user *models.User) error {
	_, err := s.repo.GetByID(assetID, user.FamilyID)
	if err != nil {
		return errors.New("asset not found")
	}

	if err := s.repo.RemovePhoto(assetID, path); err != nil {
		return err
	}

	_ = s.storage.DeleteFile(path)
	return nil
}

// 🔥 НОВИЙ МЕТОД ДЛЯ ЗАВАНТАЖЕННЯ ДОКУМЕНТІВ (PDF тощо)
func (s *assetService) UploadDocument(assetID string, file multipart.File, header *multipart.FileHeader, user *models.User) (*models.AssetDocument, error) {
	// 1. Перевіряємо права
	_, err := s.repo.GetByID(assetID, user.FamilyID)
	if err != nil {
		return nil, errors.New("asset not found")
	}

	// 2. Зберігаємо файл
	path, err := s.storage.SaveFile(header, "documents")
	if err != nil {
		return nil, err
	}

	// 3. Записуємо в БД
	doc := &models.AssetDocument{
		Base:     models.Base{ID: uuid.NewString(), CreatedAt: s.clock.NowUnixMilli()},
		AssetID:  assetID,
		Name:     header.Filename,
		Path:     path,
		FileType: filepath.Ext(header.Filename),
	}

	if err := s.repo.AddDocument(doc); err != nil {
		return nil, err
	}

	return doc, nil
}

// 🔥 НОВИЙ МЕТОД ДЛЯ ВИДАЛЕННЯ ДОКУМЕНТА
func (s *assetService) RemoveDocument(assetID string, documentID string, user *models.User) error {
	// Перевіряємо права на актив
	_, err := s.repo.GetByID(assetID, user.FamilyID)
	if err != nil {
		return errors.New("asset not found")
	}

	doc, err := s.repo.GetDocumentByID(documentID)
	if err != nil {
		return err
	}

	if err := s.repo.RemoveDocument(assetID, documentID); err != nil {
		return err
	}

	_ = s.storage.DeleteFile(doc.Path)

	return nil
}

func (s *assetService) CalculateDepreciation(a *models.Asset) int64 {
	if a.IsSold {
		return a.SoldPrice
	}
	if a.CurrentPrice > 0 {
		return a.CurrentPrice
	}

	if a.Price == 0 {
		return 0
	}

	// Якщо не вказано термін експлуатації - повертаємо ціну покупки
	if a.EstimatedLife <= 0 {
		return a.Price
	}

	// Розрахунок віку в місяцях
	now := s.clock.NowUnixMilli()
	ageMs := now - a.PurchaseDate
	if ageMs <= 0 {
		return a.Price
	}

	ageMonths := float64(ageMs) / (1000 * 60 * 60 * 24 * 30.44)
	
	// Початкова вартість для амортизації
	startingValue := float64(a.Price)
	baseValue := a.Price
	if a.InitialValue > 0 {
		baseValue = a.InitialValue
		startingValue = float64(a.InitialValue)
	}

	// Лінійна амортизація
	depreciation := (float64(baseValue) / float64(a.EstimatedLife)) * ageMonths
	currentValue := startingValue - depreciation

	if currentValue < 0 {
		return 0
	}

	return int64(currentValue)
}
