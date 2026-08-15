package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type DataMigration struct {
	ID string
	Up func(tx *gorm.DB) error
}

// DataMigrations is the ordered registry for transformations that AutoMigrate
// cannot safely perform by itself, such as backfilling template relationships.
var DataMigrations = []DataMigration{
	{
		ID: "20260730_link_existing_catalog_templates",
		Up: linkExistingCatalogTemplates,
	},
	{
		ID: "20260730_link_existing_catalog_templates_v2",
		Up: linkExistingCatalogTemplates,
	},
	{
		ID: "20260730_link_existing_catalog_templates_v3",
		Up: linkExistingCatalogTemplates,
	},
	{
		ID: "20260814_backfill_batch_import_account_balances",
		Up: backfillBatchImportAccountBalances,
	},
	{ID: "20260815_correct_legacy_round_up_transfers", Up: correctLegacyRoundUpTransfers},
	{ID: "20260815_purge_soft_deleted_transactions_and_accounts", Up: purgeSoftDeletedFinanceData},
}

// purgeSoftDeletedFinanceData permanently removes records the user already
// soft-deleted. It deliberately runs once and RunDataMigrations creates a
// SQLite backup before it is applied.
func purgeSoftDeletedFinanceData(tx *gorm.DB) error {
	var transactionIDs []string
	if err := tx.Model(&models.Transaction{}).Where("deleted_at IS NOT NULL").Pluck("id", &transactionIDs).Error; err != nil {
		return err
	}
	if len(transactionIDs) > 0 {
		if err := tx.Where("transaction_id IN ?", transactionIDs).Delete(&models.TransactionItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("transaction_id IN ?", transactionIDs).Delete(&models.TransactionPhoto{}).Error; err != nil {
			return err
		}
		if err := tx.Where("transaction_id IN ?", transactionIDs).Delete(&models.TransactionTag{}).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE receipt_sources SET linked_transaction_id = NULL WHERE linked_transaction_id IN ?", transactionIDs).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE inbox_entries SET matched_transaction_id = NULL WHERE matched_transaction_id IN ?", transactionIDs).Error; err != nil {
			return err
		}
		if err := tx.Where("id IN ?", transactionIDs).Delete(&models.Transaction{}).Error; err != nil {
			return err
		}
	}
	var accountIDs []string
	if err := tx.Model(&models.Account{}).Where("deleted_at IS NOT NULL").Pluck("id", &accountIDs).Error; err != nil {
		return err
	}
	if len(accountIDs) == 0 {
		return nil
	}
	if err := tx.Model(&models.Account{}).Where("round_up_target_account_id IN ?", accountIDs).Update("round_up_target_account_id", nil).Error; err != nil {
		return err
	}
	if err := tx.Where("internal_account_id IN ?", accountIDs).Delete(&models.BankAccountMapping{}).Error; err != nil {
		return err
	}
	return tx.Where("id IN ?", accountIDs).Delete(&models.Account{}).Error
}

// correctLegacyRoundUpTransfers fixes the old importer which represented an
// outgoing round-up as a one-sided `transfer` and therefore added it to the card.
func correctLegacyRoundUpTransfers(tx *gorm.DB) error {
	type row struct {
		AccountID string
		Amount    int64
	}
	var rows []row
	if err := tx.Raw(`SELECT account_id, amount FROM transactions WHERE deleted_at IS NULL AND type = 'transfer' AND transfer_related_id IS NULL AND note = 'Округлення залишку'`).Scan(&rows).Error; err != nil {
		return err
	}
	for _, item := range rows {
		// The prior migration added +amount; the correct source-card effect is -amount.
		if err := tx.Model(&models.Account{}).Where("id = ?", item.AccountID).Update("balance", gorm.Expr("balance - ?", item.Amount*2)).Error; err != nil {
			return err
		}
	}
	return nil
}

