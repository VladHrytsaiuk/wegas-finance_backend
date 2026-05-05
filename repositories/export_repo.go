package repositories

import (
	"fmt"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type ExportRepository interface {
	GetTransactionsForExport(familyID string, filter models.ExportFilterDTO) ([]models.Transaction, error)
}

type exportRepo struct {
	db *gorm.DB
}

func NewExportRepository(db *gorm.DB) ExportRepository {
	return &exportRepo{db: db}
}

func (r *exportRepo) GetTransactionsForExport(familyID string, filter models.ExportFilterDTO) ([]models.Transaction, error) {
	var transactions []models.Transaction

	query := r.db.Model(&models.Transaction{}).
		Where("family_id = ?", familyID).
		Where("date >= ? AND date <= ?", filter.From, filter.To).
		Where("deleted_at IS NULL")

	if len(filter.AccountIDs) > 0 {
		query = query.Where("account_id IN ?", filter.AccountIDs)
	}
	if len(filter.CategoryIDs) > 0 {
		query = query.Where("category_id IN ?", filter.CategoryIDs)
	}
	
	// 🔥 Цей рядок спрацює, коли сервіс передасть сюди ID дитини
	if len(filter.UserIDs) > 0 {
		query = query.Where("user_id IN ?", filter.UserIDs)
	}
	
	if len(filter.CounterpartyIDs) > 0 {
		query = query.Where("counterparty_id IN ?", filter.CounterpartyIDs)
	}

	if len(filter.Types) > 0 {
		query = query.Where("type IN ?", filter.Types)
	}

	// Eager Loading
	query = query.
		Preload("Account").
		Preload("Category").
		Preload("Counterparty").
		Preload("Tags").
		Preload("Items").
		Preload("User").
		Order("date DESC")

	err := query.Find(&transactions).Error

	if err != nil {
		fmt.Println("❌ Export DB Error:", err)
	}

	return transactions, err
}