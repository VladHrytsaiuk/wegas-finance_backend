package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCounterpartyService_Create(t *testing.T) {
	mockRepo := new(MockCounterpartyRepository)
	service := NewCounterpartyService(mockRepo)

	parentUser := &models.User{Base: models.Base{ID: "user1"}, FamilyID: "fam1", RoleID: "admin"}
	childUser := &models.User{Base: models.Base{ID: "user2"}, FamilyID: "fam1", RoleID: "child"}

	tests := []struct {
		name       string
		input      CounterpartyInput
		user       *models.User
		mockExpect func()
		wantErr    bool
	}{
		{
			name: "Success as admin",
			input: CounterpartyInput{
				Name: "Supermarket",
				Type: "merchant",
			},
			user: parentUser,
			mockExpect: func() {
				mockRepo.On("Create", mock.AnythingOfType("*models.Counterparty")).Return(nil)
				mockRepo.On("GetByID", mock.Anything, "fam1").Return(&models.Counterparty{Name: "Supermarket"}, nil)
			},
			wantErr: false,
		},
		{
			name: "Access denied for child",
			input: CounterpartyInput{
				Name: "Supermarket",
			},
			user:       childUser,
			mockExpect: func() {},
			wantErr:    true,
		},
		{
			name: "Logo too large",
			input: CounterpartyInput{
				Name: "BigLogo",
				Logo: string(make([]byte, MaxLogoLength+1)),
			},
			user:       parentUser,
			mockExpect: func() {},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			got, err := service.Create(tt.input, tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.input.Name, got.Name)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCounterpartyService_GetAll(t *testing.T) {
	mockRepo := new(MockCounterpartyRepository)
	service := NewCounterpartyService(mockRepo)

	user := &models.User{Base: models.Base{ID: "user1"}, FamilyID: "fam1"}
	counterparties := []models.Counterparty{
		{Name: "CP1"},
		{Name: "CP2"},
	}

	mockRepo.On("GetAll", user.FamilyID).Return(counterparties, nil)

	got, err := service.GetAll(user)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(got))
	mockRepo.AssertExpectations(t)
}

func TestCounterpartyService_CreateCategory(t *testing.T) {
	mockRepo := new(MockCounterpartyRepository)
	service := NewCounterpartyService(mockRepo)

	user := &models.User{Base: models.Base{ID: "user1"}, FamilyID: "fam1", RoleID: "admin"}
	input := CpCategoryInput{Name: "Retail"}

	mockRepo.On("CreateCategory", mock.AnythingOfType("*models.CounterpartyCategory")).Return(nil)

	got, err := service.CreateCategory(input, user)

	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "Retail", got.Name)
	mockRepo.AssertExpectations(t)
}
