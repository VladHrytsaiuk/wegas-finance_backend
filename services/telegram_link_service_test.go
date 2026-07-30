package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegramLinkService_CreateAndCompleteLink(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	nowUser := &models.User{
		Base:     models.Base{ID: uuid.NewString()},
		FamilyID: uuid.NewString(),
		RoleID:   "admin",
		Name:     "Test",
		Email:    "test@example.com",
	}
	require.NoError(t, db.Create(nowUser).Error)

	repo := repositories.NewTelegramLinkRepository(db)
	service := NewTelegramLinkService(repo, db, "wegas_finance_bot")

	tokenResponse, err := service.CreateLinkToken(nowUser)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenResponse.Token)
	assert.Contains(t, tokenResponse.DeepLink, tokenResponse.Token)

	statusBefore, err := service.GetStatus(nowUser)
	require.NoError(t, err)
	assert.False(t, statusBefore.IsLinked)

	statusAfter, err := service.CompleteLink(TelegramLinkCompleteInput{
		Token:          tokenResponse.Token,
		TelegramUserID: 111,
		TelegramChatID: 222,
		Username:       "vlad",
		FirstName:      "Vlad",
	})
	require.NoError(t, err)
	assert.True(t, statusAfter.IsLinked)
	if assert.NotNil(t, statusAfter.TelegramUserID) {
		assert.Equal(t, int64(111), *statusAfter.TelegramUserID)
	}

	statusLoaded, err := service.GetStatus(nowUser)
	require.NoError(t, err)
	assert.True(t, statusLoaded.IsLinked)
	assert.Equal(t, "vlad", statusLoaded.TelegramUsername)
}

func TestTelegramLinkService_RejectsReuseOfUsedToken(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	user := &models.User{
		Base:     models.Base{ID: uuid.NewString()},
		FamilyID: uuid.NewString(),
		RoleID:   "admin",
		Name:     "Test",
		Email:    "test2@example.com",
	}
	require.NoError(t, db.Create(user).Error)

	repo := repositories.NewTelegramLinkRepository(db)
	service := NewTelegramLinkService(repo, db, "wegas_finance_bot")

	tokenResponse, err := service.CreateLinkToken(user)
	require.NoError(t, err)

	_, err = service.CompleteLink(TelegramLinkCompleteInput{
		Token:          tokenResponse.Token,
		TelegramUserID: 111,
		TelegramChatID: 222,
	})
	require.NoError(t, err)

	_, err = service.CompleteLink(TelegramLinkCompleteInput{
		Token:          tokenResponse.Token,
		TelegramUserID: 111,
		TelegramChatID: 222,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired")
}

func TestTelegramLinkService_RevokeLink(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	user := &models.User{
		Base:     models.Base{ID: uuid.NewString()},
		FamilyID: uuid.NewString(),
		RoleID:   "admin",
		Name:     "Test",
		Email:    "test3@example.com",
	}
	require.NoError(t, db.Create(user).Error)

	repo := repositories.NewTelegramLinkRepository(db)
	service := NewTelegramLinkService(repo, db, "wegas_finance_bot")

	tokenResponse, err := service.CreateLinkToken(user)
	require.NoError(t, err)

	_, err = service.CompleteLink(TelegramLinkCompleteInput{
		Token:          tokenResponse.Token,
		TelegramUserID: 111,
		TelegramChatID: 222,
	})
	require.NoError(t, err)

	require.NoError(t, service.RevokeLink(user))

	status, err := service.GetStatus(user)
	require.NoError(t, err)
	assert.False(t, status.IsLinked)
}
