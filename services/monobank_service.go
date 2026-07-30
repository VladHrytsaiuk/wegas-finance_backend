package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// === External API Types ===
type MonoClientInfo struct {
	ClientID string        `json:"clientId"`
	Name     string        `json:"name"`
	Accounts []MonoAccount `json:"accounts"`
	Jars     []MonoJar     `json:"jars"`
}
type MonoJar struct {
	ID           string `json:"id"`
	SendId       string `json:"sendId"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	CurrencyCode int    `json:"currencyCode"`
	Balance      int64  `json:"balance"`
	Goal         int64  `json:"goal"`
}

type MonoAccount struct {
	ID           string   `json:"id"`
	CurrencyCode int      `json:"currencyCode"`
	CashbackType string   `json:"cashbackType"`
	Balance      int64    `json:"balance"`
	CreditLimit  int64    `json:"creditLimit"`
	MaskedPan    []string `json:"maskedPan"`
	Type         string   `json:"type"`
	Iban         string   `json:"iban"`
}

type MonoTransaction struct {
	ID              string `json:"id"`
	Time            int64  `json:"time"`
	Description     string `json:"description"`
	Mcc             int    `json:"mcc"`
	Amount          int64  `json:"amount"`
	OperationAmount int64  `json:"operationAmount"`
	CurrencyCode    int    `json:"currencyCode"`
	CommissionRate  int64  `json:"commissionRate"`
	CashbackAmount  int64  `json:"cashbackAmount"`
	Balance         int64  `json:"balance"`
	Comment         string `json:"comment"`
}

type SyncStatus struct {
	IsRunning     bool   `json:"is_running"`
	Message       string `json:"message"`
	TotalImported int    `json:"total_imported"`
}

type RawDataPayload struct {
	MaskedPan []string `json:"maskedPan"`
	Type      string   `json:"type"`
}

// === Webhook Payload ===
type MonoWebhookPayload struct {
	Type string `json:"type"`
	Data struct {
		Account       string          `json:"account"`
		StatementItem MonoTransaction `json:"statementItem"`
	} `json:"data"`
}

// === Service ===

type MonobankService interface {
	Connect(userID, familyID, rawToken string) ([]MonoAccount, error)
	GetUserData(userID string) ([]MonoAccount, []models.BankAccountMapping, error)
	RefreshClientInfo(userID string) ([]MonoAccount, []models.BankAccountMapping, error)
	SaveSettings(userID, familyID string, accounts []models.BankAccountMapping) error
	Sync(userID string, targetAccountID string) (int, error)
	Disconnect(userID string) error
	GetSyncStatus(userID string) SyncStatus
	GlobalResyncCounterparties() (int, error)
	ProcessWebhook(payload MonoWebhookPayload) error
}

type monobankService struct {
	db            *gorm.DB
	txService     TransactionService
	inboxService  InboxService
	accountRepo   repositories.AccountRepository
	clock         utils.Clock
	baseURL       string // <--- Додано для тестів
	SkipRateLimit bool   // <--- Додано для тестів

	// Per-user sync status
	mu      sync.RWMutex
	syncMap map[string]SyncStatus

	// Rate limiting (per token)
	lastRequestMu  sync.Mutex
	lastRequestMap map[string]time.Time
}

func NewMonobankService(db *gorm.DB, txService TransactionService, accountRepo repositories.AccountRepository, clock utils.Clock, inboxServices ...InboxService) MonobankService {
	service := &monobankService{
		db:             db,
		txService:      txService,
		accountRepo:    accountRepo,
		clock:          clock,
		baseURL:        "https://api.monobank.ua",
		syncMap:        make(map[string]SyncStatus),
		lastRequestMap: make(map[string]time.Time),
	}
	if len(inboxServices) > 0 {
		service.inboxService = inboxServices[0]
	}
	return service
}

// waitForMonobankLimit гарантує, що між запитами до Mono API з одним токеном минає мінімум 61 секунда
func (s *monobankService) waitForMonobankLimit(token string) {
	if s.SkipRateLimit {
		return
	}
	s.lastRequestMu.Lock()
	defer s.lastRequestMu.Unlock()

	lastReq, exists := s.lastRequestMap[token]
	if !exists {
		s.lastRequestMap[token] = s.clock.Now()
		return
	}

	elapsed := s.clock.Now().Sub(lastReq)
	waitTime := 61*time.Second - elapsed

	if waitTime > 0 {
		fmt.Printf("⏳ Rate Limit: Waiting %v for token...\n", waitTime)
		time.Sleep(waitTime)
	}

	s.lastRequestMap[token] = s.clock.Now()
}

func (s *monobankService) GetSyncStatus(userID string) SyncStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if status, ok := s.syncMap[userID]; ok {
		return status
	}
	return SyncStatus{Message: "Очікування..."}
}

func (s *monobankService) updateStatus(userID string, status SyncStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.syncMap[userID] = status
}

// 1. Connect
func (s *monobankService) Connect(userID, familyID, rawToken string) ([]MonoAccount, error) {
	clientInfo, err := s.fetchClientInfo(rawToken)
	if err != nil {
		return nil, err
	}

	encryptedToken, err := utils.Encrypt(rawToken)
	if err != nil {
		return nil, err
	}

	var conn models.BankConnection
	result := s.db.Where("user_id = ? AND provider = 'monobank'", userID).First(&conn)

	now := s.clock.Now()
	conn.UserID = userID
	conn.FamilyID = familyID
	conn.Provider = "monobank"
	conn.Token = encryptedToken
	conn.IsActive = true

	if result.Error != nil {
		conn.ID = uuid.NewString()
		conn.CreatedAt = now
		conn.UpdatedAt = now
		if err := s.db.Create(&conn).Error; err != nil {
			return nil, err
		}
	} else {
		conn.UpdatedAt = now
		if err := s.db.Save(&conn).Error; err != nil {
			return nil, err
		}
	}

	return clientInfo.Accounts, nil
}

// 2. GetUserData - ТІЛЬКИ З БАЗИ (Для відображення статусу в профілі)
func (s *monobankService) GetUserData(userID string) ([]MonoAccount, []models.BankAccountMapping, error) {
	var conn models.BankConnection
	err := s.db.Where("user_id = ? AND provider = 'monobank' AND is_active = ?", userID, true).First(&conn).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("no active connection")
		}
		return nil, nil, err
	}

	var mappings []models.BankAccountMapping
	s.db.Where("connection_id = ?", conn.ID).Find(&mappings)

	var cachedAccounts []MonoAccount
	for _, m := range mappings {
		var internalAcc models.Account
		currentBalance := int64(0)
		currencyCode := 980

		if m.InternalAccountID != "" {
			if err := s.db.First(&internalAcc, "id = ?", m.InternalAccountID).Error; err == nil {
				currentBalance = internalAcc.Balance
				if internalAcc.Currency == "USD" {
					currencyCode = 840
				} else if internalAcc.Currency == "EUR" {
					currencyCode = 978
				}
			}
		}

		cachedAccounts = append(cachedAccounts, MonoAccount{
			ID:           m.ExternalID,
			CurrencyCode: currencyCode,
			Balance:      currentBalance,
			MaskedPan:    []string{"*" + m.CardNumber},
			Type:         m.CardType,
			Iban:         "",
		})
	}

	return cachedAccounts, mappings, nil
}

func (s *monobankService) RefreshClientInfo(userID string) ([]MonoAccount, []models.BankAccountMapping, error) {
	var conn models.BankConnection
	if err := s.db.Where("user_id = ? AND provider = 'monobank' AND is_active = ?", userID, true).First(&conn).Error; err != nil {
		return nil, nil, errors.New("connection not found")
	}

	token, err := utils.Decrypt(conn.Token)
	if err != nil {
		return nil, nil, errors.New("failed to decrypt token")
	}

	clientInfo, apiErr := s.fetchClientInfo(token)
	if apiErr != nil {
		return nil, nil, apiErr
	}

	var mappings []models.BankAccountMapping
	s.db.Where("connection_id = ?", conn.ID).Find(&mappings)

	return clientInfo.Accounts, mappings, nil
}

func (s *monobankService) SaveSettings(userID, familyID string, accounts []models.BankAccountMapping) error {
	var conn models.BankConnection
	if err := s.db.Where("user_id = ? AND provider = 'monobank'", userID).First(&conn).Error; err != nil {
		return errors.New("connection not found")
	}

	s.db.Where("connection_id = ?", conn.ID).Delete(&models.BankAccountMapping{})

	for _, mapping := range accounts {
		var last4, paymentSystem string
		var cardType = "black"
		var creditLimit int64 = 0
		var shouldCreateGoal = false

		if mapping.RawData != "" {
			var payload struct {
				MaskedPan   []string `json:"maskedPan"`
				Type        string   `json:"type"`
				CreditLimit int64    `json:"creditLimit"`
				CreateGoal  bool     `json:"createGoal"`
			}
			if err := json.Unmarshal([]byte(mapping.RawData), &payload); err == nil {
				if payload.Type != "" {
					cardType = payload.Type
				}
				if len(payload.MaskedPan) > 0 {
					last4, paymentSystem = utils.ParseCardDetails(payload.MaskedPan)
				}
				creditLimit = payload.CreditLimit
				shouldCreateGoal = payload.CreateGoal
			}
		}

		if last4 == "" && mapping.CardNumber != "" {
			last4 = mapping.CardNumber
		}

		if mapping.InternalAccountID == "" {
			newAccountID := uuid.NewString()

			newAccount := &models.Account{
				Base: models.Base{
					ID:        newAccountID,
					CreatedAt: s.clock.NowUnixMilli(),
					UpdatedAt: s.clock.NowUnixMilli(),
					IsSynced:  true,
				},
				UserID:         userID,
				FamilyID:       familyID,
				Name:           mapping.Name,
				Currency:       mapping.Currency,
				InitialBalance: 0,
				Balance:        0,
				Color:          "#000000",
				BankName:       "monobank",
				CardType:       cardType,
				CardNumber:     last4,
				PaymentSystem:  paymentSystem,
				IsArchived:     false,
				Type:           "card",
			}

			if cardType == "jar" {
				newAccount.Type = "piggy_bank"
				storageTypeID, _ := s.getOrCreateSystemStorageType(familyID, "Банка", "jar")
				newAccount.StorageTypeID = storageTypeID

				if shouldCreateGoal {
					newGoal := models.Goal{
						Base: models.Base{
							ID:        uuid.NewString(),
							CreatedAt: s.clock.NowUnixMilli(),
							UpdatedAt: s.clock.NowUnixMilli(),
						},
						FamilyID:     familyID,
						Name:         mapping.Name,
						Description:  "Імпортовано з Monobank",
						TargetAmount: creditLimit,
						Currency:     mapping.Currency,
						DateStart:    s.clock.NowUnixMilli(),
						Color:        "#ea5353",
						Icon:         "HiSortAscending",
						Status:       "active",
					}

					if err := s.db.Create(&newGoal).Error; err != nil {
						return fmt.Errorf("failed to create goal for jar: %w", err)
					}
					newAccount.GoalID = &newGoal.ID
				}
			}

			if err := s.accountRepo.Create(newAccount); err != nil {
				return fmt.Errorf("failed to create account: %w", err)
			}
			mapping.InternalAccountID = newAccount.ID
		}

		mapping.ID = uuid.NewString()
		mapping.ConnectionID = conn.ID
		mapping.CardNumber = last4
		mapping.PaymentSystem = paymentSystem
		mapping.BankName = "monobank"
		mapping.CardType = cardType

		if err := s.db.Create(&mapping).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *monobankService) getOrCreateSystemStorageType(familyID, name, slug string) (*string, error) {
	var st models.StorageType
	err := s.db.Where("(family_id = ? OR is_system = true) AND slug = ?", familyID, slug).First(&st).Error
	if err == nil {
		return &st.ID, nil
	}
	newSt := models.StorageType{
		Base: models.Base{
			ID:        uuid.NewString(),
			CreatedAt: s.clock.NowUnixMilli(),
		},
		FamilyID: &familyID,
		Name:     name,
		Slug:     slug,
		Icon:     "GiGlassJar",
		IsSystem: false,
	}
	if err := s.db.Create(&newSt).Error; err != nil {
		return nil, err
	}
	return &newSt.ID, nil
}

func (s *monobankService) Sync(userID string, targetAccountID string) (int, error) {
	s.mu.Lock()
	if status, ok := s.syncMap[userID]; ok && status.IsRunning {
		s.mu.Unlock()
		return 0, nil
	}
	s.mu.Unlock()

	status := SyncStatus{IsRunning: true, Message: "Ініціалізація...", TotalImported: 0}
	s.updateStatus(userID, status)

	defer func() {
		status.IsRunning = false
		status.Message = "Готово"
		s.updateStatus(userID, status)
	}()

	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return 0, fmt.Errorf("user not found: %w", err)
	}

	var conn models.BankConnection
	if err := s.db.Where("user_id = ? AND provider = 'monobank'", userID).First(&conn).Error; err != nil {
		return 0, errors.New("connection not found")
	}

	token, err := utils.Decrypt(conn.Token)
	if err != nil {
		return 0, errors.New("failed to decrypt token")
	}

	var mappings []models.BankAccountMapping
	query := s.db.Where("connection_id = ? AND is_enabled = true", conn.ID)
	if targetAccountID != "" {
		query = query.Where("internal_account_id = ?", targetAccountID)
	}
	query.Find(&mappings)

	if len(mappings) == 0 {
		return 0, nil
	}

	categoryMap := make(map[string]string)
	var categories []models.Category
	if err := s.db.Where("family_id = ?", user.FamilyID).Find(&categories).Error; err == nil {
		for _, cat := range categories {
			exactKey := fmt.Sprintf("%s_%s", cat.Type, strings.ToLower(cat.Name))
			categoryMap[exactKey] = cat.ID
			categoryMap[strings.ToLower(cat.Name)] = cat.ID
		}
	}

	totalImported := 0
	for _, mapping := range mappings {
		if mapping.InternalAccountID == "" {
			continue
		}

		var lastTx models.Transaction
		err := s.db.Where("account_id = ?", mapping.InternalAccountID).Order("date desc").First(&lastTx).Error

		var startTime time.Time
		if err == nil {
			startTime = time.UnixMilli(lastTx.Date).Add(1 * time.Second)
		} else {
			if mapping.SyncFrom > 0 {
				startTime = time.UnixMilli(mapping.SyncFrom)
			} else {
				startTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			}
		}

		now := s.clock.Now()
		for startTime.Before(now) {
			endTime := startTime.AddDate(0, 0, 31)
			if endTime.After(now) {
				endTime = now
			}
			if endTime.Sub(startTime) < 10*time.Second {
				break
			}

			status.Message = fmt.Sprintf("Завантаження %s...", mapping.Name)
			s.updateStatus(userID, status)

			monoTxs, err := s.fetchStatement(token, mapping.ExternalID, startTime.Unix(), endTime.Unix())
			if err != nil {
				startTime = endTime
				continue
			}

			if len(monoTxs) > 0 {
				var inputs []CreateTransactionInput
				for _, mTx := range monoTxs {
					var exists int64
					s.db.Model(&models.Transaction{}).
						Where("external_id = ? AND account_id = ?", mTx.ID, mapping.InternalAccountID).
						Count(&exists)
					if exists > 0 {
						continue
					}

					txType := "income"
					if mTx.Amount < 0 {
						txType = "expense"
					}

					normalizedName := utils.NormalizeCounterparty(mTx.Description)
					mccStr := strconv.Itoa(mTx.Mcc)
					categoryID := utils.PredictCategoryID(mTx.Description, normalizedName, mccStr, "", txType, categoryMap)

					inputs = append(inputs, CreateTransactionInput{
						AccountID:        mapping.InternalAccountID,
						Amount:           mTx.Amount,
						Date:             mTx.Time * 1000,
						Note:             mTx.Description,
						Type:             txType,
						CategoryID:       categoryID,
						CounterpartyName: normalizedName,
						ExternalID:       mTx.ID,
					})
				}

				if len(inputs) > 0 {
					count, _ := s.txService.BatchCreate(inputs, &user)
					totalImported += count
					if count > 0 && s.inboxService != nil {
						// Matching must never stop bank synchronization. Ambiguous
						// receipts remain in Inbox for the user to review.
						_, _ = s.inboxService.AutoLinkForAccount(mapping.InternalAccountID, &user)
					}
					status.TotalImported = totalImported
					s.updateStatus(userID, status)
				}
			}
			startTime = endTime
		}
	}

	status.Message = "Оновлення балансів..."
	s.updateStatus(userID, status)

	clientInfo, err := s.fetchClientInfo(token)
	apiBalances := make(map[string]int64)
	if err == nil {
		for _, monoAcc := range clientInfo.Accounts {
			apiBalances[monoAcc.ID] = monoAcc.Balance
		}
	}

	for _, m := range mappings {
		if m.InternalAccountID == "" {
			continue
		}
		if newBalance, exists := apiBalances[m.ExternalID]; exists {
			s.db.Model(&models.Account{}).Where("id = ?", m.InternalAccountID).Update("balance", newBalance)
		}
	}

	s.db.Model(&conn).Update("last_sync", s.clock.Now())
	return totalImported, nil
}

func (s *monobankService) fetchClientInfo(token string) (*MonoClientInfo, error) {
	s.waitForMonobankLimit(token)
	req, _ := http.NewRequest("GET", s.baseURL+"/personal/client-info", nil)
	req.Header.Set("X-Token", token)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api error: %d", resp.StatusCode)
	}

	var info MonoClientInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	for _, jar := range info.Jars {
		displayPan := []string{jar.Title}
		if jar.Title == "" {
			displayPan = []string{"Скарбничка"}
		}
		info.Accounts = append(info.Accounts, MonoAccount{
			ID:           jar.ID,
			CurrencyCode: jar.CurrencyCode,
			Balance:      jar.Balance,
			CreditLimit:  jar.Goal,
			MaskedPan:    displayPan,
			Type:         "jar",
		})
	}
	return &info, nil
}

func (s *monobankService) fetchStatement(token, accountID string, from, to int64) ([]MonoTransaction, error) {
	s.waitForMonobankLimit(token)
	url := fmt.Sprintf("%s/personal/statement/%s/%d/%d", s.baseURL, accountID, from, to)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Token", token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limit")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api error: %d", resp.StatusCode)
	}

	var txs []MonoTransaction
	if err := json.NewDecoder(resp.Body).Decode(&txs); err != nil {
		return nil, err
	}
	return txs, nil
}

func (s *monobankService) Disconnect(userID string) error {
	var conn models.BankConnection
	if err := s.db.Where("user_id = ? AND provider = 'monobank'", userID).First(&conn).Error; err != nil {
		return errors.New("connection not found")
	}
	s.db.Unscoped().Where("connection_id = ?", conn.ID).Delete(&models.BankAccountMapping{})
	return s.db.Unscoped().Delete(&conn).Error
}

func (s *monobankService) GlobalResyncCounterparties() (int, error) {
	var txs []models.Transaction
	if err := s.db.Where("external_id != '' AND deleted_at IS NULL").Find(&txs).Error; err != nil {
		return 0, err
	}
	updatedCount := 0
	for _, tx := range txs {
		correctName := utils.NormalizeCounterparty(tx.Note)
		if correctName == "" {
			continue
		}
		var cp models.Counterparty
		if err := s.db.Where("family_id = ? AND name = ?", tx.FamilyID, correctName).First(&cp).Error; err == nil {
			if tx.CounterpartyID != cp.ID {
				s.db.Model(&models.Transaction{}).Where("id = ?", tx.ID).Update("counterparty_id", cp.ID)
				updatedCount++
			}
		}
	}
	return updatedCount, nil
}

func (s *monobankService) ProcessWebhook(payload MonoWebhookPayload) error {
	if payload.Type != "StatementItem" {
		return nil
	}
	mTx := payload.Data.StatementItem
	externalAccountID := payload.Data.Account

	var mapping models.BankAccountMapping
	if err := s.db.Where("external_id = ? AND is_enabled = true", externalAccountID).First(&mapping).Error; err != nil {
		return fmt.Errorf("account mapping not found or disabled: %w", err)
	}

	var conn models.BankConnection
	if err := s.db.First(&conn, "id = ?", mapping.ConnectionID).Error; err != nil {
		return fmt.Errorf("connection not found: %w", err)
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", conn.UserID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	var exists int64
	s.db.Model(&models.Transaction{}).Where("external_id = ? AND account_id = ?", mTx.ID, mapping.InternalAccountID).Count(&exists)
	if exists > 0 {
		return nil
	}

	txType := "income"
	if mTx.Amount < 0 {
		txType = "expense"
	}

	categoryMap := make(map[string]string)
	var categories []models.Category
	if err := s.db.Where("family_id = ?", user.FamilyID).Find(&categories).Error; err == nil {
		for _, cat := range categories {
			exactKey := fmt.Sprintf("%s_%s", cat.Type, strings.ToLower(cat.Name))
			categoryMap[exactKey] = cat.ID
			categoryMap[strings.ToLower(cat.Name)] = cat.ID
		}
	}

	normalizedName := utils.NormalizeCounterparty(mTx.Description)
	categoryID := utils.PredictCategoryID(mTx.Description, normalizedName, strconv.Itoa(mTx.Mcc), "", txType, categoryMap)

	input := CreateTransactionInput{
		AccountID:        mapping.InternalAccountID,
		Amount:           mTx.Amount,
		Date:             mTx.Time * 1000,
		Note:             mTx.Description,
		Type:             txType,
		CategoryID:       categoryID,
		CounterpartyName: normalizedName,
		ExternalID:       mTx.ID,
	}

	_, err := s.txService.Create(input, nil, &user)
	if err != nil {
		return fmt.Errorf("failed to create transaction from webhook: %w", err)
	}

	s.db.Model(&models.Account{}).Where("id = ?", mapping.InternalAccountID).Update("balance", mTx.Balance)
	return nil
}
