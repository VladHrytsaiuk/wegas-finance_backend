package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/pkg/telegram"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTelegramBotService_StartLink(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	user := createTelegramTestUser(t, db, "start@example.com")
	linkRepo := repositories.NewTelegramLinkRepository(db)
	linkService := NewTelegramLinkService(linkRepo, db, "wegas_finance_bot")
	token, err := linkService.CreateLinkToken(user)
	require.NoError(t, err)

	bot := new(mockBotAPI)
	bot.On("SendChatMessage", int64(555), mock.AnythingOfType("string"), (*telegram.InlineKeyboardMarkup)(nil)).Return(nil).Once()

	service := NewTelegramBotService(bot, linkRepo, repositories.NewUserRepository(db), new(MockAccountService), new(MockInboxService), new(MockReceiptIngestionService))
	err = service.HandleUpdate(&telegram.Update{
		Message: &telegram.Message{
			Text: "/start " + token.Token,
			Chat: telegram.Chat{ID: 555},
			From: &telegram.User{ID: 12345, Username: "vlad", FirstName: "Vlad"},
		},
	})
	require.NoError(t, err)
	bot.AssertExpectations(t)

	status, err := linkService.GetStatus(user)
	require.NoError(t, err)
	assert.True(t, status.IsLinked)
}

func TestTelegramBotService_URLFlowSendsSuccessMessage(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	user := createTelegramTestUser(t, db, "url@example.com")
	linkRepo := repositories.NewTelegramLinkRepository(db)
	require.NoError(t, linkTelegramUser(t, db, linkRepo, user, 12345, 555))

	bot := new(mockBotAPI)
	receipt := new(MockReceiptIngestionService)
	inboxService := new(MockInboxService)

	total := int64(20350)
	entry := &models.InboxEntry{
		Base:     models.Base{ID: "inbox-1"},
		Status:   models.InboxEntryStatusNeedsAccount,
		Total:    &total,
		Currency: "UAH",
	}
	receipt.On("IngestURL", "https://receipt.silpo.elkasa.com.ua/demo", mock.Anything).Return(entry, nil).Once()
	bot.On("SendChatMessage", int64(555), mock.AnythingOfType("string"), (*telegram.InlineKeyboardMarkup)(nil)).Return(nil).Once()

	service := NewTelegramBotService(bot, linkRepo, repositories.NewUserRepository(db), new(MockAccountService), inboxService, receipt)
	err = service.HandleUpdate(&telegram.Update{
		Message: &telegram.Message{
			Text: "https://receipt.silpo.elkasa.com.ua/demo",
			Chat: telegram.Chat{ID: 555},
			From: &telegram.User{ID: 12345},
		},
	})
	require.NoError(t, err)

	bot.AssertExpectations(t)
	receipt.AssertExpectations(t)
}

func TestTelegramBotService_ChannelPostDocumentFlow(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	user := createTelegramTestUser(t, db, "channel@example.com")
	linkRepo := repositories.NewTelegramLinkRepository(db)
	require.NoError(t, linkTelegramUser(t, db, linkRepo, user, 12345, 555))

	bot := new(mockBotAPI)
	receipt := new(MockReceiptIngestionService)
	inboxService := new(MockInboxService)

	rawXML := []byte(`<?xml version="1.0" encoding="utf-8"?><RQ V="1"></RQ>`)
	total := int64(17880)
	entry := &models.InboxEntry{
		Base:     models.Base{ID: "inbox-channel"},
		Status:   models.InboxEntryStatusNeedsAccount,
		Total:    &total,
		Currency: "UAH",
	}

	bot.On("GetFile", "file-1").Return(&telegram.BotFile{
		FileID:   "file-1",
		FilePath: "documents/test.xml",
	}, nil).Once()
	bot.On("DownloadFile", "documents/test.xml").Return(rawXML, nil).Once()
	receipt.On("IngestXMLBytes", rawXML, mock.Anything).Return(entry, nil).Once()
	bot.On("SendChatMessage", int64(555), mock.AnythingOfType("string"), (*telegram.InlineKeyboardMarkup)(nil)).Return(nil).Once()

	service := NewTelegramBotService(bot, linkRepo, repositories.NewUserRepository(db), new(MockAccountService), inboxService, receipt)
	err = service.HandleUpdate(&telegram.Update{
		ChannelPost: &telegram.Message{
			Chat: telegram.Chat{ID: 555},
			From: &telegram.User{ID: 12345},
			Document: &telegram.Document{
				FileID:   "file-1",
				FileName: "W9bEBMTh9_4.xml",
				MimeType: "application/octet-stream",
			},
		},
	})
	require.NoError(t, err)

	bot.AssertExpectations(t)
	receipt.AssertExpectations(t)
}

