package repositories

import (
	"errors"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type UtilityRepository interface {
	// Meters
	CreateMeter(meter *models.UtilityMeter) error
	GetMeters(familyID string) ([]models.UtilityMeter, error)
	GetMeterByID(id string, familyID string) (*models.UtilityMeter, error)
	UpdateMeter(meter *models.UtilityMeter) error
	DeleteMeter(id string, familyID string) error

	// Readings
	CreateReading(reading *models.UtilityReading) error
	GetReadings(familyID string, meterID string) ([]models.UtilityReading, error)

	GetPreviousReading(meterID string, date int64) (*models.UtilityReading, error)
	GetNextReading(meterID string, date int64) (*models.UtilityReading, error)

	UpdateReading(id string, updates map[string]interface{}) error
	DeleteReading(id string) error
	GetReadingByID(id string) (*models.UtilityReading, error)

	GetGlobalStats(familyID string, from int64) ([]models.UtilityStatRaw, error)
  GetMeterStats(meterID string, familyID string, from int64) ([]models.UtilityStatMeterDTO, error)
}

// 🔥 Використовуємо єдину назву структури для всього файлу
type utilityRepo struct {
	db *gorm.DB
}

func NewUtilityRepository(db *gorm.DB) UtilityRepository {
	return &utilityRepo{db: db}
}

// --- METERS ---

func (r *utilityRepo) CreateMeter(meter *models.UtilityMeter) error {
	return r.db.Create(meter).Error
}

func (r *utilityRepo) GetMeters(familyID string) ([]models.UtilityMeter, error) {
	var meters []models.UtilityMeter
	// 🔥 Preload Counterparty.Balances дозволяє фронтенду бачити сумарний борг
	err := r.db.Preload("Asset").
		Preload("Counterparty.Balances").
		Where("family_id = ?", familyID).
		Find(&meters).Error
	return meters, err
}

func (r *utilityRepo) GetMeterByID(id string, familyID string) (*models.UtilityMeter, error) {
	var meter models.UtilityMeter
	// 🔥 Важливо завантажити зв'язки, інакше сервіс не зможе створити транзакцію з CounterpartyID
	err := r.db.Preload("Asset").
		Preload("Counterparty.Balances").
		Where("id = ? AND family_id = ?", id, familyID).
		First(&meter).Error
	return &meter, err
}

func (r *utilityRepo) UpdateMeter(meter *models.UtilityMeter) error {
	return r.db.Save(meter).Error
}

func (r *utilityRepo) DeleteMeter(id string, familyID string) error {
	return r.db.Where("id = ? AND family_id = ?", id, familyID).Delete(&models.UtilityMeter{}).Error
}

// --- READINGS ---

func (r *utilityRepo) CreateReading(reading *models.UtilityReading) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Створюємо показник
		if err := tx.Create(reading).Error; err != nil {
			return err
		}

		// 2. Оновлюємо "останні показники" в лічильнику для швидкого доступу
		var latest models.UtilityReading
		if err := tx.Where("meter_id = ?", reading.MeterID).Order("date desc, created_at desc").First(&latest).Error; err == nil {
			if reading.Date >= latest.Date {
				return tx.Model(&models.UtilityMeter{}).
					Where("id = ?", reading.MeterID).
					Updates(map[string]interface{}{
						"last_reading_date":  latest.Date,
						"last_reading_value": latest.Value,
					}).Error
			}
		}
		return nil
	})
}

func (r *utilityRepo) GetReadings(familyID string, meterID string) ([]models.UtilityReading, error) {
	var reads []models.UtilityReading
	// Join для перевірки доступу через сім'ю
	db := r.db.Joins("JOIN utility_meters ON utility_meters.id = utility_readings.meter_id").
		Where("utility_meters.family_id = ?", familyID)

	if meterID != "" {
		db = db.Where("utility_readings.meter_id = ?", meterID)
	}

	// 🔥 Уточнюємо назву таблиці для created_at, щоб уникнути конфлікту імен після JOIN
	err := db.Order("utility_readings.date desc, utility_readings.created_at desc").Find(&reads).Error
	return reads, err
}

func (r *utilityRepo) GetPreviousReading(meterID string, date int64) (*models.UtilityReading, error) {
	var reading models.UtilityReading
	// <= дозволяє коректно працювати з кількома записами в один день
	err := r.db.Where("meter_id = ? AND date <= ?", meterID, date).
		Order("date desc, created_at desc").
		Limit(1).
		First(&reading).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &reading, err
}

func (r *utilityRepo) GetNextReading(meterID string, date int64) (*models.UtilityReading, error) {
	var reading models.UtilityReading
	err := r.db.Where("meter_id = ? AND date > ?", meterID, date).
		Order("date asc, created_at asc").
		Limit(1).
		First(&reading).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &reading, err
}

// backend/repositories/utility_repo.go

func (r *utilityRepo) UpdateReading(id string, updates map[string]interface{}) error {
    // 🔥 Використовуємо Updates, щоб змінити ТІЛЬКИ ті поля, що прийшли (напр. is_paid)
    // Решта полів (date, value) залишаться незмінними в базі
    return r.db.Model(&models.UtilityReading{}).Where("id = ?", id).Updates(updates).Error
}

func (r *utilityRepo) DeleteReading(id string) error {
	return r.db.Delete(&models.UtilityReading{}, "id = ?", id).Error
}

func (r *utilityRepo) GetReadingByID(id string) (*models.UtilityReading, error) {
	var reading models.UtilityReading
	// Preload Meter потрібен для доступу до тарифу та валюти
	err := r.db.Preload("Meter").Where("id = ?", id).First(&reading).Error
	if err != nil {
		return nil, err
	}
	return &reading, err
}


// Глобальна статистика: Групуємо по Місяцю та Типу послуги
func (r *utilityRepo) GetGlobalStats(familyID string, from int64) ([]models.UtilityStatRaw, error) {
	var results []models.UtilityStatRaw

	// SQL для SQLite. 
	// date/1000 — переводимо мілісекунди в секунди
	// 'unixepoch' — кажемо SQLite, що це timestamp
	query := `
		SELECT 
			strftime('%Y-%m', datetime(r.date/1000, 'unixepoch')) as month,
			m.type as type, 
			SUM(r.calculated_cost) as total
		FROM utility_readings r
		JOIN utility_meters m ON r.meter_id = m.id
		WHERE m.family_id = ? AND r.date >= ?
		GROUP BY month, m.type
		ORDER BY month ASC
	`

	err := r.db.Raw(query, familyID, from).Scan(&results).Error
	return results, err
}

// Локальна статистика: Групуємо по Місяцю для конкретного лічильника
func (r *utilityRepo) GetMeterStats(meterID string, familyID string, from int64) ([]models.UtilityStatMeterDTO, error) {
	var results []models.UtilityStatMeterDTO

	query := `
		SELECT 
			strftime('%Y-%m', datetime(r.date/1000, 'unixepoch')) as month,
			SUM(r.diff) as total_consumption,
			SUM(r.calculated_cost) as total_cost,
			AVG(r.tariff_at_date) as avg_tariff
		FROM utility_readings r
		JOIN utility_meters m ON r.meter_id = m.id
		WHERE r.meter_id = ? AND m.family_id = ? AND r.date >= ?
		GROUP BY month
		ORDER BY month ASC
	`

	err := r.db.Raw(query, meterID, familyID, from).Scan(&results).Error
	return results, err
}