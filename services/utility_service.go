package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
)

type UtilityService interface {
	CreateMeter(input models.UtilityMeter, user *models.User) error
	GetMeters(user *models.User) ([]models.UtilityMeter, error)
	GetMeterByID(id string, user *models.User) (*models.UtilityMeter, error)
	UpdateMeter(id string, input models.UtilityMeter, user *models.User) error
	DeleteMeter(id string, user *models.User) error

	CreateReading(input models.UtilityReading, user *models.User) error
	GetReadings(user *models.User, meterID string) ([]models.UtilityReading, error)
	UpdateReading(id string, input models.UtilityReading, user *models.User) error
	DeleteReading(id string, user *models.User) error
	PayReading(readingID string, accountID string, user *models.User) error

	GetGlobalStats(user *models.User) ([]models.UtilityStatGlobalDTO, error)
	GetMeterStats(meterID string, user *models.User) ([]models.UtilityStatMeterDTO, error)
}

type utilityService struct {
	repo      repositories.UtilityRepository
	txRepo    repositories.TransactionRepository
	assetRepo repositories.AssetRepository
}

func NewUtilityService(
	repo repositories.UtilityRepository,
	txRepo repositories.TransactionRepository,
	assetRepo repositories.AssetRepository,
) UtilityService {
	return &utilityService{
		repo:      repo,
		txRepo:    txRepo,
		assetRepo: assetRepo,
	}
}

// --- METERS ---

func (s *utilityService) CreateMeter(input models.UtilityMeter, user *models.User) error {
	if input.NewAsset != nil {
		newAsset := &models.Asset{
			Base:         models.Base{ID: uuid.NewString(), CreatedAt: time.Now().UnixMilli(), IsSynced: true},
			FamilyID:     user.FamilyID,
			UserID:       user.ID,
			Name:         input.NewAsset.Name,
			Type:         input.NewAsset.Type,
			SerialNumber: input.NewAsset.SerialNumber,
			WarrantyEnd:  input.NewAsset.WarrantyEnd,
			Note:         input.NewAsset.Note,
			Currency:     "UAH",
			IsSold:       false,
		}

		if err := s.assetRepo.Create(newAsset); err != nil {
			return err
		}
		input.AssetID = &newAsset.ID
	}

	input.ID = uuid.NewString()
	input.FamilyID = user.FamilyID
	input.CreatedAt = time.Now().UnixMilli()
	input.IsSynced = true
	return s.repo.CreateMeter(&input)
}

func (s *utilityService) GetMeters(user *models.User) ([]models.UtilityMeter, error) {
	return s.repo.GetMeters(user.FamilyID)
}

func (s *utilityService) GetMeterByID(id string, user *models.User) (*models.UtilityMeter, error) {
	return s.repo.GetMeterByID(id, user.FamilyID)
}

func (s *utilityService) UpdateMeter(id string, input models.UtilityMeter, user *models.User) error {
	existing, err := s.repo.GetMeterByID(id, user.FamilyID)
	if err != nil {
		return err
	}

	existing.Asset = nil
	existing.Counterparty = nil

	if input.NewAsset != nil {
		newAsset := &models.Asset{
			Base:         models.Base{ID: uuid.NewString(), CreatedAt: time.Now().UnixMilli(), IsSynced: true},
			FamilyID:     user.FamilyID,
			UserID:       user.ID,
			Name:         input.NewAsset.Name,
			Type:         input.NewAsset.Type,
			SerialNumber: input.NewAsset.SerialNumber,
			WarrantyEnd:  input.NewAsset.WarrantyEnd,
			Note:         input.NewAsset.Note,
			Currency:     "UAH",
			IsSold:       false,
		}
		if err := s.assetRepo.Create(newAsset); err != nil {
			return err
		}
		existing.AssetID = &newAsset.ID
	} else {
		existing.AssetID = input.AssetID
	}

	existing.Name = input.Name
	existing.CounterpartyID = input.CounterpartyID
	existing.PersonalAccount = input.PersonalAccount
	existing.Type = input.Type
	existing.Unit = input.Unit
	existing.Tariff = input.Tariff
	existing.UpdatedAt = time.Now().UnixMilli()

	return s.repo.UpdateMeter(existing)
}

func (s *utilityService) DeleteMeter(id string, user *models.User) error {
	return s.repo.DeleteMeter(id, user.FamilyID)
}

// --- READINGS ---

