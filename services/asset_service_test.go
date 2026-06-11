package services

import (
	"testing"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCalculateDepreciation(t *testing.T) {
	// Фіксуємо час: 2026-01-01
	fixedNow := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mockClock := utils.NewMockClock(fixedNow)
	
	service := &AssetService{clock: mockClock}

	tests := []struct {
		name     string
		asset    *models.Asset
		expected int64
	}{
		{
			name: "Sold asset should return sold price",
			asset: &models.Asset{
				IsSold:    true,
				SoldPrice: 5000,
			},
			expected: 5000,
		},
		{
			name: "Asset with manual CurrentPrice should return it",
			asset: &models.Asset{
				IsSold:       false,
				CurrentPrice: 7500,
			},
			expected: 7500,
		},
		{
			name: "New asset (0 age) should return full price",
			asset: &models.Asset{
				Price:        10000,
				PurchaseDate: fixedNow.UnixMilli(),
				EstimatedLife: 60, // 5 years
			},
			expected: 10000,
		},
		{
			name: "Asset after 2.5 years (half life) should return half price",
			asset: &models.Asset{
				Price:        10000,
				PurchaseDate: fixedNow.AddDate(-2, -6, 0).UnixMilli(),
				EstimatedLife: 60, // 5 years
			},
			expected: 5000,
		},
		{
			name: "Asset after full life should return 0",
			asset: &models.Asset{
				Price:        10000,
				PurchaseDate: fixedNow.AddDate(-6, 0, 0).UnixMilli(),
				EstimatedLife: 60, // 5 years
			},
			expected: 0,
		},
		{
			name: "Asset with InitialValue should use it for calculation",
			asset: &models.Asset{
				Price:        20000,
				InitialValue: 10000,
				PurchaseDate: fixedNow.AddDate(-2, -6, 0).UnixMilli(),
				EstimatedLife: 60,
			},
			expected: 5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateDepreciation(tt.asset)
			// Дозволяємо похибку до 20 одиниць (через розрахунки 365.25 днів)
			assert.InDelta(t, tt.expected, result, 20, "Calculation mismatch for: "+tt.name)
		})
	}
}

func TestCreateAsset(t *testing.T) {
	mockRepo := new(MockAssetRepository)
	mockClock := utils.NewMockClock(time.Now())
	service := NewAssetService(mockRepo, nil, nil, mockClock)

	user := &models.User{
		Base:     models.Base{ID: "user-1"},
		FamilyID: "family-1",
	}
	input := models.Asset{
		Name:  "iPhone 15",
		Type:  "electronics",
		Price: 40000,
	}

	// Очікуємо виклик Create один раз
	mockRepo.On("Create", mock.AnythingOfType("*models.Asset")).Return(nil)

	id, err := service.Create(input, user)

	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	
	// Перевіряємо чи всі очікувані виклики відбулися
	mockRepo.AssertExpectations(t)
}
