package services

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/mock"
)

type MockAssetRepository struct {
	mock.Mock
}

func (m *MockAssetRepository) Create(asset *models.Asset) error {
	args := m.Called(asset)
	return args.Error(0)
}

func (m *MockAssetRepository) GetAll(familyID string) ([]models.Asset, error) {
	args := m.Called(familyID)
	return args.Get(0).([]models.Asset), args.Error(1)
}

func (m *MockAssetRepository) GetByID(id string, familyID string) (*models.Asset, error) {
	args := m.Called(id, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Asset), args.Error(1)
}

func (m *MockAssetRepository) Update(asset *models.Asset) error {
	args := m.Called(asset)
	return args.Error(0)
}

func (m *MockAssetRepository) Delete(id string, familyID string) error {
	args := m.Called(id, familyID)
	return args.Error(0)
}

func (m *MockAssetRepository) UpdatePhoto(assetID string, photoPath string) error {
	args := m.Called(assetID, photoPath)
	return args.Error(0)
}

func (m *MockAssetRepository) AddPhotoToGallery(photo *models.AssetPhoto) error {
	args := m.Called(photo)
	return args.Error(0)
}

func (m *MockAssetRepository) RemovePhoto(assetID string, photoPath string) error {
	args := m.Called(assetID, photoPath)
	return args.Error(0)
}

func (m *MockAssetRepository) AddDocument(doc *models.AssetDocument) error {
	args := m.Called(doc)
	return args.Error(0)
}

func (m *MockAssetRepository) RemoveDocument(assetID string, docID string) error {
	args := m.Called(assetID, docID)
	return args.Error(0)
}

func (m *MockAssetRepository) GetDocumentByID(docID string) (*models.AssetDocument, error) {
	args := m.Called(docID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AssetDocument), args.Error(1)
}

func (m *MockAssetRepository) UpdateServiceData(assetID string, mileage int, lastService int64) error {
	args := m.Called(assetID, mileage, lastService)
	return args.Error(0)
}