func (s *utilityService) CreateReading(input models.UtilityReading, user *models.User) error {
	// 1. Отримуємо лічильник
	meter, err := s.repo.GetMeterByID(input.MeterID, user.FamilyID)
	if err != nil {
		return errors.New("meter not found")
	}

	// === 🔥 НОВА ЛОГІКА ТАРИФІВ START ===
	
	// Визначаємо, який тариф використовувати.
	// input.TariffAtDate приходить з фронтенду (з модалки).
	effectiveTariff := input.TariffAtDate

	// Якщо фронт нічого не прислав (0), або це перший запис, беремо дефолтний з лічильника
	if effectiveTariff <= 0 {
		effectiveTariff = meter.Tariff
	}

	// Якщо введений тариф відрізняється від того, що записаний в лічильнику:
	// 1. Оновлюємо лічильник (щоб наступного разу підтягнувся вже новий тариф).
	// 2. Цей показник запишеться вже з новим тарифом.
	if effectiveTariff != meter.Tariff {
		meter.Tariff = effectiveTariff
		// Зберігаємо зміну тарифу лічильника в БД
		if err := s.repo.UpdateMeter(meter); err != nil {
			return fmt.Errorf("failed to update meter tariff: %w", err)
		}
	}
	// === 🔥 НОВА ЛОГІКА ТАРИФІВ END ===

	// Валідація показників (попередній/наступний)
	prevReading, _ := s.repo.GetPreviousReading(input.MeterID, input.Date)
	nextReading, _ := s.repo.GetNextReading(input.MeterID, input.Date)

	if prevReading != nil && input.Value < prevReading.Value {
		return errors.New("value cannot be lower than previous reading")
	}

	// Розрахунок різниці (Diff)
	var diff float64 = 0
	if prevReading != nil {
		diff = input.Value - prevReading.Value
	} else {
		// Якщо це перший показник, вважаємо, що спожито стільки, скільки введено
		// АБО diff = 0, якщо хочеш починати відлік з нуля. 
		// Зазвичай перший показник - це "точка відліку", тому diff логічніше 0, 
		// але тут залишаю твою логіку:
		diff = input.Value 
	}

	// Розрахунок вартості
	// 🔥 Використовуємо effectiveTariff, який ми визначили вище
	costInCents := int64(diff * effectiveTariff * 100)

	// Заповнюємо поля моделі
	input.ID = uuid.NewString()
	input.CreatedAt = time.Now().UnixMilli()
	input.IsSynced = true
	input.Diff = diff
	
	// 🔥 Зберігаємо саме той тариф, який був актуальний на момент вводу (ІСТОРІЯ)
	input.TariffAtDate = effectiveTariff 
	
	input.CalculatedCost = costInCents
	input.IsPaid = false
	input.PaymentTransactionID = nil

	// Створення транзакції БОРГУ (debt_take)
	if meter.CounterpartyID != nil && costInCents > 0 {
		txID := uuid.NewString()
		newDebt := &models.Transaction{
			Base:           models.Base{ID: txID, CreatedAt: time.Now().UnixMilli(), IsSynced: true},
			FamilyID:       user.FamilyID,
			UserID:         user.ID,
			Type:           "debt_take",
			Amount:         costInCents,
			Currency:       meter.Currency,
			Date:           input.Date,
			CounterpartyID: *meter.CounterpartyID,
			// Додаємо тариф у примітку для ясності
			Note:           fmt.Sprintf("%s: %.2f %s (x %.2f)", meter.Name, input.Diff, meter.Unit, effectiveTariff),
			AssetID:        meter.AssetID,
		}
		if err := s.txRepo.Create(newDebt, nil, nil, nil); err != nil {
			return err
		}
		input.TransactionID = &txID
	}

	// Зберігаємо показник
	if err := s.repo.CreateReading(&input); err != nil {
		return err
	}

	// Оновлення кешу останніх показників в лічильнику
	if nextReading == nil {
		meter.LastReadingDate = &input.Date
		meter.LastReadingValue = &input.Value
		// Тут тариф вже оновлений вище, якщо треба було, тому просто зберігаємо
		_ = s.repo.UpdateMeter(meter)
	}

	return nil
}

func (s *utilityService) GetReadings(user *models.User, meterID string) ([]models.UtilityReading, error) {
	return s.repo.GetReadings(user.FamilyID, meterID)
}

func (s *utilityService) UpdateReading(id string, input models.UtilityReading, user *models.User) error {
	updates := map[string]interface{}{
		"updated_at": time.Now().UnixMilli(),
		"is_paid":    input.IsPaid,
	}

	// Якщо ми вручну прив'язуємо або змінюємо ID транзакцій
	if input.TransactionID != nil && *input.TransactionID != "" {
		updates["transaction_id"] = *input.TransactionID
	}
	// 🔥 Додано підтримку PaymentTransactionID
	if input.PaymentTransactionID != nil && *input.PaymentTransactionID != "" {
		updates["payment_transaction_id"] = *input.PaymentTransactionID
	}

	if input.Date > 0 {
		updates["date"] = input.Date
	}
	if input.Value > 0 {
		updates["value"] = input.Value
	}
	if input.Diff > 0 {
		updates["diff"] = input.Diff
	}
	if input.CalculatedCost > 0 {
		updates["calculated_cost"] = input.CalculatedCost
	}

	return s.repo.UpdateReading(id, updates)
}

