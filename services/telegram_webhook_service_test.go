package services

import (
	"errors"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/pkg/telegram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegramWebhookService_GetStatus(t *testing.T) {
	bot := new(mockWebhookBotAPI)
	bot.On("GetWebhookInfo").Return(&telegram.WebhookInfo{
		URL:                "https://app.example.com/api/telegram/webhook",
		PendingUpdateCount: 3,
		LastErrorMessage:   "",
	}, nil).Once()

	service := NewTelegramWebhookService(bot, "https://app.example.com", "secret")
	status, err := service.GetWebhookStatus()

	require.NoError(t, err)
	assert.True(t, status.Configured)
	assert.Equal(t, "https://app.example.com/api/telegram/webhook", status.ExpectedWebhookURL)
	assert.Equal(t, 3, status.PendingUpdateCount)
}

func TestTelegramWebhookService_SyncWebhook(t *testing.T) {
	bot := new(mockWebhookBotAPI)
	bot.On("SetWebhook", "https://app.example.com/api/telegram/webhook", "secret").Return(nil).Once()
	bot.On("GetWebhookInfo").Return(&telegram.WebhookInfo{
		URL:                "https://app.example.com/api/telegram/webhook",
		PendingUpdateCount: 0,
	}, nil).Once()

	service := NewTelegramWebhookService(bot, "https://app.example.com", "secret")
	status, err := service.SyncWebhook()

	require.NoError(t, err)
	assert.True(t, status.Configured)
	bot.AssertExpectations(t)
}

func TestTelegramWebhookService_DeleteWebhook(t *testing.T) {
	bot := new(mockWebhookBotAPI)
	bot.On("DeleteWebhook", true).Return(nil).Once()

	service := NewTelegramWebhookService(bot, "https://app.example.com", "secret")
	require.NoError(t, service.DeleteWebhook(true))
	bot.AssertExpectations(t)
}

func TestTelegramWebhookService_SyncWebhookRequiresAppURL(t *testing.T) {
	service := NewTelegramWebhookService(new(mockWebhookBotAPI), "", "secret")
	_, err := service.SyncWebhook()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "app url")
}

type mockWebhookBotAPI struct {
	mockBotAPI
}

func (m *mockWebhookBotAPI) SetWebhook(webhookURL string, secretToken string) error {
	args := m.Called(webhookURL, secretToken)
	return args.Error(0)
}

func (m *mockWebhookBotAPI) GetWebhookInfo() (*telegram.WebhookInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*telegram.WebhookInfo), args.Error(1)
}

func (m *mockWebhookBotAPI) DeleteWebhook(dropPendingUpdates bool) error {
	args := m.Called(dropPendingUpdates)
	return args.Error(0)
}

func (m *mockWebhookBotAPI) SendChatMessage(chatID int64, text string, replyMarkup *telegram.InlineKeyboardMarkup) error {
	return errors.New("not used")
}

func (m *mockWebhookBotAPI) AnswerCallbackQuery(callbackQueryID string, text string) error {
	return errors.New("not used")
}

func (m *mockWebhookBotAPI) GetFile(fileID string) (*telegram.BotFile, error) {
	return nil, errors.New("not used")
}

func (m *mockWebhookBotAPI) DownloadFile(filePath string) ([]byte, error) {
	return nil, errors.New("not used")
}

func (m *mockWebhookBotAPI) GetUpdates(offset int, timeout int) ([]telegram.Update, error) {
	return nil, errors.New("not used")
}
