package repositories

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type AssetRepository interface {
  Create(asset *models.Asset) error
  GetAll(familyID string) ([]models.Asset, error)
  GetByID(id string, familyID string) (*models.Asset, error)
  Update(asset *models.Asset) error
  Delete(id string, familyID string) error
  
  // Фото
  RemovePhoto(assetID string, path string) error
  UpdatePhoto(assetID string, path string) error
  AddPhotoToGallery(photo *models.AssetPhoto) error
  
  // 🔥 ДОКУМЕНТИ
  AddDocument(doc *models.AssetDocument) error
  RemoveDocument(assetID string, documentID string) error
  GetDocumentByID(documentID string) (*models.AssetDocument, error)
  
  // Спеціальний метод для оновлення сервісних даних
  UpdateServiceData(assetID string, mileage int, date int64) error
}

type assetRepository struct {
  db *gorm.DB
}

func NewAssetRepository(db *gorm.DB) AssetRepository {
  return &assetRepository{db: db}
}

func (r *assetRepository) Create(asset *models.Asset) error {
  return r.db.Create(asset).Error
}

func (r *assetRepository) GetAll(familyID string) ([]models.Asset, error) {
  var assets []models.Asset
  // 🔥 Додано Preload("Documents")
  err := r.db.Preload("Photos").Preload("Documents").Where("family_id = ?", familyID).Find(&assets).Error
  return assets, err
}

func (r *assetRepository) GetByID(id string, familyID string) (*models.Asset, error) {
  var asset models.Asset
  // 🔥 Додано Preload("Documents")
  err := r.db.Preload("Photos").Preload("Documents").Where("id = ? AND family_id = ?", id, familyID).First(&asset).Error
  return &asset, err
}

func (r *assetRepository) Update(asset *models.Asset) error {
  return r.db.Save(asset).Error
}

func (r *assetRepository) Delete(id string, familyID string) error {
  return r.db.Where("id = ? AND family_id = ?", id, familyID).Delete(&models.Asset{}).Error
}

func (r *assetRepository) UpdatePhoto(assetID string, path string) error {
  return r.db.Model(&models.Asset{}).Where("id = ?", assetID).Update("photo", path).Error
}

func (r *assetRepository) AddPhotoToGallery(photo *models.AssetPhoto) error {
  return r.db.Create(photo).Error
}

func (r *assetRepository) RemovePhoto(assetID string, path string) error {
  r.db.Model(&models.Asset{}).Where("id = ? AND photo = ?", assetID, path).Update("photo", "")
  return r.db.Where("asset_id = ? AND path = ?", assetID, path).Delete(&models.AssetPhoto{}).Error
}

// 🔥 РЕАЛІЗАЦІЯ ДЛЯ ДОКУМЕНТІВ
func (r *assetRepository) AddDocument(doc *models.AssetDocument) error {
  return r.db.Create(doc).Error
}

func (r *assetRepository) GetDocumentByID(documentID string) (*models.AssetDocument, error) {
  var doc models.AssetDocument
  err := r.db.Where("id = ?", documentID).First(&doc).Error
  return &doc, err
}

func (r *assetRepository) RemoveDocument(assetID string, documentID string) error {
  return r.db.Where("id = ? AND asset_id = ?", documentID, assetID).Delete(&models.AssetDocument{}).Error
}

func (r *assetRepository) UpdateServiceData(assetID string, mileage int, date int64) error {
  updates := map[string]interface{}{}
  if mileage > 0 { updates["mileage"] = mileage }
  if date > 0 { updates["last_service_date"] = date }
  if len(updates) == 0 { return nil }
  return r.db.Model(&models.Asset{}).Where("id = ?", assetID).Updates(updates).Error
}