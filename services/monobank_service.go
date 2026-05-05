package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

var currentSyncStatus = SyncStatus{
	IsRunning:     false,
	Message:       "Очікування...",
	TotalImported: 0,
}

type RawDataPayload struct {
	MaskedPan []string `json:"maskedPan"`
	Type      string   `json:"type"`
}

// === Service ===

type MonobankService struct {
	db          *gorm.DB
	txService   TransactionService
	accountRepo repositories.AccountRepository
}

func NewMonobankService(db *gorm.DB, txService TransactionService, accountRepo repositories.AccountRepository) *MonobankService {
	return &MonobankService{db: db, txService: txService, accountRepo: accountRepo}
}

func (s *MonobankService) GetSyncStatus() SyncStatus {
	return currentSyncStatus
}

// 1. Connect
func (s *MonobankService) Connect(userID, familyID, rawToken string) ([]MonoAccount, error) {
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

	now := time.Now()
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
// Ніяких запитів до API Monobank. Дуже швидко.
func (s *MonobankService) GetUserData(userID string) ([]MonoAccount, []models.BankAccountMapping, error) {
	var conn models.BankConnection
	// 1. Перевіряємо, чи є активне підключення
	err := s.db.Where("user_id = ? AND provider = 'monobank' AND is_active = ?", userID, true).First(&conn).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("no active connection")
		}
		return nil, nil, err
	}

	// 2. Беремо маппінги
	var mappings []models.BankAccountMapping
	s.db.Where("connection_id = ?", conn.ID).Find(&mappings)

	// 3. Формуємо "фейковий" список акаунтів на основі збережених даних
	// Це потрібно, щоб фронтенд міг відмалювати список підключених карток без запиту до банку
	var cachedAccounts []MonoAccount

	for _, m := range mappings {
		// Отримуємо баланс з нашої внутрішньої таблиці Accounts
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

		// Відновлюємо дані для фронтенду
		cachedAccounts = append(cachedAccounts, MonoAccount{
			ID:           m.ExternalID,
			CurrencyCode: currencyCode,
			Balance:      currentBalance,
			MaskedPan:    []string{"*" + m.CardNumber}, // Показуємо хоча б останні цифри
			Type:         m.CardType,
			Iban:         "",
		})
	}

	return cachedAccounts, mappings, nil
}

// 2.5. RefreshClientInfo - ЗАПИТ ДО API (Для модалки налаштувань)
// Викликається тільки коли юзер тисне "Налаштувати". Знаходить нові рахунки.
func (s *MonobankService) RefreshClientInfo(userID string) ([]MonoAccount, []models.BankAccountMapping, error) {
	var conn models.BankConnection
	if err := s.db.Where("user_id = ? AND provider = 'monobank' AND is_active = ?", userID, true).First(&conn).Error; err != nil {
		return nil, nil, errors.New("connection not found")
	}

	token, err := utils.Decrypt(conn.Token)
	if err != nil {
		return nil, nil, errors.New("failed to decrypt token")
	}

	// 1. 🔥 Робимо реальний запит до Монобанку
	clientInfo, apiErr := s.fetchClientInfo(token)
	if apiErr != nil {
		return nil, nil, apiErr // Повертаємо помилку (429 або 401), щоб фронт показав її
	}

	// 2. Завантажуємо існуючі налаштування
	var mappings []models.BankAccountMapping
	s.db.Where("connection_id = ?", conn.ID).Find(&mappings)

	// Повертаємо СВІЖІ акаунти від банку і наші налаштування
	return clientInfo.Accounts, mappings, nil
}

// 3. SaveSettings

