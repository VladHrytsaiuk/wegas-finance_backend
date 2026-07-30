package services

import (
	"errors"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/pkg/telegram"
)

type TelegramWebhookStatus struct {
	Configured       bool   `json:"configured"`
	WebhookURL       string `json:"webhook_url,omitempty"`
	ExpectedWebhookURL string `json:"expected_webhook_url,omitempty"`
	PendingUpdateCount int   `json:"pending_update_count"`
	LastErrorDate    int64  `json:"last_error_date,omitempty"`
	LastErrorMessage string `json:"last_error_message,omitempty"`
	IPAddress        string `json:"ip_address,omitempty"`
}

type TelegramWebhookService interface {
	GetWebhookStatus() (*TelegramWebhookStatus, error)
	SyncWebhook() (*TelegramWebhookStatus, error)
	DeleteWebhook(dropPendingUpdates bool) error
}

type telegramWebhookService struct {
	botAPI        telegram.BotAPI
	appURL        string
	webhookSecret string
}

func NewTelegramWebhookService(botAPI telegram.BotAPI, appURL string, webhookSecret string) TelegramWebhookService {
	return &telegramWebhookService{
		botAPI:        botAPI,
		appURL:        strings.TrimRight(strings.TrimSpace(appURL), "/"),
		webhookSecret: webhookSecret,
	}
}

func (s *telegramWebhookService) GetWebhookStatus() (*TelegramWebhookStatus, error) {
	info, err := s.botAPI.GetWebhookInfo()
	if err != nil {
		return nil, err
	}

	return &TelegramWebhookStatus{
		Configured:         strings.TrimSpace(info.URL) != "",
		WebhookURL:         info.URL,
		ExpectedWebhookURL: s.expectedWebhookURL(),
		PendingUpdateCount: info.PendingUpdateCount,
		LastErrorDate:      info.LastErrorDate,
		LastErrorMessage:   info.LastErrorMessage,
		IPAddress:          info.IPAddress,
	}, nil
}

func (s *telegramWebhookService) SyncWebhook() (*TelegramWebhookStatus, error) {
	if strings.TrimSpace(s.appURL) == "" {
		return nil, errors.New("app url is not configured")
	}

	if err := s.botAPI.SetWebhook(s.expectedWebhookURL(), s.webhookSecret); err != nil {
		return nil, err
	}

	return s.GetWebhookStatus()
}

func (s *telegramWebhookService) DeleteWebhook(dropPendingUpdates bool) error {
	return s.botAPI.DeleteWebhook(dropPendingUpdates)
}

func (s *telegramWebhookService) expectedWebhookURL() string {
	if s.appURL == "" {
		return ""
	}
	return s.appURL + "/api/telegram/webhook"
}
