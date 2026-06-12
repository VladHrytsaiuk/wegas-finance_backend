package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseAmountString(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"100", 10000, false},
		{"100.50", 10050, false},
		{"1 000,50", 100050, false},
		{"-1.234,56", -123456, false},
		{"1234,56-", -123456, false},
		{"–50,00", -5000, false}, // En dash
		{"—30,00", -3000, false}, // Em dash
		{"1.250.000,00", 125000000, false},
		{"0,05", 5, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		val, err := ParseAmountString(tt.input)
		if tt.hasError {
			assert.Error(t, err, "input: %s", tt.input)
		} else {
			assert.NoError(t, err, "input: %s", tt.input)
			assert.Equal(t, tt.expected, val, "input: %s", tt.input)
		}
	}
}

func TestParseBankDate(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Kiev")

	tests := []struct {
		input    string
		expected time.Time
		hasError bool
	}{
		{"12.06.2026 15:30", time.Date(2026, 6, 12, 15, 30, 0, 0, loc), false},
		{"12.06.2026", time.Date(2026, 6, 12, 0, 0, 0, 0, loc), false},
		{"2026-06-12 10:00:00", time.Date(2026, 6, 12, 10, 0, 0, 0, loc), false},
		{"12/06/2026", time.Date(2026, 6, 12, 0, 0, 0, 0, loc), false},
		{"01.01.1990", time.Time{}, true}, // Year < 2000
		{"invalid", time.Time{}, true},
		{"", time.Time{}, true},
	}

	for _, tt := range tests {
		val, err := ParseBankDate(tt.input)
		if tt.hasError {
			assert.Error(t, err, "input: %s", tt.input)
		} else {
			assert.NoError(t, err, "input: %s", tt.input)
			assert.True(t, tt.expected.Equal(val), "input: %s", tt.input)
		}
	}
}
