package services

import (
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
)

type AssetService struct {
	repo    repositories.AssetRepository
	txRepo  repositories.TransactionRepository
	storage StorageService
}

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

func NewAssetService(repo repositories.AssetRepository, txRepo repositories.TransactionRepository, storage StorageService) *AssetService {
	return &AssetService{repo: repo, txRepo: txRepo, storage: storage}
}

func (s *AssetService) Create(input models.Asset, user *models.User) (string, error) {
	input.ID = uuid.NewString()
	input.FamilyID = user.FamilyID
	input.CreatedAt = time.Now().UnixMilli()
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

func (s *AssetService) GetAll(user *models.User) ([]models.Asset, error) {
	return s.repo.GetAll(user.FamilyID)
}

func (s *AssetService) GetByID(id string, user *models.User) (*AssetWithStats, error) {
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

func (s *AssetService) Update(id string, input models.Asset, user *models.User) error {
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
	existing.Mileage = input.Mileage
	existing.InitialMileage = input.InitialMileage
	existing.InsuranceExpiry = input.InsuranceExpiry
	existing.LastServiceDate = input.LastServiceDate
	existing.Address = input.Address
	existing.Area = input.Area
	existing.CadastralNum = input.CadastralNum
	existing.UpdatedAt = time.Now().UnixMilli()

	return s.repo.Update(existing)
}

func (s *AssetService) UpdateMileage(id string, newMileage int, user *models.User) error {
	existing, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return err
	}
	existing.Mileage = newMileage
	existing.UpdatedAt = time.Now().UnixMilli()
	return s.repo.Update(existing)
}

func (s *AssetService) Delete(id string, user *models.User) error {
	return s.repo.Delete(id, user.FamilyID)
}

func (s *AssetService) UploadPhoto(assetID string, file multipart.File, header *multipart.FileHeader, user *models.User) (string, error) {
	asset, err := s.repo.GetByID(assetID, user.FamilyID)
	if err != nil {
		return "", errors.New("asset not found")
	}

	path, err := s.storage.SaveImage(header, "assets")
	if err != nil {
		return "", err
	}

	galleryPhoto := &models.AssetPhoto{
		Base:    models.Base{ID: uuid.NewString(), CreatedAt: time.Now().UnixMilli(), UpdatedAt: time.Now().UnixMilli()},
		AssetID: assetID,
		Path:    path,
	}

	if err := s.repo.AddPhotoToGallery(galleryPhoto); err != nil {
		_ = s.storage.DeleteFile(path)
		return "", err
	}

	if asset.Photo == "" {
		_ = s.repo.UpdatePhoto(assetID, path)
	}

	return path, nil
}

func (s *AssetService) RemovePhoto(assetID string, path string, user *models.User) error {
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
func (s *AssetService) UploadDocument(assetID string, file multipart.File, header *multipart.FileHeader, user *models.User) (*models.AssetDocument, error) {
	// 1. Перевіряємо права
	_, err := s.repo.GetByID(assetID, user.FamilyID)
	if err != nil {
		return nil, errors.New("asset not found")
	}

	// 2. Валідація розміру (наприклад, макс 10 МБ)
	const maxUploadSize = 10 << 20 // 10 MB
	if header.Size > maxUploadSize {
		return nil, errors.New("file too large (max 10MB)")
	}

	// 3. Зберігаємо оригінальну назву і розширення
	originalExt := strings.ToLower(filepath.Ext(header.Filename))
	path, err := s.storage.SaveFile(header, "documents")
	if err != nil {
		return nil, err
	}

	// 4. Створюємо запис в БД
	doc := &models.AssetDocument{
		Base: models.Base{
			ID:        uuid.NewString(),
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
		},
		AssetID:  assetID,
		Name:     header.Filename, // Оригінальна назва файлу, яку побачить користувач
		Path:     path,
		FileType: strings.TrimPrefix(originalExt, "."),
		Size:     header.Size,
	}

	if err := s.repo.AddDocument(doc); err != nil {
		_ = s.storage.DeleteFile(path) // Відкат, якщо помилка БД
		return nil, err
	}

	return doc, nil
}

// 🔥 НОВИЙ МЕТОД ДЛЯ ВИДАЛЕННЯ ДОКУМЕНТА
func (s *AssetService) RemoveDocument(assetID string, documentID string, user *models.User) error {
	// Перевіряємо права на актив
	_, err := s.repo.GetByID(assetID, user.FamilyID)
	if err != nil {
		return errors.New("asset not found")
	}

	doc, err := s.repo.GetDocumentByID(documentID)
	if err != nil {
		return errors.New("document not found")
	}

	if err := s.repo.RemoveDocument(assetID, documentID); err != nil {
		return err
	}

	// Фізично видаляємо файл
	_ = s.storage.DeleteFile(doc.Path)

	return nil
}

func (s *AssetService) CalculateDepreciation(a *models.Asset) int64 {
	if a.IsSold {
		return a.SoldPrice
	}
	if a.CurrentPrice > 0 {
		return a.CurrentPrice
	}

	now := time.Now().UnixMilli()
	ageInYears := float64(now-a.PurchaseDate) / (1000 * 60 * 60 * 24 * 365.25)

	life := 5.0
	if a.EstimatedLife > 0 {
		life = float64(a.EstimatedLife) / 12.0
	}
	if ageInYears > life {
		return 0
	}

	val := float64(a.Price)
	if a.InitialValue > 0 {
		val = float64(a.InitialValue)
	}

	depreciationRate := 1.0 / life
	currentValue := val * (1 - (ageInYears * depreciationRate))

	if currentValue < 0 {
		return 0
	}
	return int64(currentValue)
}
