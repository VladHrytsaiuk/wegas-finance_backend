package controllers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestImportController_Limits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockImportService)
	controller := NewImportController(mockService)

	r := gin.New()
	// Set the same limit as in main.go
	r.MaxMultipartMemory = 8 << 20 // 8MB for test
	r.POST("/import/upload", controller.UploadStatement)

	t.Run("Upload Large File - Within Limit", func(t *testing.T) {
		// 1MB file
		largeContent := make([]byte, 1*1024*1024)
		for i := range largeContent {
			largeContent[i] = 'A'
		}

		mockService.On("ProcessFile", mock.Anything, "acc-1", "monobank").
			Return(&services.PreviewResult{}, nil).Once()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("account_id", "acc-1")
		writer.WriteField("bank_type", "monobank")
		part, _ := writer.CreateFormFile("file", "large.pdf")
		part.Write(largeContent)
		writer.Close()

		req, _ := http.NewRequest("POST", "/import/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Upload Very Large File - Exceeding 20MB Hard Limit", func(t *testing.T) {
		// 21MB file
		hugeContent := make([]byte, 21*1024*1024)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("account_id", "acc-1")
		writer.WriteField("bank_type", "monobank")
		part, _ := writer.CreateFormFile("file", "too-big.pdf")
		part.Write(hugeContent)
		writer.Close()

		req, _ := http.NewRequest("POST", "/import/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "exceeds limit")
	})
}
