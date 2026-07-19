package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/pkg/telegram"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTelegramBotController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockTelegramBotService)
	controller := NewTelegramBotController(mockService, "secret-token")

	r := gin.Default()
	r.POST("/telegram/webhook", controller.Webhook)

	t.Run("Webhook Success", func(t *testing.T) {
		mockService.On("HandleUpdate", mock.MatchedBy(func(update *telegram.Update) bool {
			return update != nil && update.Message != nil && update.Message.Text == "/start abc"
		})).Return(nil).Once()

		body, _ := json.Marshal(telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				Text: "/start abc",
				Chat: telegram.Chat{ID: 10},
			},
		})

		req, _ := http.NewRequest("POST", "/telegram/webhook", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "secret-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Webhook Unauthorized", func(t *testing.T) {
		body, _ := json.Marshal(telegram.Update{})
		req, _ := http.NewRequest("POST", "/telegram/webhook", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Webhook Service Error", func(t *testing.T) {
		mockService.On("HandleUpdate", mock.Anything).Return(errors.New("boom")).Once()

		body, _ := json.Marshal(telegram.Update{
			UpdateID: 2,
			Message: &telegram.Message{
				Text: "https://example.com/receipt",
				Chat: telegram.Chat{ID: 10},
			},
		})

		req, _ := http.NewRequest("POST", "/telegram/webhook", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "secret-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
