package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCardDetails(t *testing.T) {
	tests := []struct {
		name          string
		maskedPans    []string
		expectedLast4 string
		expectedPS    string
	}{
		{
			name:          "Visa card",
			maskedPans:    []string{"444455******1234"},
			expectedLast4: "1234",
			expectedPS:    "visa",
		},
		{
			name:          "Mastercard (starts with 5)",
			maskedPans:    []string{"512345******5678"},
			expectedLast4: "5678",
			expectedPS:    "mastercard",
		},
		{
			name:          "Mastercard (starts with 2)",
			maskedPans:    []string{"222100******9999"},
			expectedLast4: "9999",
			expectedPS:    "mastercard",
		},
		{
			name:          "Empty list",
			maskedPans:    []string{},
			expectedLast4: "",
			expectedPS:    "mastercard",
		},
		{
			name:          "Short PAN",
			maskedPans:    []string{"4123"},
			expectedLast4: "4123",
			expectedPS:    "visa",
		},
		{
			name:          "Very short PAN",
			maskedPans:    []string{"5"},
			expectedLast4: "0000",
			expectedPS:    "mastercard",
		},
		{
			name:          "Unknown system (default mastercard)",
			maskedPans:    []string{"612345******0000"},
			expectedLast4: "0000",
			expectedPS:    "mastercard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			last4, ps := ParseCardDetails(tt.maskedPans)
			assert.Equal(t, tt.expectedLast4, last4)
			assert.Equal(t, tt.expectedPS, ps)
		})
	}
}