// backfillBatchImportAccountBalances repairs rows written by the legacy batch
// importer. That path inserted transactions without their account currency and
// without changing the account's cached balance.
func backfillBatchImportAccountBalances(tx *gorm.DB) error {
	type legacyTransaction struct {
		ID              string
		AccountID       string
		AccountCurrency string
		Amount          int64
		Type            string
		IsForgiveness   bool
	}

	var transactions []legacyTransaction
	if err := tx.Raw(`
		SELECT t.id, t.account_id, a.currency AS account_currency, t.amount, t.type, t.is_forgiveness
		FROM transactions t
		JOIN accounts a ON a.id = t.account_id
		WHERE t.deleted_at IS NULL AND (t.currency IS NULL OR t.currency = '')
	`).Scan(&transactions).Error; err != nil {
		return err
	}

	balanceChanges := make(map[string]int64)
	for _, transaction := range transactions {
		if err := tx.Model(&models.Transaction{}).Where("id = ?", transaction.ID).
			Update("currency", transaction.AccountCurrency).Error; err != nil {
			return err
		}
		if !transaction.IsForgiveness {
			balanceChanges[transaction.AccountID] += transactionBalanceDelta(transaction.Type, transaction.Amount)
		}
	}

	for accountID, change := range balanceChanges {
		if change == 0 {
			continue
		}
		if err := tx.Model(&models.Account{}).Where("id = ?", accountID).
			Update("balance", gorm.Expr("balance + ?", change)).Error; err != nil {
			return err
		}
	}
	return nil
}

func transactionBalanceDelta(transactionType string, amount int64) int64 {
	switch transactionType {
	case "expense", "loan_give", "debt_repay", "transfer_out":
		return -amount
	default:
		return amount
	}
}

