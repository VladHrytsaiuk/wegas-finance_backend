package repositories

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Family{},
		&models.Account{},
		&models.Transaction{},
		&models.TransactionItem{},
		&models.ReceiptMerchantPreference{},
		&models.ReceiptItemCategoryPreference{},
		&models.TransactionPhoto{},
		&models.TransactionTag{},
		&models.ReceiptSource{},
		&models.ReceiptSourceItem{},
		&models.InboxEntry{},
		&models.TelegramLink{},
		&models.TelegramLinkToken{},
		&models.Tag{},
		&models.Counterparty{},
		&models.CounterpartyBalance{},
		&models.CounterpartyCategory{},
		&models.Asset{},
		&models.AssetPhoto{},
		&models.AssetDocument{},
		&models.BankAccountMapping{},
		&models.BankConnection{},
		&models.Category{},
		&models.Goal{},
		&models.StorageType{},
		&models.ShoppingList{},
		&models.ShoppingItem{},
		&models.WishlistGroup{},
		&models.WishlistItem{},
		&models.FamilyJoinCode{},
		&models.MedicalRecord{},
		&models.MedicalFile{},
		&models.UtilityMeter{},
		&models.UtilityReading{},
		&models.ExchangeRate{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
