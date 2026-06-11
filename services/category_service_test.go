package services

import (
	"errors"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCategoryService_Create(t *testing.T) {
	parentUser := &models.User{Base: models.Base{ID: "user1"}, FamilyID: "fam1", RoleID: "admin"}
	childUser := &models.User{Base: models.Base{ID: "user2"}, FamilyID: "fam1", RoleID: "child"}

	tests := []struct {
		name          string
		input         CategoryInput
		user          *models.User
		mockExpect    func(m *MockCategoryRepository)
		wantErr       bool
		expectedError error
	}{
		{
			name: "Success as admin",
			input: CategoryInput{
				Name: "Food",
				Type: "expense",
			},
			user: parentUser,
			mockExpect: func(m *MockCategoryRepository) {
				m.On("Create", mock.AnythingOfType("*models.Category")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Access denied for child",
			input: CategoryInput{
				Name: "Food",
				Type: "expense",
			},
			user:       childUser,
			mockExpect: func(m *MockCategoryRepository) {},
			wantErr:    true,
			expectedError: ErrAccessDenied,
		},
		{
			name: "Repo error",
			input: CategoryInput{
				Name: "Food",
				Type: "expense",
			},
			user: parentUser,
			mockExpect: func(m *MockCategoryRepository) {
				m.On("Create", mock.AnythingOfType("*models.Category")).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			service := NewCategoryService(mockRepo)
			tt.mockExpect(mockRepo)
			got, err := service.Create(tt.input, tt.user)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.input.Name, got.Name)
				assert.Equal(t, tt.user.FamilyID, got.FamilyID)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCategoryService_GetAll(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	user := &models.User{Base: models.Base{ID: "user1"}, FamilyID: "fam1"}
	categories := []models.Category{
		{Name: "Food"},
		{Name: "Transport"},
	}

	mockRepo.On("GetAll", user.FamilyID).Return(categories, nil)

	got, err := service.GetAll(user)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(got))
	assert.Equal(t, "Food", got[0].Name)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Delete(t *testing.T) {
	parentUser := &models.User{Base: models.Base{ID: "user1"}, FamilyID: "fam1", RoleID: "admin"}
	childUser := &models.User{Base: models.Base{ID: "user2"}, FamilyID: "fam1", RoleID: "child"}

	tests := []struct {
		name       string
		id         string
		user       *models.User
		mockExpect func(m *MockCategoryRepository)
		wantErr    bool
	}{
		{
			name: "Success as admin",
			id:   "cat1",
			user: parentUser,
			mockExpect: func(m *MockCategoryRepository) {
				m.On("GetByID", "cat1", "fam1").Return(&models.Category{Base: models.Base{ID: "cat1"}}, nil)
				m.On("Delete", "cat1", "fam1").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Access denied for child",
			id:   "cat1",
			user: childUser,
			mockExpect: func(m *MockCategoryRepository) {
			},
			wantErr: true,
		},
		{
			name: "Not found",
			id:   "cat2",
			user: parentUser,
			mockExpect: func(m *MockCategoryRepository) {
				m.On("GetByID", "cat2", "fam1").Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			service := NewCategoryService(mockRepo)
			tt.mockExpect(mockRepo)
			err := service.Delete(tt.id, tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
