package repositories

import (
	"fmt"
	"strings"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

	type TransactionFilter struct {
		FamilyID        string
		UserID          string
		AccountIDs      []string
		CategoryIDs     []string
		CounterpartyIDs []string
		Type            string
		Search          string
		DateFrom        int64
		DateTo          int64
		MinAmount       *int64
		MaxAmount       *int64
		Sort            string
		Limit           int
		Offset          int
		AssetID         string
	}

	type TransactionRepository interface {
		Create(t *models.Transaction, items []models.TransactionItem, tagIDs []string, newAsset *models.Asset) error
		CreateTransfer(from *models.Transaction, to *models.Transaction) error

		GetAll(filter TransactionFilter) ([]models.Transaction, int64, error)
		GetByID(id string, familyID string) (*models.Transaction, error)
		Update(id string, familyID string, t *models.Transaction, items []models.TransactionItem, tagIDs []string) error

		UpdateReceiptImage(id string, imagePath string) error
		GetPhotoByID(photoID string) (*models.TransactionPhoto, error)
		DeletePhotoByID(photoID string) error
		DeleteAllPhotos(txID string) error

		Delete(tx *models.Transaction, relatedTx *models.Transaction) error
		BatchCreate(transactions []models.Transaction) (int, error)
		GetPredictedCategory(familyID string, itemName string) (string, error)
	}

	type txRepo struct {
		db *gorm.DB
	}

	func NewTransactionRepository(db *gorm.DB) TransactionRepository {
		return &txRepo{db: db}
	}

	// === ДОПОМІЖНА ФУНКЦІЯ ДЛЯ БОРГІВ ===
	func calculateDebtDelta(txType string, amount int64) int64 {
		switch txType {
		case "loan_give": // Я дав у борг -> Вони винні мені більше (+)
			return amount
		case "loan_repay": // Мені повернули борг -> Вони винні мені менше (-)
			return -amount
		case "debt_take": // Я взяв у борг -> Я винен їм (баланс йде в мінус) (-)
			return -amount
		case "debt_repay": // Я повернув борг -> Я винен менше (ближче до 0) (+)
			return amount
		default:
			return 0
		}
	}

	// 🔥 ОНОВЛЕНА ЛОГІКА: Оновлює баланс у конкретній валюті
	func (r *txRepo) applyCounterpartyBalance(txDB *gorm.DB, counterpartyID string, currency string, delta int64) error {
		if counterpartyID == "" || delta == 0 {
			return nil
		}
		if currency == "" {
			currency = "UAH"
		}

		// 1. Спробуємо оновити існуючий запис
		result := txDB.Model(&models.CounterpartyBalance{}).
			Where("counterparty_id = ? AND currency = ?", counterpartyID, currency).
			Update("balance", gorm.Expr("balance + ?", delta))

		if result.Error != nil {
			return result.Error
		}

		// 2. Якщо записів для оновлення не знайдено (RowsAffected == 0), створюємо новий
		if result.RowsAffected == 0 {
			newBalance := models.CounterpartyBalance{
				CounterpartyID: counterpartyID,
				Currency:       currency,
				Balance:        delta,
			}
			return txDB.Create(&newBalance).Error
		}

		return nil
	}

	// === CREATE ===
	func (r *txRepo) Create(t *models.Transaction, items []models.TransactionItem, tagIDs []string, newAsset *models.Asset) error {
		return r.db.Transaction(func(txDB *gorm.DB) error {
			
			// 🔥 ПЕРЕВІРКА ТА ВАЛЮТА:
			if t.AccountID != "" {
				var account models.Account
				// Блокуємо запис для перевірки (FOR UPDATE), щоб уникнути race condition
				if err := txDB.Set("gorm:query_option", "FOR UPDATE").First(&account, "id = ?", t.AccountID).Error; err != nil {
					return fmt.Errorf("account not found: %w", err)
				}
				
				// Блокуємо ручне створення транзакцій для синхронізованих рахунків
				// Виняток: якщо транзакція прийшла з ExternalID (тобто це сам процес синхронізації)
				if account.IsSynced && t.ExternalID == "" {
					return fmt.Errorf("🚫 Цей рахунок (%s) синхронізується автоматично. Ручне додавання заборонено.", account.Name)
				}
				
				t.Currency = account.Currency
			} else if t.Currency == "" {
				// Якщо немає ні рахунку, ні валюти - повертаємо помилку
				return gorm.ErrInvalidField
			}

			// Обробка активів (якщо є)
			if newAsset != nil {
				newAsset.Currency = t.Currency
				if err := txDB.Create(newAsset).Error; err != nil {
					return err
				}
				t.AssetID = &newAsset.ID
			}

			if t.Type == "income" && t.AssetID != nil {
				updates := map[string]interface{}{
					"is_sold":    true,
					"sold_date":  t.Date,
					"sold_price": t.Amount,
					"updated_at": time.Now().UnixMilli(),
				}
				if err := txDB.Model(&models.Asset{}).Where("id = ?", *t.AssetID).Updates(updates).Error; err != nil {
					return err
				}
			}

			// Створюємо саму транзакцію
			if err := txDB.Create(t).Error; err != nil {
				return err
			}

			// Зберігаємо товари
			for _, item := range items {
				item.TransactionID = t.ID
				if err := txDB.Create(&item).Error; err != nil {
					return err
				}
			}

			// Зберігаємо теги
			for _, tagID := range tagIDs {
				tt := models.TransactionTag{
					Base:          models.Base{ID: uuid.NewString(), CreatedAt: time.Now().UnixMilli(), IsSynced: true},
					TransactionID: t.ID,
					TagID:         tagID,
				}
				if err := txDB.Create(&tt).Error; err != nil {
					return err
				}
			}

			// Оновлюємо рахунок (ТІЛЬКИ ЯКЩО ВІН Є)
			if t.AccountID != "" && !t.IsForgiveness {
				var change int64
				// Визначаємо знак: Витрата, Дав борг, Повернув борг, Переказ (вихід) -> МІНУС
				if t.Type == "expense" || t.Type == "loan_give" || t.Type == "debt_repay" || t.Type == "transfer_out" {
					change = -t.Amount
				} else {
					// Дохід, Взяв борг, Мені повернули, Переказ (вхід) -> ПЛЮС
					change = t.Amount
				}

				if err := txDB.Model(&models.Account{}).Where("id = ?", t.AccountID).
					Update("balance", gorm.Expr("balance + ?", change)).Error; err != nil {
					return err
				}
			}

			// 3. Оновлюємо баланс КОНТРАГЕНТА (ЗАВЖДИ)
			cpDelta := calculateDebtDelta(t.Type, t.Amount)
			if err := r.applyCounterpartyBalance(txDB, t.CounterpartyID, t.Currency, cpDelta); err != nil {
				return err
			}

			return nil
		})
	}

	// === CREATE TRANSFER ===
	func (r *txRepo) CreateTransfer(from *models.Transaction, to *models.Transaction) error {
		return r.db.Transaction(func(txDB *gorm.DB) error {
			// Отримуємо дані рахунків з блокуванням
			var accFrom, accTo models.Account
			if err := txDB.Set("gorm:query_option", "FOR UPDATE").First(&accFrom, "id = ?", from.AccountID).Error; err != nil {
				return fmt.Errorf("source account not found: %w", err)
			}
			if err := txDB.Set("gorm:query_option", "FOR UPDATE").First(&accTo, "id = ?", to.AccountID).Error; err != nil {
				return fmt.Errorf("target account not found: %w", err)
			}

			// Блокуємо перекази для синхронізованих рахунків
			if accFrom.IsSynced {
				return fmt.Errorf("🚫 Рахунок відправника (%s) синхронізується автоматично. Ручні перекази заборонено.", accFrom.Name)
			}
			if accTo.IsSynced {
				return fmt.Errorf("🚫 Рахунок отримувача (%s) синхронізується автоматично. Ручні перекази заборонено.", accTo.Name)
			}

			from.Currency = accFrom.Currency
			to.Currency = accTo.Currency

			if err := txDB.Create(from).Error; err != nil {
				return err
			}
			if err := txDB.Model(&models.Account{}).Where("id = ?", from.AccountID).
				Update("balance", gorm.Expr("balance - ?", from.Amount)).Error; err != nil {
				return err
			}

			if err := txDB.Create(to).Error; err != nil {
				return err
			}
			if err := txDB.Model(&models.Account{}).Where("id = ?", to.AccountID).
				Update("balance", gorm.Expr("balance + ?", to.Amount)).Error; err != nil {
				return err
			}
			return nil
		})
	}

	// === GET ALL ===
	func (r *txRepo) GetAll(f TransactionFilter) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var totalCount int64

	// 1. Починаємо запит і відразу приєднуємо (Join) пов'язані таблиці для пошуку
	// Використовуємо left join, щоб транзакції не зникали, якщо у них немає категорії або контрагента
	query := r.db.Debug().Model(&models.Transaction{}).
		Joins("LEFT JOIN categories ON categories.id = transactions.category_id").
		Joins("LEFT JOIN accounts ON accounts.id = transactions.account_id").
		Joins("LEFT JOIN counterparties ON counterparties.id = transactions.counterparty_id").
		Where("transactions.family_id = ? AND transactions.deleted_at IS NULL", f.FamilyID)

	if f.UserID != "" {
		query = query.Where("transactions.user_id = ?", f.UserID)
	}
	if f.AssetID != "" {
		query = query.Where("transactions.asset_id = ?", f.AssetID)
	}

	if len(f.AccountIDs) > 0 {
		query = query.Where("transactions.account_id IN ?", f.AccountIDs)
	}
	if len(f.CategoryIDs) > 0 {
		query = query.Where("transactions.category_id IN ?", f.CategoryIDs)
	}
	if len(f.CounterpartyIDs) > 0 {
		query = query.Where("transactions.counterparty_id IN ?", f.CounterpartyIDs)
	}

	if f.Type != "" {
		switch f.Type {
		case "transfer":
			query = query.Where("transactions.type IN ?", []string{"transfer", "transfer_out", "transfer_in"})
		case "expense":
			query = query.Where("transactions.type IN ?", []string{"expense", "transfer_out", "loan_give", "debt_repay"})
		case "income":
			query = query.Where("transactions.type IN ?", []string{"income", "transfer_in", "loan_repay", "debt_take"})
		case "debt":
			query = query.Where("transactions.type IN ?", []string{"loan_give", "loan_repay", "debt_take", "debt_repay"})
		default:
			query = query.Where("transactions.type = ?", f.Type)
		}
	}

	// 🔥 2. ГЛОБАЛЬНИЙ ПОШУК (Покращено підтримку Кирилиці)
	if f.Search != "" {
		search := f.Search
		lower := strings.ToLower(search)
		upper := strings.ToUpper(search)
		
		// Створюємо список паттернів для пошуку (оригінал + нижній + верхній регістр)
		// Це дозволяє обійти обмеження SQLite LIKE, який не є Case-Insensitive для Unicode
		patterns := []string{"%" + search + "%"}
		if lower != search {
			patterns = append(patterns, "%" + lower + "%")
		}
		if upper != search && upper != lower {
			patterns = append(patterns, "%" + upper + "%")
		}

		var conditions []string
		var args []interface{}
		columns := []string{"transactions.note", "categories.name", "accounts.name", "counterparties.name"}
		
		for _, col := range columns {
			for _, p := range patterns {
				conditions = append(conditions, col+" LIKE ?")
				args = append(args, p)
			}
		}
		
		query = query.Where("("+strings.Join(conditions, " OR ")+")", args...)
	}

	if f.DateFrom > 0 {
		query = query.Where("transactions.date >= ?", f.DateFrom)
	}
	if f.DateTo > 0 {
		query = query.Where("transactions.date <= ?", f.DateTo)
	}

	if f.MinAmount != nil {
		query = query.Where("transactions.amount >= ?", *f.MinAmount)
	}
	if f.MaxAmount != nil {
		query = query.Where("transactions.amount <= ?", *f.MaxAmount)
	}

	// 3. Рахуємо кількість (важливо: Count працює з тими ж Joins)
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// 4. Сортування (додаємо префікс таблиці 'transactions.', щоб уникнути ambiguity)
	switch f.Sort {
	case "amount-desc":
		query = query.Order("transactions.date desc, transactions.amount desc, transactions.id desc")
	case "amount-asc":
		query = query.Order("transactions.date desc, transactions.amount asc, transactions.id desc")
	case "date-asc":
		query = query.Order("transactions.date asc, transactions.created_at asc, transactions.id asc")
	default:
		query = query.Order("transactions.date desc, transactions.created_at desc, transactions.id desc")
	}

	// 5. Фінальний запит з Preload
	err := query.Limit(f.Limit).Offset(f.Offset).
		Preload("Category").
		Preload("User").
		Preload("Account", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("Counterparty", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("Tags").
		Preload("Items").
		Preload("Asset").
		Preload("Photos").
		Preload("TransferRelated").
		Find(&transactions).Error

	return transactions, totalCount, err
}

	// === GET BY ID ===
	func (r *txRepo) GetByID(id string, familyID string) (*models.Transaction, error) {
		var t models.Transaction
		err := r.db.
			Preload("Category").
			Preload("User").
			Preload("Account", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
			Preload("Counterparty").
			Preload("Tags").
			Preload("Items").
			Preload("Asset").
			Preload("Photos").
			Preload("TransferRelated").
			Where("id = ? AND family_id = ?", id, familyID).
			First(&t).Error
		return &t, err
	}

	// === UPDATE ===
	func (r *txRepo) Update(id string, familyID string, newData *models.Transaction, items []models.TransactionItem, tagIDs []string) error {
		return r.db.Transaction(func(txDB *gorm.DB) error {
			var oldTx models.Transaction
			// Блокуємо стару транзакцію для оновлення
			if err := txDB.Set("gorm:query_option", "FOR UPDATE").Where("id = ? AND family_id = ?", id, familyID).First(&oldTx).Error; err != nil {
				return err
			}

			// Перевірка рахунку старої транзакції
			if oldTx.AccountID != "" {
				var oldAccount models.Account
				if err := txDB.Set("gorm:query_option", "FOR UPDATE").First(&oldAccount, "id = ?", oldTx.AccountID).Error; err == nil {
					// Якщо транзакція банківська (має ExternalID), забороняємо будь-які зміни через цей метод
					// Банківські транзакції мають оновлюватись тільки через спеціальні механізми (напр. зміна категорії)
					if oldAccount.IsSynced && oldTx.ExternalID != "" {
						return fmt.Errorf("🚫 Це банківська транзакція. Ручне редагування основних полів заборонено.")
					}
				}
			}

			// Якщо валюти немає в старій транзакції, спробуємо дістати з рахунку
			if oldTx.Currency == "" && oldTx.AccountID != "" {
				var oldAccount models.Account
				if err := txDB.Select("currency").First(&oldAccount, "id = ?", oldTx.AccountID).Error; err == nil {
					oldTx.Currency = oldAccount.Currency
				}
			}

			// 1. ВІДКАТУЄМО стару транзакцію
			if oldTx.AccountID != "" && !oldTx.IsForgiveness {
				var revertChange int64
				if oldTx.Type == "expense" || oldTx.Type == "loan_give" || oldTx.Type == "debt_repay" || oldTx.Type == "transfer_out" {
					revertChange = oldTx.Amount
				} else {
					revertChange = -oldTx.Amount
				}
				if err := txDB.Model(&models.Account{}).Where("id = ?", oldTx.AccountID).
					Update("balance", gorm.Expr("balance + ?", revertChange)).Error; err != nil {
					return err
				}
			}

			// Контрагент (відкат)
			oldCpDelta := calculateDebtDelta(oldTx.Type, oldTx.Amount)
			if err := r.applyCounterpartyBalance(txDB, oldTx.CounterpartyID, oldTx.Currency, -oldCpDelta); err != nil {
				return err
			}

			// 2. Підготовка НОВИХ даних
			var newCurrency string
			if newData.AccountID != "" {
				var newAccount models.Account
				if err := txDB.Set("gorm:query_option", "FOR UPDATE").First(&newAccount, "id = ?", newData.AccountID).Error; err != nil {
					return err
				}
				
				// Перевірка: чи не намагаємось перенести на синхронізований рахунок
				if newAccount.IsSynced && newData.ExternalID == "" {
					return fmt.Errorf("🚫 Не можна переносити транзакцію на автоматичний рахунок (%s).", newAccount.Name)
				}
				
				newCurrency = newAccount.Currency
			} else {
				newCurrency = oldTx.Currency
			}

			updates := map[string]interface{}{
				"account_id":  newData.AccountID,
				"category_id": newData.CategoryID,
				"amount":      newData.Amount,
				"date":        newData.Date,
				"note":        newData.Note,
				"type":        newData.Type,
				"updated_at":  time.Now().UnixMilli(),
				"receipt_img": newData.ReceiptImg,
				"asset_id":    newData.AssetID,
				"currency":    newCurrency,
				
				// 🔥 ДОДАНО: Тепер пробіг буде оновлюватись!
				"mileage":     newData.Mileage, 
			}

			if newData.CounterpartyID == "" {
				updates["counterparty_id"] = nil
			} else {
				updates["counterparty_id"] = newData.CounterpartyID
			}

			if err := txDB.Model(&oldTx).
				Updates(updates).Error; err != nil {
				return err
			}

			// Оновлюємо Items і Tags
			if err := txDB.Where("transaction_id = ?", id).Delete(&models.TransactionItem{}).Error; err != nil {
				return err
			}
			for _, item := range items {
				item.TransactionID = id
				item.ID = uuid.NewString()
				item.CreatedAt = time.Now().UnixMilli()
				if err := txDB.Create(&item).Error; err != nil {
					return err
				}
			}

			if err := txDB.Where("transaction_id = ?", id).Delete(&models.TransactionTag{}).Error; err != nil {
				return err
			}
			for _, tagID := range tagIDs {
				tt := models.TransactionTag{
					Base:          models.Base{ID: uuid.NewString(), CreatedAt: time.Now().UnixMilli(), IsSynced: true},
					TransactionID: id,
					TagID:         tagID,
				}
				if err := txDB.Create(&tt).Error; err != nil {
					return err
				}
			}

			// 3. ЗАСТОСОВУЄМО зміни НОВОЇ транзакції
			if newData.AccountID != "" && !newData.IsForgiveness {
				var newChange int64
				if newData.Type == "expense" || newData.Type == "loan_give" || newData.Type == "debt_repay" || newData.Type == "transfer_out" {
					newChange = -newData.Amount
				} else {
					newChange = newData.Amount
				}
				if err := txDB.Model(&models.Account{}).Where("id = ?", newData.AccountID).
					Update("balance", gorm.Expr("balance + ?", newChange)).Error; err != nil {
					return err
				}
			}

			// Контрагент (застосування)
			newCpDelta := calculateDebtDelta(newData.Type, newData.Amount)
			if err := r.applyCounterpartyBalance(txDB, newData.CounterpartyID, newCurrency, newCpDelta); err != nil {
				return err
			}

			return nil
		})
	}

	// === DELETE ===
	func (r *txRepo) Delete(tx *models.Transaction, relatedTx *models.Transaction) error {
		return r.db.Transaction(func(txDB *gorm.DB) error {
			// Блокуємо основну транзакцію
			if err := txDB.Set("gorm:query_option", "FOR UPDATE").First(tx).Error; err != nil {
				return err
			}

			// Перевірка: заборона видалення банківських транзакцій
			if tx.AccountID != "" {
				var account models.Account
				if err := txDB.Set("gorm:query_option", "FOR UPDATE").First(&account, "id = ?", tx.AccountID).Error; err == nil {
					if account.IsSynced && tx.ExternalID != "" {
						return fmt.Errorf("🚫 Це банківська транзакція. Видалення заборонено.")
					}
				}
			}

			now := time.Now().UnixMilli()
			if err := txDB.Model(tx).Update("deleted_at", now).Error; err != nil {
				return err
			}

			// Якщо валюти немає, спробуємо дістати з рахунку
			if tx.Currency == "" && tx.AccountID != "" {
				var acc models.Account
				if err := txDB.Select("currency").First(&acc, "id = ?", tx.AccountID).Error; err == nil {
					tx.Currency = acc.Currency
				}
			}

			// Відкат рахунку
			if tx.AccountID != "" {
				var change int64
				if tx.Type == "expense" || tx.Type == "loan_give" || tx.Type == "debt_repay" || tx.Type == "transfer_out" || (tx.Type == "transfer" && tx.Amount > 0) {
					change = tx.Amount
				} else {
					change = -tx.Amount
				}
				if err := txDB.Model(&models.Account{}).Where("id = ?", tx.AccountID).
					Update("balance", gorm.Expr("balance + ?", change)).Error; err != nil {
					return err
				}
			}

			// Відкат боргу
			cpDelta := calculateDebtDelta(tx.Type, tx.Amount)
			if err := r.applyCounterpartyBalance(txDB, tx.CounterpartyID, tx.Currency, -cpDelta); err != nil {
				return err
			}

			// Related transaction
			if relatedTx != nil {
				if err := txDB.Model(relatedTx).Update("deleted_at", now).Error; err != nil {
					return err
				}
				if relatedTx.AccountID != "" {
					var relChange int64
					if relatedTx.Type == "expense" || relatedTx.Type == "transfer_out" {
						relChange = relatedTx.Amount
					} else {
						relChange = -relatedTx.Amount
					}
					if err := txDB.Model(&models.Account{}).Where("id = ?", relatedTx.AccountID).
						Update("balance", gorm.Expr("balance + ?", relChange)).Error; err != nil {
						return err
					}
				}
			}

			return nil
		})
	}

	// === (РЕШТА ФУНКЦІЙ) ===
	func (r *txRepo) UpdateReceiptImage(id string, imagePath string) error {
		return r.db.Transaction(func(txDB *gorm.DB) error {
			if err := txDB.Model(&models.Transaction{}).Where("id = ?", id).Update("receipt_img", imagePath).Error; err != nil {
				return err
			}
			if imagePath != "" {
				photo := models.TransactionPhoto{
					Base:          models.Base{ID: uuid.NewString(), CreatedAt: time.Now().UnixMilli(), IsSynced: true},
					TransactionID: id,
					Path:          imagePath,
				}
				return txDB.Create(&photo).Error
			}
			return nil
		})
	}

	func (r *txRepo) GetPhotoByID(photoID string) (*models.TransactionPhoto, error) {
		var photo models.TransactionPhoto
		err := r.db.Joins("JOIN transactions ON transactions.id = transaction_photos.transaction_id").Where("transaction_photos.id = ?", photoID).First(&photo).Error
		return &photo, err
	}

	func (r *txRepo) DeletePhotoByID(photoID string) error {
		return r.db.Transaction(func(txDB *gorm.DB) error {
			var photo models.TransactionPhoto
			if err := txDB.First(&photo, "id = ?", photoID).Error; err != nil {
				return err
			}
			txDB.Model(&models.Transaction{}).Where("id = ? AND receipt_img = ?", photo.TransactionID, photo.Path).Update("receipt_img", "")
			return txDB.Delete(&models.TransactionPhoto{}, "id = ?", photoID).Error
		})
	}

	func (r *txRepo) DeleteAllPhotos(txID string) error {
		return r.db.Transaction(func(txDB *gorm.DB) error {
			if err := txDB.Model(&models.Transaction{}).Where("id = ?", txID).Update("receipt_img", "").Error; err != nil {
				return err
			}
			if err := txDB.Where("transaction_id = ?", txID).Delete(&models.TransactionPhoto{}).Error; err != nil {
				return err
			}
			return nil
		})
	}

	func (r *txRepo) BatchCreate(transactions []models.Transaction) (int, error) {
			createdCount := 0
			err := r.db.Transaction(func(txDB *gorm.DB) error {
					for _, t := range transactions {
							// 🔥 Якщо CounterpartyID — пустий рядок, примусово робимо його nil для БД
							if t.CounterpartyID == "" {
									// Це спрацює, якщо в моделі Transaction поле CounterpartyID має тип *string
									// Якщо ж тип просто string, то в базу запишеться пустий рядок, що теж ок,
									// але GORM при Preload може намагатися шукати порожній ID.
							}
							if err := txDB.Create(&t).Error; err != nil {
									return err
							}
							createdCount++
					}
					return nil
			})
			return createdCount, err
	}

	func (r *txRepo) GetPredictedCategory(familyID string, itemName string) (string, error) {
		type ItemRow struct {
			Name       string
			CategoryID string
		}
		var rows []ItemRow
		query := `SELECT ti.name, ti.category_id FROM transaction_items ti JOIN transactions t ON ti.transaction_id = t.id WHERE t.family_id = ? AND t.deleted_at IS NULL AND ti.category_id IS NOT NULL AND ti.category_id != '' ORDER BY t.date DESC LIMIT 1000`
		if err := r.db.Raw(query, familyID).Scan(&rows).Error; err != nil {
			return "", err
		}

		search := strings.ToLower(strings.TrimSpace(itemName))
		if search == "" {
			return "", nil
		}

		categoryCounts := make(map[string]int)
		for _, row := range rows {
			dbName := strings.ToLower(row.Name)
			if strings.HasPrefix(dbName, search) {
				categoryCounts[row.CategoryID] += 5
			} else if strings.Contains(dbName, " "+search) {
				categoryCounts[row.CategoryID] += 3
			} else if strings.Contains(dbName, search) {
				categoryCounts[row.CategoryID] += 1
			}
		}
		var bestCategoryID string
		maxPoints := 0
		for catID, points := range categoryCounts {
			if points > maxPoints {
				maxPoints = points
				bestCategoryID = catID
			}
		}
		return bestCategoryID, nil
	}