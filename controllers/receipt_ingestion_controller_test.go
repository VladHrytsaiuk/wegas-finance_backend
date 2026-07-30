package controllers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReceiptIngestionController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockReceiptIngestionService)
	controller := NewReceiptIngestionController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "user-123", "family-456")
		c.Next()
	})

	r.POST("/receipt-ingestion/xml", controller.IngestXML)
	r.POST("/receipt-ingestion/url", controller.IngestURL)

	t.Run("Ingest XML Success", func(t *testing.T) {
		total := int64(20350)
		mockService.On("IngestXML", mock.Anything, mock.Anything).Return(&models.InboxEntry{
			Base:   models.Base{ID: "inbox-xml-1"},
			Status: models.InboxEntryStatusNeedsAccount,
			Total:  &total,
		}, nil).Once()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "receipt.xml")
		_, _ = part.Write([]byte(`<?xml version="1.0"?><RQ></RQ>`))
		_ = writer.Close()

		req, _ := http.NewRequest("POST", "/receipt-ingestion/xml", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Ingest URL Success", func(t *testing.T) {
		total := int64(20350)
		mockService.On("IngestURL", "https://receipt.silpo.elkasa.com.ua/example", mock.Anything).Return(&models.InboxEntry{
			Base:   models.Base{ID: "inbox-url-1"},
			Status: models.InboxEntryStatusNeedsAccount,
			Total:  &total,
		}, nil).Once()

		w := PerformRequest(r, "POST", "/receipt-ingestion/url", ReceiptURLIngestionJSON{
			URL: "https://receipt.silpo.elkasa.com.ua/example",
		})

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp models.InboxEntry
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "inbox-url-1", resp.ID)
	})
}
