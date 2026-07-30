package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type BotAPI interface {
	SendChatMessage(chatID int64, text string, replyMarkup *InlineKeyboardMarkup) error
	AnswerCallbackQuery(callbackQueryID string, text string) error
	GetFile(fileID string) (*BotFile, error)
	DownloadFile(filePath string) ([]byte, error)
	SetWebhook(webhookURL string, secretToken string) error
	GetWebhookInfo() (*WebhookInfo, error)
	DeleteWebhook(dropPendingUpdates bool) error
}

type botClient struct {
	token      string
	httpClient *http.Client
}

type Update struct {
	UpdateID          int64          `json:"update_id"`
	Message           *Message       `json:"message,omitempty"`
	EditedMessage     *Message       `json:"edited_message,omitempty"`
	ChannelPost       *Message       `json:"channel_post,omitempty"`
	EditedChannelPost *Message       `json:"edited_channel_post,omitempty"`
	CallbackQuery     *CallbackQuery `json:"callback_query,omitempty"`
}

type Message struct {
	MessageID int64     `json:"message_id"`
	Text      string    `json:"text,omitempty"`
	Chat      Chat      `json:"chat"`
	From      *User     `json:"from,omitempty"`
	Document  *Document `json:"document,omitempty"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message,omitempty"`
	Data    string   `json:"data"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type BotFile struct {
	FileID   string `json:"file_id"`
	FilePath string `json:"file_path"`
}

type WebhookInfo struct {
	URL                  string `json:"url"`
	HasCustomCertificate bool   `json:"has_custom_certificate"`
	PendingUpdateCount   int    `json:"pending_update_count"`
	LastErrorDate        int64  `json:"last_error_date"`
	LastErrorMessage     string `json:"last_error_message"`
	MaxConnections       int    `json:"max_connections"`
	IPAddress            string `json:"ip_address"`
}

type botResponse[T any] struct {
	OK          bool   `json:"ok"`
	Result      T      `json:"result"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

func NewBotClient(token string) BotAPI {
	return &botClient{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *botClient) SendChatMessage(chatID int64, text string, replyMarkup *InlineKeyboardMarkup) error {
	url := fmt.Sprintf(tgBaseURL, c.token, "sendMessage")

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}

	body, _ := json.Marshal(payload)
	return c.sendJSON(url, body)
}

func (c *botClient) AnswerCallbackQuery(callbackQueryID string, text string) error {
	url := fmt.Sprintf(tgBaseURL, c.token, "answerCallbackQuery")

	body, _ := json.Marshal(map[string]interface{}{
		"callback_query_id": callbackQueryID,
		"text":              text,
	})
	return c.sendJSON(url, body)
}

func (c *botClient) GetFile(fileID string) (*BotFile, error) {
	url := fmt.Sprintf(tgBaseURL, c.token, "getFile")

	body, _ := json.Marshal(map[string]interface{}{
		"file_id": fileID,
	})

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("telegram api error: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var payloadResp botResponse[BotFile]
	if err := json.NewDecoder(resp.Body).Decode(&payloadResp); err != nil {
		return nil, err
	}
	if !payloadResp.OK {
		return nil, fmt.Errorf("telegram api error: %s", payloadResp.Description)
	}
	return &payloadResp.Result, nil
}

func (c *botClient) DownloadFile(filePath string) ([]byte, error) {
	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.token, filePath)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("telegram file api error: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
}

func (c *botClient) SetWebhook(webhookURL string, secretToken string) error {
	url := fmt.Sprintf(tgBaseURL, c.token, "setWebhook")

	payload := map[string]interface{}{
		"url": webhookURL,
	}
	if secretToken != "" {
		payload["secret_token"] = secretToken
	}

	body, _ := json.Marshal(payload)
	return c.sendJSON(url, body)
}

func (c *botClient) GetWebhookInfo() (*WebhookInfo, error) {
	url := fmt.Sprintf(tgBaseURL, c.token, "getWebhookInfo")

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("telegram api error: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var payloadResp botResponse[WebhookInfo]
	if err := json.NewDecoder(resp.Body).Decode(&payloadResp); err != nil {
		return nil, err
	}
	if !payloadResp.OK {
		return nil, fmt.Errorf("telegram api error: %s", payloadResp.Description)
	}

	return &payloadResp.Result, nil
}

func (c *botClient) DeleteWebhook(dropPendingUpdates bool) error {
	url := fmt.Sprintf(tgBaseURL, c.token, "deleteWebhook")

	body, _ := json.Marshal(map[string]interface{}{
		"drop_pending_updates": dropPendingUpdates,
	})
	return c.sendJSON(url, body)
}

func (c *botClient) sendJSON(url string, body []byte) error {
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api error: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var payloadResp botResponse[json.RawMessage]
	if err := json.Unmarshal(bodyBytes, &payloadResp); err != nil {
		return fmt.Errorf("telegram api malformed response: %w, body: %s", err, string(bodyBytes))
	}
	if !payloadResp.OK {
		return fmt.Errorf("telegram api error: %s", payloadResp.Description)
	}

	return nil
}
