package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestAccountService_Security(t *testing.T) {
	mockRepo := new(MockAccountRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	service := NewAccountService(mockRepo, db)

	familyA := "family-A"
	familyB := "family-B"
	
	parentA := &models.User{Base: models.Base{ID: "parent-A"}, FamilyID: familyA, RoleID: "admin"}
	childA := &models.User{Base: models.Base{ID: "child-A"}, FamilyID: familyA, RoleID: "child"}
	
	accountA := &models.Account{Base: models.Base{ID: "acc-A"}, FamilyID: familyA, UserID: "parent-A"}
	accountChildA := &models.Account{Base: models.Base{ID: "acc-child-A"}, FamilyID: familyA, UserID: "child-A"}
	accountB := &models.Account{Base: models.Base{ID: "acc-B"}, FamilyID: familyB, UserID: "some-user-B"}

	t.Run("Child cannot view parent account", func(t *testing.T) {
		mockRepo.On("GetByID", "acc-A").Return(accountA, nil).Once()
		
		res, err := service.GetByID("acc-A", childA)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
		assert.Nil(t, res)
	})

	t.Run("User cannot view other family account", func(t *testing.T) {
		mockRepo.On("GetByID", "acc-B").Return(accountB, nil).Once()
		
		res, err := service.GetByID("acc-B", parentA)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
		assert.Nil(t, res)
	})

	t.Run("Child can view own account", func(t *testing.T) {
		mockRepo.On("GetByID", "acc-child-A").Return(accountChildA, nil).Once()
		
		res, err := service.GetByID("acc-child-A", childA)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("Parent can view child account", func(t *testing.T) {
		mockRepo.On("GetByID", "acc-child-A").Return(accountChildA, nil).Once()
		
		res, err := service.GetByID("acc-child-A", parentA)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}

func TestCategoryService_Security(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	familyA := "family-A"
	parentA := &models.User{Base: models.Base{ID: "parent-A"}, FamilyID: familyA, RoleID: "admin"}
	childA := &models.User{Base: models.Base{ID: "child-A"}, FamilyID: familyA, RoleID: "child"}

	t.Run("Child cannot create category", func(t *testing.T) {
		res, err := service.Create(CategoryInput{Name: "Forbidden"}, childA)
		assert.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
		assert.Nil(t, res)
	})

	t.Run("Parent can create category", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything).Return(nil).Once()
		res, err := service.Create(CategoryInput{Name: "Allowed"}, parentA)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}
