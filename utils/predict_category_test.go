package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPredictCategoryID(t *testing.T) {
	catMap := map[string]string{
		"income_перекази на картку": "cat-income-transfer",
		"income_доходи":             "cat-income-general",
		"income_зарплата":           "cat-salary",
		"income_пасивний дохід":     "cat-passive-income",
		"expense_благодійність":      "cat-charity",
		"expense_допомога рідним":    "cat-family-help",
		"expense_продукти":          "cat-food",
		"expense_кафе та ресторани":  "cat-cafe",
		"expense_покупки":           "cat-shopping",
		"expense_здоров'я":          "cat-health",
		"expense_власне авто":       "cat-car",
		"expense_розваги":           "cat-entertainment",
		"expense_інше":              "cat-other-expense",
		"income_інше":               "cat-other-income",
		"продукти":                  "cat-food-no-prefix",
	}

	tests := []struct {
		name         string
		desc         string
		counterparty string
		mcc          string
		txType       string
		expectedID   string
	}{
		{
			name:       "Salary income",
			desc:       "Зарплата за жовтень",
			txType:     "income",
			expectedID: "cat-salary",
		},
		{
			name:       "Cashback income",
			desc:       "Кешбек",
			txType:     "income",
			expectedID: "cat-passive-income",
		},
		{
			name:       "Transfer income by regex",
			desc:       "від: Олександра",
			txType:     "income",
			expectedID: "cat-income-transfer",
		},
		{
			name:       "Donation by keyword",
			desc:       "На дрони для ЗСУ",
			txType:     "expense",
			expectedID: "cat-charity",
		},
		{
			name:       "Transfer expense by MCC",
			mcc:        "4829",
			txType:     "expense",
			expectedID: "cat-family-help",
		},
		{
			name:       "Grocery by MCC",
			mcc:        "5411",
			txType:     "expense",
			expectedID: "cat-food",
		},
		{
			name:         "Known counterparty - ATB",
			counterparty: "атб",
			txType:       "expense",
			expectedID:   "cat-food",
		},
		{
			name:         "Known counterparty - McDonald's",
			counterparty: "McDonald's",
			txType:       "expense",
			expectedID:   "cat-cafe",
		},
		{
			name:       "Description rule - coffee",
			desc:       "Смачна кава",
			txType:     "expense",
			expectedID: "cat-cafe",
		},
		{
			name:       "Default income",
			desc:       "Unknown",
			txType:     "income",
			expectedID: "cat-other-income",
		},
		{
			name:       "Default expense",
			desc:       "Unknown",
			txType:     "expense",
			expectedID: "cat-other-expense",
		},
		{
			name:       "Empty catMap",
			desc:       "Any",
			txType:     "expense",
			expectedID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := catMap
			if tt.name == "Empty catMap" {
				cm = nil
			}
			id := PredictCategoryID(tt.desc, tt.counterparty, tt.mcc, tt.txType, cm)
			assert.Equal(t, tt.expectedID, id)
		})
	}
}
