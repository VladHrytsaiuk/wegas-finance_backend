package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

var tgBaseURL = "https://api.telegram.org/bot%s/%s"

type Client struct {
	token  string
	chatID string
	client *http.Client
}

func NewClient(token string, chatID string) *Client {
	return &Client{
		token:  token,
		chatID: chatID,
		client: &http.Client{Timeout: 30 * time.Second}, // Збільшив таймаут, бо фото можуть важити багато
	}
}

// SendMessage: Тільки текст
func (c *Client) SendMessage(text string) error {
	url := fmt.Sprintf(tgBaseURL, c.token, "sendMessage")

	body, _ := json.Marshal(map[string]string{
		"chat_id":    c.chatID,
		"text":       text,
		"parse_mode": "HTML",
	})

	return c.sendJSON(url, body)
}

// SendPhoto: Одне фото + підпис
func (c *Client) SendPhoto(caption string, photoName string, photoBytes []byte) error {
	url := fmt.Sprintf(tgBaseURL, c.token, "sendPhoto")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("chat_id", c.chatID)
	_ = writer.WriteField("caption", caption)
	_ = writer.WriteField("parse_mode", "HTML")

	part, err := writer.CreateFormFile("photo", photoName)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, bytes.NewReader(photoBytes)); err != nil {
		return err
	}

	writer.Close()
	return c.sendMultipart(url, writer.FormDataContentType(), body)
}

// SendMediaGroup: Декілька фото (альбом) + підпис
func (c *Client) SendMediaGroup(caption string, photos [][]byte) error {
	url := fmt.Sprintf(tgBaseURL, c.token, "sendMediaGroup")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("chat_id", c.chatID)

	// Структура для JSON об'єкта media
	type mediaEntry struct {
		Type      string `json:"type"`
		Media     string `json:"media"`
		Caption   string `json:"caption,omitempty"`
		ParseMode string `json:"parse_mode,omitempty"`
	}

	var mediaList []mediaEntry

	for i, photoBytes := range photos {
		// Унікальне ім'я поля для кожного файлу (photo_0, photo_1...)
		fieldName := fmt.Sprintf("photo_%d", i)
		fileName := fmt.Sprintf("image_%d.jpg", i)

		entry := mediaEntry{
			Type:  "photo",
			Media: "attach://" + fieldName, // Посилання на multipart field
		}

		// Підпис додаємо тільки до першого фото (так працює Telegram альбом)
		if i == 0 {
			entry.Caption = caption
			entry.ParseMode = "HTML"
		}

		mediaList = append(mediaList, entry)

		// Додаємо сам файл у multipart
		part, err := writer.CreateFormFile(fieldName, fileName)
		if err != nil {
			return err
		}
		if _, err := io.Copy(part, bytes.NewReader(photoBytes)); err != nil {
			return err
		}
	}

	// Серіалізуємо список медіа в JSON і додаємо як поле "media"
	jsonMedia, _ := json.Marshal(mediaList)
	_ = writer.WriteField("media", string(jsonMedia))

	writer.Close()
	return c.sendMultipart(url, writer.FormDataContentType(), body)
}

// Допоміжні приватні методи для зменшення дублювання коду
func (c *Client) sendJSON(url string, body []byte) error {
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram api error: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

func (c *Client) sendMultipart(url, contentType string, body io.Reader) error {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram api error: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}