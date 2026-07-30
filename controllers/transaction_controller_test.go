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

func TestTransactionController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockTransactionService)
	controller := NewTransactionController(mockService)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "user-123", "family-456")
		c.Next()
	})
	
	r.POST("/transactions", controller.Create)
	r.GET("/transactions", controller.GetAll)
	r.GET("/transactions/:id", controller.GetOne)
	r.PUT("/transactions/:id", controller.Update)
	r.DELETE("/transactions/:id", controller.Delete)
	r.POST("/transactions/batch", controller.BatchCreate)
	r.POST("/transactions/:id/receipt", controller.UploadReceipt)
	r.DELETE("/transactions/:id/receipt", controller.DeleteReceipt)
	r.DELETE("/transactions/photos/:id", controller.DeletePhoto)
	r.GET("/transactions/predict", controller.PredictCategory)
	r.GET("/transactions/:id/receipt-sources", controller.GetLinkedReceipts)
	r.POST("/transactions/:id/receipt-sources/unlink", controller.UnlinkReceiptSource)

	t.Run("Create Success JSON", func(t *testing.T) {
		inputJSON := CreateTxJSON{
			AccountID: "acc-1",
			Amount:    1000,
			Date:      123456789,
			Type:      "expense",
		}

		mockService.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("tx-1", nil).Once()

		w := PerformRequest(r, "POST", "/transactions", inputJSON)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("GetAll", func(t *testing.T) {
		mockService.On("GetAll", mock.Anything, mock.Anything).Return([]models.Transaction{
			{Base: models.Base{ID: "tx-1"}, Amount: 100},
		}, int64(1), nil).Once()

		w := PerformRequest(r, "GET", "/transactions", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["count"])
	})

	t.Run("GetOne Success", func(t *testing.T) {
		mockService.On("GetByID", "tx-1", mock.Anything).Return(&models.Transaction{
			Base: models.Base{ID: "tx-1"},
		}, nil).Once()

		w := PerformRequest(r, "GET", "/transactions/tx-1", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Update Success", func(t *testing.T) {
		input := CreateTxJSON{
			AccountID: "acc-1",
			Amount:    2000,
			Date:      123456789,
			Type:      "expense",
		}
		mockService.On("Update", "tx-1", mock.Anything, mock.Anything).Return(nil).Once()

		w := PerformRequest(r, "PUT", "/transactions/tx-1", input)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Delete Success", func(t *testing.T) {
		mockService.On("Delete", "tx-1", mock.Anything).Return(nil).Once()

		w := PerformRequest(r, "DELETE", "/transactions/tx-1", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("BatchCreate Success", func(t *testing.T) {
		inputs := []CreateTxJSON{
			{AccountID: "acc-1", Amount: 100, Type: "expense", Date: 123},
			{AccountID: "acc-1", Amount: 200, Type: "income", Date: 124},
		}
		mockService.On("BatchCreate", mock.Anything, mock.Anything).Return(2, nil).Once()

		w := PerformRequest(r, "POST", "/transactions/batch", inputs)

		assert.Equal(t, http.StatusCreated, w.Code)
		var res map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &res)
		assert.Equal(t, float64(2), res["count"])
	})

	t.Run("Delete Receipt Success", func(t *testing.T) {
		mockService.On("DeleteReceipt", "tx-1", mock.Anything).Return(nil).Once()
		w := PerformRequest(r, "DELETE", "/transactions/tx-1/receipt", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Delete Photo Success", func(t *testing.T) {
		mockService.On("DeletePhoto", "photo-1", mock.Anything).Return(nil).Once()
		w := PerformRequest(r, "DELETE", "/transactions/photos/photo-1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Predict Category Success", func(t *testing.T) {
		mockService.On("PredictCategory", "Milk", mock.Anything).Return("cat-1", nil).Once()
		w := PerformRequest(r, "GET", "/transactions/predict?name=Milk", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		var res map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &res)
		assert.Equal(t, "cat-1", res["category_id"])
	})

	t.Run("Upload Receipt Success", func(t *testing.T) {
		mockService.On("UploadReceipt", "tx-1", mock.Anything, mock.Anything, mock.Anything).Return("/path/to/receipt.jpg", nil).Once()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.jpg")
		part.Write([]byte("fake image data"))
		writer.Close()

		req, _ := http.NewRequest("POST", "/transactions/tx-1/receipt", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var res map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &res)
		assert.Equal(t, "/path/to/receipt.jpg", res["path"])
	})

	t.Run("Get Linked Receipts Success", func(t *testing.T) {
		mockService.On("GetLinkedReceipts", "tx-1", mock.Anything).Return([]models.ReceiptSource{
			{Base: models.Base{ID: "rs-1"}, Merchant: "Silpo"},
		}, nil).Once()

		w := PerformRequest(r, "GET", "/transactions/tx-1/receipt-sources", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Unlink Receipt Source Success", func(t *testing.T) {
		mockService.On("UnlinkReceiptSource", "tx-1", "rs-1", mock.Anything).Return(nil).Once()

		w := PerformRequest(r, "POST", "/transactions/tx-1/receipt-sources/unlink", UnlinkReceiptSourceJSON{
			ReceiptSourceID: "rs-1",
		})
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