func (s *MonobankService) SaveSettings(userID, familyID string, accounts []models.BankAccountMapping) error {
	var conn models.BankConnection
	if err := s.db.Where("user_id = ? AND provider = 'monobank'", userID).First(&conn).Error; err != nil {
		return errors.New("connection not found")
	}

	// Очищаємо старі маппінги
	s.db.Where("connection_id = ?", conn.ID).Delete(&models.BankAccountMapping{})

	for _, mapping := range accounts {
		var last4, paymentSystem string
		var cardType = "black"
		var creditLimit int64 = 0
		var shouldCreateGoal = false // 🔥 За замовчуванням FALSE

		// 🔥 ПАРСИНГ ДАНИХ
		if mapping.RawData != "" {
			var payload struct {
				MaskedPan   []string `json:"maskedPan"`
				Type        string   `json:"type"`
				CreditLimit int64    `json:"creditLimit"`
				CreateGoal  bool     `json:"createGoal"` // Читаємо прапорець
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

		// --- Логіка створення рахунку (якщо InternalID пустий) ---
		if mapping.InternalAccountID == "" {
			newAccountID := uuid.NewString()

			newAccount := &models.Account{
				Base: models.Base{
					ID:        newAccountID,
					CreatedAt: time.Now().UnixMilli(),
					UpdatedAt: time.Now().UnixMilli(),
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
				Type:           "card", // Дефолт
			}

			// 🔥 ЛОГІКА ДЛЯ БАНОК
			if cardType == "jar" {
				newAccount.Type = "piggy_bank"
				// Знаходимо тип сховища "Банка"
				storageTypeID, _ := s.getOrCreateSystemStorageType(familyID, "Банка", "jar")
				newAccount.StorageTypeID = storageTypeID

				// 🔥 ПЕРЕВІРКА ПРАПОРЦЯ + СУМИ
				if shouldCreateGoal {
					newGoal := models.Goal{
						Base: models.Base{
							ID:        uuid.NewString(),
							CreatedAt: time.Now().UnixMilli(),
							UpdatedAt: time.Now().UnixMilli(),
						},
						FamilyID:     familyID,
						Name:         mapping.Name,
						Description:  "Імпортовано з Monobank",
						TargetAmount: creditLimit, // 🔥 ВИКОРИСТОВУЄМО creditLimit ЯК ЦІЛЬ
						Currency:     mapping.Currency,
						DateStart:    time.Now().UnixMilli(),
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

		// Зберігаємо маппінг
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

// Допоміжний метод для пошуку/створення типу збереження (StorageType)
// Щоб ми могли записати, що це саме "Банка", а не "Сейф" чи "Конверт"
func (s *MonobankService) getOrCreateSystemStorageType(familyID, name, slug string) (*string, error) {
	var st models.StorageType
	
	// 1. Шукаємо системний або сімейний тип з таким слагом
	err := s.db.Where("(family_id = ? OR is_system = true) AND slug = ?", familyID, slug).First(&st).Error
	
	if err == nil {
		return &st.ID, nil
	}

	// 2. Якщо не знайшли - створюємо для сім'ї
	newSt := models.StorageType{
		Base: models.Base{
			ID:        uuid.NewString(),
			CreatedAt: time.Now().UnixMilli(),
		},
		FamilyID: &familyID,
		Name:     name,
		Slug:     slug,
		Icon:     "GiGlassJar", // Потрібно підібрати іконку з бібліотеки
		IsSystem: false,
	}
	
	if err := s.db.Create(&newSt).Error; err != nil {
		return nil, err
	}

	return &newSt.ID, nil
}
// 4. Sync (ВИПРАВЛЕНО: Оптимізація працює завжди)
func (s *MonobankService) Sync(userID string, targetAccountID string) (int, error) {
	currentSyncStatus = SyncStatus{IsRunning: true, Message: "Ініціалізація...", TotalImported: 0}

	defer func() {
		time.Sleep(2 * time.Second)
		currentSyncStatus.IsRunning = false
		currentSyncStatus.Message = "Готово"
	}()

	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return 0, fmt.Errorf("user not found: %w", err)
	}

	fmt.Println("------------------------------------------------")
	fmt.Printf("🔄 SYNC STARTED for user: %s\n", user.Name)

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
		query = s.db.Where("connection_id = ? AND is_enabled = true AND internal_account_id = ?", conn.ID, targetAccountID)
	}
	query.Find(&mappings)

	if len(mappings) == 0 {
		return 0, nil
	}

	// 🔥 ЗМІНА 1: Правильно збираємо categoryMap з префіксами income_ та expense_
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
	requestCounter := 0

	// --- 1. ІМПОРТ ТРАНЗАКЦІЙ ---
	for _, mapping := range mappings {
		if mapping.InternalAccountID == "" {
			continue
		}

		var lastTx models.Transaction
		err := s.db.Where("account_id = ?", mapping.InternalAccountID).Order("date desc").First(&lastTx).Error

		var startTime time.Time
		logReason := ""

		if err == nil {
			startTime = time.UnixMilli(lastTx.Date).Add(1 * time.Second)
			logReason = fmt.Sprintf("Last DB Transaction (%s)", startTime.Format("2006-01-02"))
		} else {
			if mapping.SyncFrom > 0 {
				startTime = time.UnixMilli(mapping.SyncFrom)
				logReason = fmt.Sprintf("User Setting SyncFrom (%s)", startTime.Format("2006-01-02"))
			} else {
				startTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				logReason = "Default (2024-01-01)"
			}
		}

		if conn.LastSync != nil && !conn.LastSync.IsZero() && conn.LastSync.After(startTime) {
			safeStart := conn.LastSync.Add(-1 * time.Hour)
			if safeStart.After(startTime) {
				startTime = safeStart
				logReason = "Global LastSync adjustment"
			}
		}

		now := time.Now()

		fmt.Printf("📆 Account [%s]: Start Date determined by [%s] -> %s\n", mapping.Name, logReason, startTime.Format("2006-01-02 15:04:05"))

		for startTime.Before(now) {
			endTime := startTime.AddDate(0, 0, 31)
			if endTime.After(now) {
				endTime = now
			}
			if endTime.Sub(startTime) < 10*time.Second {
				break
			}

			if requestCounter > 0 {
				currentSyncStatus.Message = fmt.Sprintf("Очікування API... Імпортовано: %d", totalImported)
				fmt.Println("⏳ Rate Limit prevention: Sleeping 61s...")
				time.Sleep(61 * time.Second)
			}
			requestCounter++

			currentSyncStatus.Message = fmt.Sprintf("Завантаження %s...", mapping.Name)

			fmt.Printf("📥 Fetching [%s]: %s -> %s\n",
				mapping.Name,
				startTime.Format("2006-01-02 15:04"),
				endTime.Format("2006-01-02 15:04"),
			)

			monoTxs, err := s.fetchStatement(token, mapping.ExternalID, startTime.Unix(), endTime.Unix())
			if err != nil {
				fmt.Printf("⚠️ Skip fetch for %s: %v\n", mapping.Name, err)
				startTime = endTime
				continue
			}

			if len(monoTxs) > 0 {
				fmt.Printf("   ✅ Received %d transactions\n", len(monoTxs))
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

					// 🔥 ЗМІНА 2: Використовуємо нашу універсальну функцію для підбору категорії
					normalizedName := utils.NormalizeCounterparty(mTx.Description)
					mccStr := strconv.Itoa(mTx.Mcc)
					categoryID := utils.PredictCategoryID(mTx.Description, normalizedName, mccStr, txType, categoryMap)

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
					currentSyncStatus.TotalImported = totalImported
					fmt.Printf("   💾 Saved %d new transactions to DB\n", count)
				} else {
					fmt.Println("   ⚪ All received transactions are duplicates")
				}
			} else {
				fmt.Println("   ⚪ No transactions in this period")
			}
			startTime = endTime
		}
	}

	// --- 2. ОНОВЛЕННЯ БАЛАНСІВ ---
	currentSyncStatus.Message = "Оновлення балансів..."

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
		} else {
			var calculatedBalance int64
			err := s.db.Model(&models.Transaction{}).
				Where("account_id = ?", m.InternalAccountID).
				Select(`COALESCE(SUM(CASE WHEN type IN ('expense', 'transfer_out', 'loan_give', 'debt_repay') THEN -ABS(amount) ELSE ABS(amount) END), 0)`).
				Scan(&calculatedBalance).Error

			if err == nil {
				s.db.Model(&models.Account{}).Where("id = ?", m.InternalAccountID).Update("balance", calculatedBalance)

				if calculatedBalance == 0 {
					fmt.Printf("🗑️ Soft deleting broken jar: %s\n", m.Name)
					var acc models.Account
					if s.db.First(&acc, "id = ?", m.InternalAccountID).Error == nil {
						if acc.GoalID != nil {
							var goal models.Goal
							if s.db.First(&goal, "id = ?", *acc.GoalID).Error == nil {
								var otherBalance int64
								s.db.Model(&models.Account{}).
									Where("goal_id = ? AND id != ? AND deleted_at IS NULL", goal.ID, acc.ID).
									Select("COALESCE(SUM(balance), 0)").
									Scan(&otherBalance)

								if otherBalance >= goal.TargetAmount && goal.TargetAmount > 0 {
									s.db.Model(&goal).Update("status", "reached")
								}
							}
						}
						s.db.Delete(&acc)
						s.db.Where("internal_account_id = ?", m.InternalAccountID).Delete(&models.BankAccountMapping{})
					}
				}
			}
		}
	}

	nowTime := time.Now()
	s.db.Model(&conn).Update("last_sync", nowTime)

	fmt.Printf("✅ SYNC FINISHED. Total imported: %d\n", totalImported)
	return totalImported, nil
}
// --- Helpers ---

func (s *MonobankService) fetchClientInfo(token string) (*MonoClientInfo, error) {
	req, _ := http.NewRequest("GET", "https://api.monobank.ua/personal/client-info", nil)
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

	// Декодуємо JSON
	var info MonoClientInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	// 🔥 ГОЛОВНА МАГІЯ: Об'єднуємо банки (Jars) з акаунтами (Accounts)
	// Це потрібно, щоб фронтенд бачив банки як звичайні рахунки
	for _, jar := range info.Jars {
		// Використовуємо назву банки замість номера карти
		displayPan := []string{jar.Title}
		if jar.Title == "" {
			displayPan = []string{"Скарбничка"}
		}

		jarAccount := MonoAccount{
			ID:           jar.ID,
			CurrencyCode: jar.CurrencyCode,
			Balance:      jar.Balance,
			
			// Записуємо Ціль банки у CreditLimit (це використає наш SaveSettings для створення Goal)
			CreditLimit:  jar.Goal, 
			
			MaskedPan:    displayPan,
			Type:         "jar", // Явно вказуємо тип
			Iban:         "",
		}

		// Додаємо до загального списку
		info.Accounts = append(info.Accounts, jarAccount)
	}

	return &info, nil
}

func (s *MonobankService) fetchStatement(token, accountID string, from, to int64) ([]MonoTransaction, error) {
	url := fmt.Sprintf("https://api.monobank.ua/personal/statement/%s/%d/%d", accountID, from, to)

	// 🔥 ЛОГУВАННЯ ЗАПИТУ
	fmt.Printf("   🌐 GET Request: %s\n", url)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Token", token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		fmt.Println("   🛑 Monobank API 429: Rate Limit Hit")
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


func (s *MonobankService) Disconnect(userID string) error {
  var conn models.BankConnection
  
  // Шукаємо підключення (звичайний пошук ігнорує soft-deleted, це ок)
  if err := s.db.Where("user_id = ? AND provider = 'monobank'", userID).First(&conn).Error; err != nil {
    return errors.New("connection not found")
  }

  // 1. Видаляємо маппінги (HARD DELETE)
  // Використовуємо Unscoped(), щоб стерти їх назавжди
  s.db.Unscoped().Where("connection_id = ?", conn.ID).Delete(&models.BankAccountMapping{})

  // 2. Видаляємо саме з'єднання (HARD DELETE)
  // Unscoped() ігнорує DeletedAt і робить DELETE FROM ...
  if err := s.db.Unscoped().Delete(&conn).Error; err != nil {
    return err
  }

  return nil
}
// services/monobank_service.go
// GlobalResyncCounterparties примусово перепаршує контрагентів для ВСІХ транзакцій у базі
func (s *MonobankService) GlobalResyncCounterparties() (int, error) {
  // 1. Беремо ВСІ транзакції з Монобанку по всій базі (ігноруємо користувача)
  var txs []models.Transaction
  if err := s.db.Where("external_id != '' AND deleted_at IS NULL").Find(&txs).Error; err != nil {
    return 0, err
  }

  updatedCount := 0

  for _, tx := range txs {
    // 2. Нормалізуємо назву
    correctName := utils.NormalizeCounterparty(tx.Note)
    if correctName == "" {
      continue
    }

    // 3. Шукаємо контрагента для СІМ'Ї, якій належить ЦЯ транзакція
    var cp models.Counterparty
    err := s.db.Where("family_id = ? AND name = ?", tx.FamilyID, correctName).First(&cp).Error
    
    if err != nil {
      // 🔥 ВИПРАВЛЕННЯ: Якщо контрагента немає в базі — просто пропускаємо цю транзакцію.
      // Ми більше НЕ створюємо автоматично нових контрагентів з нотаток Монобанку.
      continue 
    }

    // 4. Оновлюємо транзакцію (тільки якщо знайшли існуючого контрагента)
    if tx.CounterpartyID != cp.ID {
      s.db.Model(&models.Transaction{}).Where("id = ?", tx.ID).Update("counterparty_id", cp.ID)
      updatedCount++
    }
  }

  return updatedCount, nil
}