// linkExistingCatalogTemplates backfills the relationship for families that
// existed before global catalog templates were introduced. Only exact matches
// are linked, so custom family records remain independent.
func linkExistingCatalogTemplates(tx *gorm.DB) error {
	var globalCategories []models.Category
	if err := tx.Where("family_id = ? AND deleted_at IS NULL", "").Find(&globalCategories).Error; err != nil {
		return err
	}
	byCategoryKey := make(map[string]models.Category, len(globalCategories))
	for _, item := range globalCategories {
		byCategoryKey[item.Type+"\x00"+item.Name+"\x00"+item.ParentID] = item
	}

	var localCategories []models.Category
	if err := tx.Where("family_id <> ? AND deleted_at IS NULL", "").Find(&localCategories).Error; err != nil {
		return err
	}
	localByID := make(map[string]*models.Category, len(localCategories))
	for index := range localCategories {
		localByID[localCategories[index].ID] = &localCategories[index]
	}
	// Parents must be linked before their children. A bounded pass handles the
	// two-level seeded tree and remains safe for deeper trees.
	for pass := 0; pass < len(localCategories); pass++ {
		changed := false
		for index := range localCategories {
			local := &localCategories[index]
			if local.GlobalTemplateID != nil {
				continue
			}
			globalParentID := ""
			if local.ParentID != "" {
				parent := localByID[local.ParentID]
				if parent == nil || parent.GlobalTemplateID == nil {
					continue
				}
				globalParentID = *parent.GlobalTemplateID
			}
			global, ok := byCategoryKey[local.Type+"\x00"+local.Name+"\x00"+globalParentID]
			if !ok {
				continue
			}
			if err := tx.Model(&models.Category{}).Where("id = ?", local.ID).Update("global_template_id", global.ID).Error; err != nil {
				return err
			}
			local.GlobalTemplateID = &global.ID
			changed = true
		}
		if !changed {
			break
		}
	}

	var globalCPCategories []models.CounterpartyCategory
	if err := tx.Where("family_id = ? AND deleted_at IS NULL", "").Find(&globalCPCategories).Error; err != nil {
		return err
	}
	byCPGroupKey := map[string]models.CounterpartyCategory{}
	for _, item := range globalCPCategories {
		byCPGroupKey[item.Type+"\x00"+item.Name] = item
	}
	var localCPCategories []models.CounterpartyCategory
	if err := tx.Where("family_id <> ? AND global_template_id IS NULL AND deleted_at IS NULL", "").Find(&localCPCategories).Error; err != nil {
		return err
	}
	for index := range localCPCategories {
		local := &localCPCategories[index]
		if global, ok := byCPGroupKey[local.Type+"\x00"+local.Name]; ok {
			if err := tx.Model(&models.CounterpartyCategory{}).Where("id = ?", local.ID).Update("global_template_id", global.ID).Error; err != nil {
				return err
			}
			local.GlobalTemplateID = &global.ID
		}
	}

	var globalCounterparties []models.Counterparty
	if err := tx.Where("family_id = ? AND deleted_at IS NULL", "").Find(&globalCounterparties).Error; err != nil {
		return err
	}
	byCounterpartyKey := map[string]models.Counterparty{}
	for _, item := range globalCounterparties {
		categoryID := ""
		if item.CategoryID != nil {
			categoryID = *item.CategoryID
		}
		byCounterpartyKey[item.Type+"\x00"+item.Name+"\x00"+categoryID] = item
	}
	var localCounterparties []models.Counterparty
	if err := tx.Where("family_id <> ? AND global_template_id IS NULL AND deleted_at IS NULL", "").Find(&localCounterparties).Error; err != nil {
		return err
	}
	localGroupByID := map[string]*models.CounterpartyCategory{}
	for index := range localCPCategories {
		localGroupByID[localCPCategories[index].ID] = &localCPCategories[index]
	}
	for _, local := range localCounterparties {
		globalCategoryID := ""
		if local.CategoryID != nil {
			if group := localGroupByID[*local.CategoryID]; group != nil && group.GlobalTemplateID != nil {
				globalCategoryID = *group.GlobalTemplateID
			} else {
				continue
			}
		}
		if global, ok := byCounterpartyKey[local.Type+"\x00"+local.Name+"\x00"+globalCategoryID]; ok {
			if err := tx.Model(&models.Counterparty{}).Where("id = ?", local.ID).Update("global_template_id", global.ID).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func RunDataMigrations(db *gorm.DB, dbPath string, migrations []DataMigration) error {
	if err := db.AutoMigrate(&models.SchemaMigration{}); err != nil {
		return fmt.Errorf("create schema migrations table: %w", err)
	}

	pending := make([]DataMigration, 0, len(migrations))
	for _, migration := range migrations {
		if migration.ID == "" || migration.Up == nil {
			return fmt.Errorf("invalid data migration")
		}

		var applied models.SchemaMigration
		err := db.Where("id = ?", migration.ID).First(&applied).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("check migration %s: %w", migration.ID, err)
		}
		pending = append(pending, migration)
	}

	if len(pending) == 0 {
		return nil
	}
	if err := backupSQLiteDatabase(db, dbPath); err != nil {
		return err
	}

	for _, migration := range pending {
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := migration.Up(tx); err != nil {
				return err
			}
			return tx.Create(&models.SchemaMigration{
				ID:        migration.ID,
				AppliedAt: time.Now().UnixMilli(),
			}).Error
		}); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.ID, err)
		}
	}

	return nil
}

func backupSQLiteDatabase(db *gorm.DB, dbPath string) error {
	path := strings.Split(dbPath, "?")[0]
	if path == "" || path == ":memory:" {
		return nil
	}

	directory := filepath.Join(filepath.Dir(path), "backups")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create backup directory: %w", err)
	}
	backupPath := filepath.Join(directory, fmt.Sprintf("finance-%s.db", time.Now().Format("20060102-150405")))
	quotedPath := strings.ReplaceAll(backupPath, "'", "''")
	if err := db.Exec("VACUUM INTO '" + quotedPath + "'").Error; err != nil {
		return fmt.Errorf("backup database: %w", err)
	}
	return nil
}
