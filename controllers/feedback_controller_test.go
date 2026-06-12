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

func TestFeedbackController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockFeedbackService)
	controller := NewFeedbackController(mockService)

	r := gin.Default()
	r.POST("/feedback", controller.Submit)

	t.Run("Submit Success", func(t *testing.T) {
		mockService.On("SendFeedback", "John", "john@test.com", "Hello", "high", mock.Anything).Return(nil).Once()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("name", "John")
		writer.WriteField("contact", "john@test.com")
		writer.WriteField("message", "Hello")
		writer.WriteField("priority", "high")
		
		part, _ := writer.CreateFormFile("images", "test.jpg")
		part.Write([]byte("fake image data"))
		writer.Close()

		req, _ := http.NewRequest("POST", "/feedback", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Submit Fail - No Message", func(t *testing.T) {
		w := PerformRequest(r, "POST", "/feedback", nil) // PerformRequest sends JSON, message will be empty
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
