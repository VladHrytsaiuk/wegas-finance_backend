package repositories

import (
	"fmt"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type TopStat struct {
	Name     string `json:"name"`
	Metadata string `json:"metadata"`
	Currency string `json:"currency"`
	Total    int64  `json:"total"`
}

type TrendStat struct {
	Date     string `json:"date"`
	Currency string `json:"currency"`
	Total    int64  `json:"total"`
}

type StatsRepository interface {
	GetBalances(familyID, restrictUserID string, accountIDs []string) ([]models.Account, error)
	GetTopFlow(familyID, restrictUserID, flowType, entityType string, from, to int64, accountIDs []string) ([]TopStat, error)
	GetTrend(familyID, restrictUserID, flowType string, from, to int64, accountIDs []string) ([]TrendStat, error)
	GetRecentTransactions(familyID, restrictUserID string, limit int, accountIDs []string) ([]models.Transaction, error)
	GetTotalSumByCurrency(familyID, restrictUserID, flowType string, from, to int64, accountIDs []string) ([]TopStat, error)
}

type statsRepo struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) StatsRepository {
	return &statsRepo{db: db}
}

func applyFilters(db *gorm.DB, familyID string, restrictUserID string, accountIDs []string) *gorm.DB {
	query := db.Where("transactions.family_id = ?", familyID)
	if restrictUserID != "" {
		query = query.Where("transactions.user_id = ?", restrictUserID)
	}
	if len(accountIDs) > 0 && accountIDs[0] != "" {
		query = query.Where("transactions.account_id IN ?", accountIDs)
	}
	return query
}

func (r *statsRepo) GetBalances(familyID, restrictUserID string, accountIDs []string) ([]models.Account, error) {
	var accounts []models.Account
	query := r.db.Where("family_id = ?", familyID)
	if restrictUserID != "" {
		query = query.Where("user_id = ?", restrictUserID)
	}
	if len(accountIDs) > 0 && accountIDs[0] != "" {
		query = query.Where("id IN ?", accountIDs)
	}
	err := query.Find(&accounts).Error
	return accounts, err
}

// 2. GET TOP FLOW
func (r *statsRepo) GetTopFlow(familyID, restrictUserID, flowType, entityType string, from, to int64, accountIDs []string) ([]TopStat, error) {
	var results []TopStat

	if entityType == "category" {
		whereClause := " AND t.family_id = ? AND t.deleted_at IS NULL "
		baseArgs := []interface{}{flowType, from, to, familyID}

		if restrictUserID != "" {
			whereClause += " AND t.user_id = ? "
			baseArgs = append(baseArgs, restrictUserID)
		}
		if len(accountIDs) > 0 && accountIDs[0] != "" {
			whereClause += " AND t.account_id IN ? "
			baseArgs = append(baseArgs, accountIDs)
		}

		finalArgs := append(baseArgs, baseArgs...)

		// 🔥 Тут важливо: a.currency as currency
		sql := `
      SELECT 
        COALESCE(c.name, 'Без категорії') as name, 
        COALESCE(c.color, '#94a3b8') as metadata, 
        a.currency as currency, 
        SUM(combined.amount) as total
      FROM (
        SELECT 
          COALESCE(ti.category_id, t.category_id) as final_category_id, 
          ti.total_amount as amount,
          t.account_id
        FROM transaction_items ti
        JOIN transactions t ON ti.transaction_id = t.id
        WHERE t.type = ? AND t.date >= ? AND t.date <= ? ` + whereClause + `

        UNION ALL

        SELECT 
          t.category_id as final_category_id, 
          t.amount,
          t.account_id
        FROM transactions t
        LEFT JOIN transaction_items ti ON t.id = ti.transaction_id
        WHERE ti.id IS NULL AND t.type = ? AND t.date >= ? AND t.date <= ? ` + whereClause + `
      ) as combined
      LEFT JOIN categories c ON combined.final_category_id = c.id
      JOIN accounts a ON combined.account_id = a.id
      GROUP BY c.name, c.color, a.currency
      ORDER BY total DESC
    `
		err := r.db.Raw(sql, finalArgs...).Scan(&results).Error
		return results, err
	}

	query := r.db.Table("transactions").
		Where("transactions.type = ? AND transactions.date >= ? AND transactions.date <= ? AND transactions.deleted_at IS NULL", flowType, from, to)

	query = applyFilters(query, familyID, restrictUserID, accountIDs)

	switch entityType {
	case "counterparty":
		query = query.Select("counterparties.name, counterparties.icon as metadata, accounts.currency as currency, SUM(transactions.amount) as total").
			Joins("JOIN counterparties ON transactions.counterparty_id = counterparties.id").
			Joins("JOIN accounts ON transactions.account_id = accounts.id").
			Group("counterparties.name, counterparties.icon, accounts.currency")

	case "tag":
		query = query.Select("tags.name, tags.color as metadata, accounts.currency as currency, SUM(transactions.amount) as total").
			Joins("JOIN transaction_tags ON transactions.id = transaction_tags.transaction_id").
			Joins("JOIN tags ON transaction_tags.tag_id = tags.id").
			Joins("JOIN accounts ON transactions.account_id = accounts.id").
			Group("tags.name, tags.color, accounts.currency")
	}

	err := query.Order("total DESC").Scan(&results).Error
	return results, err
}

