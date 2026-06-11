package services

import (
	"testing"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGoalService_Create(t *testing.T) {
	mockRepo := new(MockGoalRepository)
	mockAccRepo := new(MockAccountRepository)
	mockCurrSvc := new(MockCurrencyService)
	service := NewGoalService(mockRepo, mockAccRepo, mockCurrSvc)

	userID := "user1"
	goal := &models.Goal{
		Name:         "Buy a Car",
		TargetAmount: 10000,
		Currency:     "USD",
	}

	mockRepo.On("Create", mock.AnythingOfType("*models.Goal")).Return(nil)

	err := service.Create(goal, userID)

	assert.NoError(t, err)
	assert.Equal(t, userID, goal.UserID)
	assert.Equal(t, "active", goal.Status)
	assert.Equal(t, "public", goal.Visibility)
	assert.NotEmpty(t, goal.ID)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_CheckOverdueGoals(t *testing.T) {
	mockRepo := new(MockGoalRepository)
	mockAccRepo := new(MockAccountRepository)
	mockCurrSvc := new(MockCurrencyService)
	service := NewGoalService(mockRepo, mockAccRepo, mockCurrSvc)

	now := time.Now().UnixMilli()
	past := now - 100000
	future := now + 100000

	activeGoals := []models.Goal{
		{
			Base:         models.Base{ID: "goal1"},
			Name:         "Overdue Failed",
			DateDeadline: &past,
			TargetAmount: 1000,
			Currency:     "USD",
			Status:       "active",
			Accounts: []models.Account{
				{Balance: 500, Currency: "USD"},
			},
		},
		{
			Base:         models.Base{ID: "goal2"},
			Name:         "Overdue Success",
			DateDeadline: &past,
			TargetAmount: 1000,
			Currency:     "USD",
			Status:       "active",
			Accounts: []models.Account{
				{Balance: 1500, Currency: "USD"},
			},
		},
		{
			Base:         models.Base{ID: "goal3"},
			Name:         "Not Overdue",
			DateDeadline: &future,
			TargetAmount: 1000,
			Currency:     "USD",
			Status:       "active",
			Accounts: []models.Account{
				{Balance: 100, Currency: "USD"},
			},
		},
	}

	mockRepo.On("FindAllActive").Return(activeGoals, nil)
	// For goal1: 500 < 1000 -> failed
	mockRepo.On("Update", mock.MatchedBy(func(g *models.Goal) bool {
		return g.ID == "goal1" && g.Status == "failed"
	})).Return(nil)

	err := service.CheckOverdueGoals()

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_GetAll_StatusUpdate(t *testing.T) {
	mockRepo := new(MockGoalRepository)
	mockAccRepo := new(MockAccountRepository)
	mockCurrSvc := new(MockCurrencyService)
	service := NewGoalService(mockRepo, mockAccRepo, mockCurrSvc)

	familyID := "fam1"
	userID := "user1"
	past := time.Now().UnixMilli() - 100000

	goals := []models.Goal{
		{
			Base:         models.Base{ID: "goal1"},
			DateDeadline: &past,
			TargetAmount: 1000,
			Currency:     "USD",
			Status:       "active",
			Accounts: []models.Account{
				{Balance: 200, Currency: "USD"},
			},
		},
	}

	mockRepo.On("FindAllByFamily", familyID, userID).Return(goals, nil)
	// The service will update the goal status to failed in a goroutine
	// but we can't easily wait for it unless we change the service.
	// Actually, the service uses `go s.repo.Update(&goals[i])`.
	// For testing, we might want to wait a bit or mock it.
	mockRepo.On("Update", mock.AnythingOfType("*models.Goal")).Return(nil).Maybe()

	result, err := service.GetAll(familyID, userID)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "failed", result[0].Status)
}
