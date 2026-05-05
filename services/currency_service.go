package services

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NbuRateResponse struct {
	Rate float64 `json:"rate"`
	Cc   string  `json:"cc"`
}

type CurrencyService interface {
	SyncRates() error
	Convert(amount int64, fromCurrency, toCurrency string) (int64, error)
	GetAllRates() ([]models.ExchangeRate, error)
}

type currencyService struct {
	db        *gorm.DB
	rates     map[string]float64
	ratesLock sync.RWMutex
}

func NewCurrencyService(db *gorm.DB) CurrencyService {
	return &currencyService{
		db:    db,
		rates: make(map[string]float64),
	}
}

// 1. СИНХРОНІЗАЦІЯ (З перевіркою дати)
func (s *currencyService) SyncRates() error {
	// КРОК 1: Перевіряємо, чи є у нас ВЖЕ свіжий курс у базі?
	// Беремо USD як маркер (якщо він свіжий, то і решта теж)
	var usd models.ExchangeRate
	err := s.db.Where("currency_code = ?", "USD").First(&usd).Error

	if err == nil {
		// Перетворюємо збережений timestamp у дату
		lastUpdate := time.UnixMilli(usd.UpdatedAt)
		now := time.Now()

		// 🔥 ПЕРЕВІРКА: Якщо рік і день року збігаються — ми вже оновлювались сьогодні
		if lastUpdate.Year() == now.Year() && lastUpdate.YearDay() == now.YearDay() {
			log.Println("✅ Currency rates are up to date. Using DB cache.")
			s.loadRatesFromDB() // Важливо: завантажити в пам'ять (RAM), бо сервер міг перезапуститись
			return nil
		}
	}

	log.Println("🔄 Fetching new currency rates from NBU API...")

	// КРОК 2: Якщо даних немає або вони старі — робимо запит до НБУ
	client := http.Client{Timeout: 10 * time.Second} // Таймаут, щоб не висіло
	resp, err := client.Get("https://bank.gov.ua/NBUStatService/v1/statdirectory/exchange?json")
	if err != nil {
		log.Println("❌ NBU API failed, loading old rates from DB:", err)
		s.loadRatesFromDB()
		return err
	}
	defer resp.Body.Close()

	var nbuRates []NbuRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&nbuRates); err != nil {
		return err
	}

	// КРОК 3: Зберігаємо нові дані
	return s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		var ratesToUpsert []models.ExchangeRate
		
		s.ratesLock.Lock()
		s.rates["UAH"] = 1.0
		for _, r := range nbuRates {
			ratesToUpsert = append(ratesToUpsert, models.ExchangeRate{
				CurrencyCode: r.Cc,
				Rate:         r.Rate,
				UpdatedAt:    now,
			})
			s.rates[r.Cc] = r.Rate
		}
		s.ratesLock.Unlock()

		if len(ratesToUpsert) > 0 {
			err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "currency_code"}},
				DoUpdates: clause.AssignmentColumns([]string{"rate", "updated_at"}),
			}).Create(&ratesToUpsert).Error
			if err != nil {
				return err
			}
		}
		log.Println("✅ Currency rates updated successfully from NBU.")
		return nil
	})
}

// Helper: Завантажити курси з БД в пам'ять
func (s *currencyService) loadRatesFromDB() {
	var dbRates []models.ExchangeRate
	if err := s.db.Find(&dbRates).Error; err == nil {
		s.ratesLock.Lock()
		defer s.ratesLock.Unlock()
		s.rates["UAH"] = 1.0
		for _, r := range dbRates {
			s.rates[r.CurrencyCode] = r.Rate
		}
		// log.Println("Loaded rates from DB into RAM")
	}
}

// 2. КОНВЕРТАЦІЯ
func (s *currencyService) Convert(amount int64, fromCurrency, toCurrency string) (int64, error) {
	if fromCurrency == toCurrency {
		return amount, nil
	}

	s.ratesLock.RLock()
	rateFrom, okFrom := s.rates[fromCurrency]
	rateTo, okTo := s.rates[toCurrency]
	s.ratesLock.RUnlock()

	// Якщо кеш порожній (наприклад, сервер тільки встав і Sync ще не пройшов), пробуємо завантажити
	if !okFrom || !okTo {
		s.loadRatesFromDB()
		
		s.ratesLock.RLock()
		rateFrom = s.rates[fromCurrency]
		rateTo = s.rates[toCurrency]
		s.ratesLock.RUnlock()
	}

	// Fallback (захист від ділення на нуль)
	if rateFrom == 0 { rateFrom = 1.0 }
	
	if rateTo == 0 { 
		if toCurrency == "USD" {
			rateTo = 41.7 
		} else if toCurrency == "EUR" {
			rateTo = 44.5 
		} else {
			rateTo = 1.0 
		}
	}

	result := float64(amount) * rateFrom / rateTo
	return int64(result), nil
}

func (s *currencyService) GetAllRates() ([]models.ExchangeRate, error) {
	// Варіант А: Брати з пам'яті (швидше)
	s.ratesLock.RLock()
	defer s.ratesLock.RUnlock()

	// Якщо кеш порожній, спробуємо завантажити з БД
	if len(s.rates) == 0 {
		s.ratesLock.RUnlock() // Розблокуємо перед викликом loadRatesFromDB
		s.loadRatesFromDB()   // Ця функція сама бере Lock
		s.ratesLock.RLock()   // Знову блокуємо для читання
	}

	var result []models.ExchangeRate
	
	// Формуємо список для фронтенду
	// Можна брати з мапи s.rates, але там немає ID і UpdatedAt, 
	// тому краще дістати "красивий" список з БД, якщо треба метадані, 
	// АБО сформувати прості об'єкти з мапи.
    // Для швидкості і простоти, давай візьмемо з БД, бо там небагато записів (~30-40 валют)
	
	err := s.db.Find(&result).Error
	return result, err
}