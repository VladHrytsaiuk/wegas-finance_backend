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
