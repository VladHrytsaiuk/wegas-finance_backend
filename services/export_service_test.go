package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportService(t *testing.T) {
	mockRepo := new(MockExportRepository)
	service := NewExportService(mockRepo)

	t.Run("GetTransactions - Normal User", func(t *testing.T) {
		user := &models.User{Base: models.Base{ID: "u-1"}, FamilyID: "f-1", RoleID: "admin"}
		filter := models.ExportFilterDTO{From: 1600000000, To: 1700000000} // Seconds

		mockRepo.On("GetTransactionsForExport", "f-1", mock.MatchedBy(func(f models.ExportFilterDTO) bool {
			return f.From == 1600000000000 && f.To == 1700000000000
		})).Return([]models.Transaction{{Base: models.Base{ID: "tx-1"}}}, nil).Once()

		res, err := service.GetTransactions(user, filter)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetTransactions - Child Role Protection", func(t *testing.T) {
		user := &models.User{Base: models.Base{ID: "u-child"}, FamilyID: "f-1", RoleID: "child"}
		filter := models.ExportFilterDTO{UserIDs: []string{"other-user"}}

		mockRepo.On("GetTransactionsForExport", "f-1", mock.MatchedBy(func(f models.ExportFilterDTO) bool {
			return len(f.UserIDs) == 1 && f.UserIDs[0] == "u-child"
		})).Return([]models.Transaction{}, nil).Once()

		_, err := service.GetTransactions(user, filter)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

// MockExportRepository is needed here. I'll add it to mocks.go if it's missing.
