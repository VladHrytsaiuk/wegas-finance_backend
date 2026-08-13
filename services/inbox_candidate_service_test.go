package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestFindAccountCandidatesByPaymentMask(t *testing.T) {
	accounts := []models.Account{
		{Base: models.Base{ID: "card-1"}, Type: "card", Name: "Mono", Currency: "UAH", CardNumber: "1234", CardNumbers: []string{"1234", "6666"}},
		{Base: models.Base{ID: "card-2"}, Type: "card", Name: "Privat", Currency: "UAH", CardNumber: "7777"},
		{Base: models.Base{ID: "cash-1"}, Type: "cash", Name: "Cash", CardNumber: "6666"},
	}

	t.Run("exact four digit token match is a recommendation", func(t *testing.T) {
		candidates := findAccountCandidates("XXXXXXXXXXXX6666", "UAH", accounts)
		if assert.Len(t, candidates, 1) {
			assert.Equal(t, "card-1", candidates[0].AccountID)
			assert.Equal(t, "6666", candidates[0].MatchedCardNumber)
			assert.Equal(t, "exact", candidates[0].Confidence)
			assert.Equal(t, 70, candidates[0].Score)
			assert.True(t, candidates[0].Recommended)
		}
	})

	t.Run("one digit mask remains a manual suggestion", func(t *testing.T) {
		candidates := findAccountCandidates("XXXXXXXXXXXXXXX7", "UAH", accounts)
		if assert.Len(t, candidates, 1) {
			assert.Equal(t, "card-2", candidates[0].AccountID)
			assert.Equal(t, "partial", candidates[0].Confidence)
			assert.Equal(t, 0, candidates[0].Score)
			assert.False(t, candidates[0].Recommended)
		}
	})

	t.Run("partial mask with trailing separators is still offered for manual selection", func(t *testing.T) {
		candidates := findAccountCandidates("ЕПЗ ••66 **", "UAH", accounts)
		if assert.Len(t, candidates, 1) {
			assert.Equal(t, "card-1", candidates[0].AccountID)
			assert.Equal(t, 2, candidates[0].MatchedDigits)
			assert.False(t, candidates[0].Recommended)
		}
	})
}
