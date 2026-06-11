package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestAccountService_Create(t *testing.T) {
	mockRepo := new(MockAccountRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	// AccountService.Create does not use db for now, but we pass it anyway
	service := NewAccountService(mockRepo, db)

	user := &models.User{
		Base:     models.Base{ID: "user-1"},
		FamilyID: "family-1",
		RoleID:   "parent",
	}

	t.Run("Create account - parent success", func(t *testing.T) {
		input := CreateAccountInput{
			Name:           "Cash",
			Type:           "cash",
			Currency:       "UAH",
			InitialBalance: 1000,
			Color:          "#ffffff",
		}

		mockRepo.On("Create", mock.MatchedBy(func(acc *models.Account) bool {
			return acc.Name == "Cash" && acc.UserID == "user-1" && acc.FamilyID == "family-1"
		})).Return(nil).Once()

		acc, err := service.Create(input, user)

		assert.NoError(t, err)
		assert.NotNil(t, acc)
		assert.Equal(t, "Cash", acc.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Create account - child only for self", func(t *testing.T) {
		childUser := &models.User{
			Base:     models.Base{ID: "child-1"},
			FamilyID: "family-1",
			RoleID:   "child",
		}
		input := CreateAccountInput{
			Name:    "Pocket Money",
			OwnerID: "other-user", // Child tries to create for someone else
		}

		mockRepo.On("Create", mock.MatchedBy(func(acc *models.Account) bool {
			return acc.UserID == "child-1" // Should be forced to child-1
		})).Return(nil).Once()

		acc, err := service.Create(input, childUser)

		assert.NoError(t, err)
		assert.Equal(t, "child-1", acc.UserID)
		mockRepo.AssertExpectations(t)
	})
}
