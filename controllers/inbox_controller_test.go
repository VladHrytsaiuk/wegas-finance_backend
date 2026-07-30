package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInboxController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(services.MockInboxService)
	mockStorage := new(services.MockStorageService)
	controller := NewInboxController(mockService, mockStorage)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		SetupTestUser(c, "user-123", "family-456")
		c.Next()
	})

	r.POST("/inbox", controller.Create)
	r.POST("/inbox/photo", controller.CreatePhoto)
	r.GET("/inbox", controller.GetAll)
	r.GET("/inbox/:id/account-candidates", controller.GetAccountCandidates)
	r.GET("/inbox/:id/transaction-candidates", controller.GetTransactionCandidates)
	r.GET("/inbox/:id", controller.GetOne)
	r.PATCH("/inbox/:id/account", controller.SelectAccount)
	r.POST("/inbox/:id/link", controller.Link)
	r.POST("/inbox/:id/unlink", controller.Unlink)

	t.Run("Create Success", func(t *testing.T) {
		total := int64(10000)
		mockService.On("Create", mock.Anything, mock.Anything).Return(&models.InboxEntry{
			Base:   models.Base{ID: "inbox-1"},
			Status: models.InboxEntryStatusNeedsReview,
			Total:  &total,
		}, nil).Once()

		w := PerformRequest(r, "POST", "/inbox", InboxCreateJSON{
			SourceType: "xml",
			Origin:     "telegram_xml",
			Merchant:   "Silpo",
			Total:      &total,
		})

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("returns transaction candidates", func(t *testing.T) {
		mockService.On("FindTransactionCandidates", "inbox-1", mock.Anything).Return([]services.InboxTransactionCandidate{{
			TransactionID: "tx-1",
			Score:         85,
		}}, nil).Once()

		w := PerformRequest(r, "GET", "/inbox/inbox-1/transaction-candidates", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetAll Success", func(t *testing.T) {
		total := int64(5000)
		mockService.On("GetAll", mock.Anything, mock.Anything).Return([]models.InboxEntry{
			{Base: models.Base{ID: "inbox-1"}, Status: models.InboxEntryStatusNeedsLink, Total: &total},
		}, int64(1), nil).Once()

		w := PerformRequest(r, "GET", "/inbox", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["total"])
	})

	t.Run("GetOne Success", func(t *testing.T) {
		mockService.On("GetByID", "inbox-1", mock.Anything).Return(&models.InboxEntry{
			Base:   models.Base{ID: "inbox-1"},
			Status: models.InboxEntryStatusNeedsReview,
		}, nil).Once()

		w := PerformRequest(r, "GET", "/inbox/inbox-1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetAccountCandidates Success", func(t *testing.T) {
		mockService.On("FindAccountCandidates", "inbox-1", mock.Anything).Return([]services.InboxAccountCandidate{{
			AccountID:     "acc-1",
			MatchedDigits: 4,
			Confidence:    "exact",
			Recommended:   true,
		}}, nil).Once()

		w := PerformRequest(r, "GET", "/inbox/inbox-1/account-candidates", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("SelectAccount Success", func(t *testing.T) {
		accountID := "acc-1"
		mockService.On("SelectAccount", "inbox-1", accountID, mock.Anything).Return(&models.InboxEntry{
			Base:              models.Base{ID: "inbox-1"},
			Status:            models.InboxEntryStatusNeedsLink,
			SelectedAccountID: &accountID,
		}, nil).Once()

		w := PerformRequest(r, "PATCH", "/inbox/inbox-1/account", InboxSelectAccountJSON{
			AccountID: accountID,
		})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Link Success", func(t *testing.T) {
		txID := "tx-1"
		mockService.On("Link", "inbox-1", txID, true, false, mock.Anything).Return(&models.InboxEntry{
			Base:                 models.Base{ID: "inbox-1"},
			Status:               models.InboxEntryStatusLinked,
			MatchedTransactionID: &txID,
		}, nil).Once()

		w := PerformRequest(r, "POST", "/inbox/inbox-1/link", InboxLinkJSON{
			TransactionID: txID,
		})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Unlink Success", func(t *testing.T) {
		mockService.On("Unlink", "inbox-1", mock.Anything).Return(&models.InboxEntry{
			Base:   models.Base{ID: "inbox-1"},
			Status: models.InboxEntryStatusUnlinked,
			Reason: "unlinked_by_user",
		}, nil).Once()

		w := PerformRequest(r, "POST", "/inbox/inbox-1/unlink", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
