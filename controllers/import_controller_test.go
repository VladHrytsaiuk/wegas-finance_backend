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

func TestImportController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockImportService)
	controller := NewImportController(mockService)

	r := gin.Default()
	r.POST("/import/upload", controller.UploadStatement)

	t.Run("Upload Success", func(t *testing.T) {
		mockService.On("ProcessFile", mock.Anything, "acc-1", "monobank").Return(&services.PreviewResult{
			Transactions: []services.PreviewTransaction{{Amount: 100}},
		}, nil).Once()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("account_id", "acc-1")
		writer.WriteField("bank_type", "monobank")
		part, _ := writer.CreateFormFile("file", "statement.pdf")
		part.Write([]byte("%PDF-1.4...%%EOF"))
		writer.Close()

		req, _ := http.NewRequest("POST", "/import/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