func TestTelegramBotService_ChannelPostWithoutFromIsIgnored(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	linkRepo := repositories.NewTelegramLinkRepository(db)
	bot := new(mockBotAPI)
	receipt := new(MockReceiptIngestionService)
	accountService := new(MockAccountService)
	inboxService := new(MockInboxService)

	service := NewTelegramBotService(bot, linkRepo, repositories.NewUserRepository(db), accountService, inboxService, receipt)
	err = service.HandleUpdate(&telegram.Update{
		ChannelPost: &telegram.Message{
			Chat: telegram.Chat{ID: -1001234567890},
			Document: &telegram.Document{
				FileID:   "file-1",
				FileName: "W9bEBMTh9_4.xml",
				MimeType: "application/octet-stream",
			},
		},
	})
	require.NoError(t, err)

	bot.AssertNotCalled(t, "GetFile", mock.Anything)
	bot.AssertNotCalled(t, "SendChatMessage", mock.Anything, mock.Anything, mock.Anything)
	receipt.AssertNotCalled(t, "IngestXMLBytes", mock.Anything, mock.Anything)
	accountService.AssertNotCalled(t, "GetAll", mock.Anything)
	inboxService.AssertNotCalled(t, "SelectAccount", mock.Anything, mock.Anything, mock.Anything)
}

type mockBotAPI struct {
	mock.Mock
}

func (m *mockBotAPI) SendChatMessage(chatID int64, text string, replyMarkup *telegram.InlineKeyboardMarkup) error {
	args := m.Called(chatID, text, replyMarkup)
	return args.Error(0)
}

func (m *mockBotAPI) AnswerCallbackQuery(callbackQueryID string, text string) error {
	args := m.Called(callbackQueryID, text)
	return args.Error(0)
}

func (m *mockBotAPI) GetFile(fileID string) (*telegram.BotFile, error) {
	args := m.Called(fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*telegram.BotFile), args.Error(1)
}

func (m *mockBotAPI) DownloadFile(filePath string) ([]byte, error) {
	args := m.Called(filePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockBotAPI) SetWebhook(webhookURL string, secretToken string) error {
	args := m.Called(webhookURL, secretToken)
	return args.Error(0)
}

func (m *mockBotAPI) GetWebhookInfo() (*telegram.WebhookInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*telegram.WebhookInfo), args.Error(1)
}

func (m *mockBotAPI) DeleteWebhook(dropPendingUpdates bool) error {
	args := m.Called(dropPendingUpdates)
	return args.Error(0)
}

func (m *mockBotAPI) GetUpdates(offset int, timeout int) ([]telegram.Update, error) {
	args := m.Called(offset, timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]telegram.Update), args.Error(1)
}

func createTelegramTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
	t.Helper()

	user := &models.User{
		Base:     models.Base{ID: uuid.NewString()},
		FamilyID: uuid.NewString(),
		RoleID:   "admin",
		Name:     "Test",
		Email:    email,
	}
	require.NoError(t, db.Create(user).Error)
	return user
}

func linkTelegramUser(t *testing.T, db *gorm.DB, repo repositories.TelegramLinkRepository, user *models.User, telegramUserID int64, chatID int64) error {
	t.Helper()
	service := NewTelegramLinkService(repo, db, "wegas_finance_bot")
	token, err := service.CreateLinkToken(user)
	if err != nil {
		return err
	}
	_, err = service.CompleteLink(TelegramLinkCompleteInput{
		Token:          token.Token,
		TelegramUserID: telegramUserID,
		TelegramChatID: chatID,
		Username:       "user",
		FirstName:      "First",
	})
	return err
}