// 3. GET TREND
func (r *statsRepo) GetTrend(familyID, restrictUserID, flowType string, from, to int64, accountIDs []string) ([]TrendStat, error) {
	var results []TrendStat

	effectiveFrom := from
	if from < 946684800000 {
		var firstTx models.Transaction
		err := r.db.Where("family_id = ? AND type = ?", familyID, flowType).
			Order("date asc").Limit(1).Find(&firstTx).Error
		if err == nil && firstTx.Date > 0 {
			effectiveFrom = firstTx.Date
		}
	}

	daysDiff := (to - effectiveFrom) / (1000 * 60 * 60 * 24)
	var dateSelector string

	if daysDiff > 366 {
		// Group by Month
		dateSelector = "strftime('%Y-%m', datetime(transactions.date/1000, 'unixepoch'))"
	} else {
		// Group by Day
		dateSelector = "strftime('%Y-%m-%d', datetime(transactions.date/1000, 'unixepoch'))"
	}

	query := r.db.Table("transactions").
		Select(fmt.Sprintf("%s as date, accounts.currency as currency, SUM(transactions.amount) as total", dateSelector)).
		Joins("JOIN accounts ON transactions.account_id = accounts.id").
		Where("transactions.type = ? AND transactions.date >= ? AND transactions.date <= ? AND transactions.deleted_at IS NULL", flowType, from, to)

	query = applyFilters(query, familyID, restrictUserID, accountIDs)

	err := query.Group("date, accounts.currency").Order("date").Scan(&results).Error
	return results, err
}

// 4. RECENT TX
func (r *statsRepo) GetRecentTransactions(familyID, restrictUserID string, limit int, accountIDs []string) ([]models.Transaction, error) {
	var transactions []models.Transaction
	query := r.db.Model(&models.Transaction{}).
		Preload("Category").
		Preload("Account").
		Preload("Counterparty").
		Preload("Tags").
		Preload("Items").
		Preload("TransferRelated"). // 🔥 Додали Preload, щоб знати про другу частину переказу
		
		// 🔥 ДОДАЛИ НОВІ ТИПИ В ФІЛЬТР
		Where("transactions.type IN ? AND transactions.deleted_at IS NULL", []string{"expense", "income", "transfer", "transfer_out", "transfer_in"}).
		Order("date DESC").
		Limit(limit)
	query = applyFilters(query, familyID, restrictUserID, accountIDs)
	err := query.Find(&transactions).Error
	return transactions, err
}

// 5. GET TOTAL SUM (Оновлено для мультивалютності)
func (r *statsRepo) GetTotalSumByCurrency(familyID, restrictUserID, flowType string, from, to int64, accountIDs []string) ([]TopStat, error) {
	var results []TopStat
	query := r.db.Table("transactions").
		Select("accounts.currency as currency, SUM(transactions.amount) as total").
		Joins("JOIN accounts ON transactions.account_id = accounts.id").
		Where("transactions.type = ? AND transactions.date >= ? AND transactions.date <= ? AND transactions.deleted_at IS NULL", flowType, from, to)
	
	query = applyFilters(query, familyID, restrictUserID, accountIDs)
	
	err := query.Group("accounts.currency").Scan(&results).Error
	return results, err
}