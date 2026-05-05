package services

import (
	"log"
	"sort"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
)

type DashboardData struct {
	TotalBalance int64 `json:"total_balance"`
	TotalIncome  int64 `json:"total_income"`
	TotalExpense int64 `json:"total_expense"`
}

type StatsService interface {
	GetDashboardData(user *models.User, targetCurrency string, from, to int64, accountIDs []string) (*DashboardData, error)
	GetTopStats(user *models.User, flowType, entityType, targetCurrency string, from, to int64, accountIDs []string) ([]repositories.TopStat, error)
	GetTrendStats(user *models.User, flowType, targetCurrency string, from, to int64, accountIDs []string) ([]repositories.TrendStat, error)
	GetRecentTransactions(user *models.User, accountIDs []string) ([]models.Transaction, error)
}

type statsService struct {
	repo        repositories.StatsRepository
	currencySvc CurrencyService
}

func NewStatsService(repo repositories.StatsRepository, currSvc CurrencyService) StatsService {
	return &statsService{repo: repo, currencySvc: currSvc}
}

func (s *statsService) resolveAccess(user *models.User) (string, string) {
	familyID := user.FamilyID
	userID := ""
	if user.RoleID == "child" {
		userID = user.ID
	}
	return familyID, userID
}

// 1. Dashboard Data
func (s *statsService) GetDashboardData(user *models.User, targetCurrency string, from, to int64, accountIDs []string) (*DashboardData, error) {
	familyID, restrictUserID := s.resolveAccess(user)
	data := &DashboardData{}

	// Баланс
	accounts, err := s.repo.GetBalances(familyID, restrictUserID, accountIDs)
	if err == nil {
		var total int64
		for _, acc := range accounts {
			converted, err := s.currencySvc.Convert(acc.Balance, acc.Currency, targetCurrency)
			if err != nil {
				total += acc.Balance // Якщо помилка, плюсуємо як є (хоча це рідкісний кейс)
			} else {
				total += converted
			}
		}
		data.TotalBalance = total
	}

	// Витрати
	expStats, _ := s.repo.GetTotalSumByCurrency(familyID, restrictUserID, "expense", from, to, accountIDs)
	var totalExp int64
	for _, stat := range expStats {
		conv, _ := s.currencySvc.Convert(stat.Total, stat.Currency, targetCurrency)
		totalExp += conv
	}
	data.TotalExpense = totalExp

	// Доходи
	incStats, _ := s.repo.GetTotalSumByCurrency(familyID, restrictUserID, "income", from, to, accountIDs)
	var totalInc int64
	for _, stat := range incStats {
		conv, _ := s.currencySvc.Convert(stat.Total, stat.Currency, targetCurrency)
		totalInc += conv
	}
	data.TotalIncome = totalInc

	return data, nil
}

// 2. Top Stats
func (s *statsService) GetTopStats(user *models.User, flowType, entityType, targetCurrency string, from, to int64, accountIDs []string) ([]repositories.TopStat, error) {
	familyID, restrictUserID := s.resolveAccess(user)

	// Отримуємо "сирі" дані в різних валютах
	rawStats, err := s.repo.GetTopFlow(familyID, restrictUserID, flowType, entityType, from, to, accountIDs)
	if err != nil {
		return nil, err
	}

	// Агрегуємо по імені, конвертуючи валюту
	grouped := make(map[string]*repositories.TopStat)

	for _, item := range rawStats {
		convertedVal, _ := s.currencySvc.Convert(item.Total, item.Currency, targetCurrency)

		key := item.Name

		if _, exists := grouped[key]; !exists {
			grouped[key] = &repositories.TopStat{
				Name:     item.Name,
				Metadata: item.Metadata,
				Currency: targetCurrency, // Всі результати будуть в цільовій валюті
				Total:    0,
			}
		}
		grouped[key].Total += convertedVal
	}

	// Перетворюємо мапу в слайс
	var result []repositories.TopStat
	for _, v := range grouped {
		result = append(result, *v)
	}

	// Сортуємо (від найбільшого до найменшого)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Total > result[j].Total
	})

	if len(result) > 10 {
		return result[:10], nil
	}
	return result, nil
}

// 3. Trend Stats
func (s *statsService) GetTrendStats(user *models.User, flowType, targetCurrency string, from, to int64, accountIDs []string) ([]repositories.TrendStat, error) {
	familyID, restrictUserID := s.resolveAccess(user)

	rawStats, err := s.repo.GetTrend(familyID, restrictUserID, flowType, from, to, accountIDs)
	if err != nil {
		return nil, err
	}

	// Агрегуємо по даті (оскільки в один день могли бути операції в різних валютах)
	aggregated := make(map[string]int64)

	for _, stat := range rawStats {
		// Log для дебагу, якщо щось не сходиться
		log.Printf("[StatsService] Trend Convert: Date=%s Amount=%d From=%s To=%s", stat.Date, stat.Total, stat.Currency, targetCurrency)

		val, _ := s.currencySvc.Convert(stat.Total, stat.Currency, targetCurrency)
		aggregated[stat.Date] += val
	}

	var result []repositories.TrendStat
	for date, total := range aggregated {
		result = append(result, repositories.TrendStat{
			Date:     date,
			Currency: targetCurrency,
			Total:    total,
		})
	}

	// Сортуємо по даті
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result, nil
}

func (s *statsService) GetRecentTransactions(user *models.User, accountIDs []string) ([]models.Transaction, error) {
	familyID, restrictUserID := s.resolveAccess(user)
	return s.repo.GetRecentTransactions(familyID, restrictUserID, 10, accountIDs)
}