package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCategoryByMCC(t *testing.T) {
	tests := []struct {
		mcc      string
		expected string
		found    bool
	}{
		{"5411", "Продукти", true},
		{"5812", "Кафе та Ресторани", true},
		{"5541", "Власне Авто", true},
		{"4111", "Громадський Транспорт", true},
		{"4112", "Подорожі", true},
		{"3000", "Подорожі", true}, // Boundary of range
		{"3150", "Подорожі", true}, // Middle of range
		{"3299", "Подорожі", true}, // Boundary of range
		{"5912", "Здоров'я", true},
		{"5399", "Покупки", true},
		{"5310", "Дім та Побут", true},
		{"4900", "Житло", true},
		{"4814", "Зв'язок та Інтернет", true},
		{"7832", "Розваги", true},
		{"8211", "Освіта", true},
		{"4215", "Послуги та Сервіс", true},
		{"4829", "Фінанси та Допомога", true},
		{"9999", "", false}, // Unknown MCC
		{"", "", false},     // Empty string
		{"123", "", false},  // Too short
		{"12345", "", false}, // Too long
	}

	for _, tt := range tests {
		t.Run(tt.mcc, func(t *testing.T) {
			category, found := GetCategoryByMCC(tt.mcc)
			assert.Equal(t, tt.expected, category)
			assert.Equal(t, tt.found, found)
		})
	}
}
