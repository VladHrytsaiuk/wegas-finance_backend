package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestMonobankService_Integration(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(
		&models.BankAccountMapping{},
		&models.BankConnection{},
		&models.User{},
		&models.Transaction{},
		&models.Account{},
	)

	mockTxService := new(MockTransactionService)
	mockAccRepo := new(MockAccountRepository)
	mockInboxService := new(MockInboxService)
	mockClock := utils.NewMockClock(time.Now())

	svc := NewMonobankService(db, mockTxService, mockAccRepo, mockClock, mockInboxService).(*monobankService)
	svc.SkipRateLimit = true

	t.Run("Connect - success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("MOCK RECEIVED: %s %s\n", r.Method, r.URL.Path)
			receivedToken := r.Header.Get("X-Token")
			if receivedToken != "test-token" {
				fmt.Printf("MOCK: WRONG TOKEN: '%s'\n", receivedToken)
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"clientId":"c1","name":"N","accounts":[{"id":"a1","currencyCode":980,"balance":100}]}`)
		}))
		defer server.Close()

		svc.baseURL = server.URL

		accounts, err := svc.Connect("u1", "f1", "test-token")
		assert.NoError(t, err)
		if assert.Len(t, accounts, 1) {
			assert.Equal(t, "a1", accounts[0].ID)
		}

		// Verify connection saved in DB
		var conn models.BankConnection
		err = db.First(&conn, "user_id = ?", "u1").Error
		assert.NoError(t, err)
		assert.Equal(t, "monobank", conn.Provider)
	})

	t.Run("Sync - success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/personal/client-info" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"accounts":[{"id":"a1","balance":150}]}`)
				return
			}
			// Statement path
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `[{"id":"tx1","time":1710000000,"description":"D","mcc":5411,"amount":-50}]`)
		}))
		defer server.Close()

		svc.baseURL = server.URL
		// Set clock to Feb 1, 2024 to limit sync to one month
		mockClock.FixedTime = time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
		// Clear last request map to avoid wait
		svc.lastRequestMap = make(map[string]time.Time)

		var conn models.BankConnection
		db.First(&conn, "user_id = ?", "u1")

		// Setup mapping
		db.Create(&models.BankAccountMapping{
			ConnectionID:      conn.ID,
			ExternalID:        "a1",
			InternalAccountID: "internal-a1",
			IsEnabled:         true,
			Name:              "Test Acc",
		})

		// Setup user
		user := &models.User{Base: models.Base{ID: "u-sync"}, Name: "Sync User", FamilyID: "f-sync"}
		db.Create(user)
		db.Model(&conn).Updates(map[string]interface{}{"user_id": "u-sync", "family_id": "f-sync"})

		mockTxService.On("BatchCreate", mock.Anything, mock.Anything).Return(1, nil).Once()
		mockInboxService.On("AutoLinkForAccount", "internal-a1", mock.Anything).Return(0, nil).Once()

		count, err := svc.Sync("u-sync", "")
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
		mockInboxService.AssertExpectations(t)
	})
}