// 🔥 ВИПРАВЛЕНИЙ DELETE READING (ВИДАЛЯЄ ОБИДВІ ТРАНЗАКЦІЇ)
func (s *utilityService) DeleteReading(id string, user *models.User) error {
	// 1. Отримуємо показник
	reading, err := s.repo.GetReadingByID(id)
	if err != nil {
		return err
	}

	// 2. А. Видаляємо транзакцію НАРАХУВАННЯ (debt_take)
	// Це зменшить борг контрагента
	if reading.TransactionID != nil && *reading.TransactionID != "" {
		tx, err := s.txRepo.GetByID(*reading.TransactionID, user.FamilyID)
		if err == nil && tx != nil {
			_ = s.txRepo.Delete(tx, nil)
		}
	}

	// 2. Б. Видаляємо транзакцію ОПЛАТИ (debt_repay), якщо вона є
	// Це поверне гроші на баланс (скасує оплату)
	if reading.PaymentTransactionID != nil && *reading.PaymentTransactionID != "" {
		payTx, err := s.txRepo.GetByID(*reading.PaymentTransactionID, user.FamilyID)
		if err == nil && payTx != nil {
			_ = s.txRepo.Delete(payTx, nil)
		}
	}

	// 3. Видаляємо сам показник
	if err := s.repo.DeleteReading(id); err != nil {
		return err
	}

	// 4. Оновлюємо кеш лічильника
	meter, _ := s.repo.GetMeterByID(reading.MeterID, user.FamilyID)
	latestReading, _ := s.repo.GetPreviousReading(reading.MeterID, time.Now().UnixMilli())

	if latestReading != nil {
		meter.LastReadingDate = &latestReading.Date
		meter.LastReadingValue = &latestReading.Value
	} else {
		var zero float64 = 0
		meter.LastReadingDate = nil
		meter.LastReadingValue = &zero
	}

	return s.repo.UpdateMeter(meter)
}

// 🔥 ВИПРАВЛЕНИЙ PAY READING (ПИШЕ В НОВЕ ПОЛЕ)
func (s *utilityService) PayReading(readingID string, accountID string, user *models.User) error {
	reading, err := s.repo.GetReadingByID(readingID)
	if err != nil {
		return err
	}
	meter, _ := s.repo.GetMeterByID(reading.MeterID, user.FamilyID)

	// 1. Створюємо транзакцію ОПЛАТИ (debt_repay)
	payTx := &models.Transaction{
		Base:           models.Base{ID: uuid.NewString(), CreatedAt: time.Now().UnixMilli(), IsSynced: true},
		FamilyID:       user.FamilyID,
		UserID:         user.ID,
		AccountID:      accountID,
		Type:           "debt_repay", // Погашення боргу
		Amount:         reading.CalculatedCost,
		Currency:       meter.Currency,
		CounterpartyID: *meter.CounterpartyID,
		Date:           time.Now().UnixMilli(),
		Note:           "Оплата: " + meter.Name,
	}

	if err := s.txRepo.Create(payTx, nil, nil, nil); err != nil {
		return err
	}

	// 2. Оновлюємо показник
	// 🔥 Записуємо ID оплати в PaymentTransactionID
	// TransactionID (борг) залишаємо без змін!
	return s.repo.UpdateReading(readingID, map[string]interface{}{
		"is_paid":                true,
		"payment_transaction_id": payTx.ID,
		"updated_at":             time.Now().UnixMilli(),
	})
}


func (s *utilityService) GetGlobalStats(user *models.User) ([]models.UtilityStatGlobalDTO, error) {
	// Беремо статистику за останні 12 місяців
	oneYearAgo := time.Now().AddDate(-1, 0, 0).UnixMilli()

	rawRows, err := s.repo.GetGlobalStats(user.FamilyID, oneYearAgo)
	if err != nil {
		return nil, err
	}

	// 1. Групуємо raw дані в map по місяцях
	// map[ "2025-01" ] -> map[ "electricity": 500, "gas": 200 ]
	grouped := make(map[string]map[string]float64)

	// Зберігаємо порядок місяців (бо map не гарантує порядок)
	var months []string 
	seenMonths := make(map[string]bool)

	for _, row := range rawRows {
		if _, exists := grouped[row.Month]; !exists {
			grouped[row.Month] = make(map[string]float64)
		}

		if !seenMonths[row.Month] {
			months = append(months, row.Month)
			seenMonths[row.Month] = true
		}

		// Конвертуємо копійки в гривні (або основну валюту) для відображення
		grouped[row.Month][row.Type] = float64(row.Total) / 100.0
	}

	// 2. Формуємо фінальний слайс
	var result []models.UtilityStatGlobalDTO
	for _, m := range months {
		dto := models.UtilityStatGlobalDTO{
			Month: m,
			Data:  grouped[m],
		}
		result = append(result, dto)
	}

	return result, nil
}

// GET METER STATS (Local)
func (s *utilityService) GetMeterStats(meterID string, user *models.User) ([]models.UtilityStatMeterDTO, error) {
	// Також за рік
	oneYearAgo := time.Now().AddDate(-1, 0, 0).UnixMilli()

	stats, err := s.repo.GetMeterStats(meterID, user.FamilyID, oneYearAgo)
	if err != nil {
		return nil, err
	}

	// Конвертуємо копійки в гривні для зручності
	for i := range stats {
		stats[i].TotalCost = stats[i].TotalCost / 100.0
	}

	return stats, nil
}