package repositories

import (
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type AccountRepository interface {
	Create(account *models.Account) error
	GetAllByFamilyID(familyID string) ([]models.Account, error)
	GetAllByUserID(userID string) ([]models.Account, error)
	GetByID(id string) (*models.Account, error)
	Update(account *models.Account) error
	Delete(id string) error
	GetByExternalID(externalID string) (*models.Account, error)
}

type accountRepo struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepo{db: db}
}

func (r *accountRepo) Create(account *models.Account) error {
	return r.db.Create(account).Error
}

// Повертає всі рахунки сім'ї (для батьків)
func (r *accountRepo) GetAllByFamilyID(familyID string) ([]models.Account, error) {
	var accounts []models.Account
	// Додаємо Preload("StorageType"), щоб фронт знав іконки та типи
	err := r.db.Preload("StorageType").
		Where("family_id = ? AND deleted_at IS NULL", familyID).
		Find(&accounts).Error
	return accounts, err
}

// Повертає тільки власні рахунки (для дітей)
func (r *accountRepo) GetAllByUserID(userID string) ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.Preload("StorageType").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Find(&accounts).Error
	return accounts, err
}

func (r *accountRepo) GetByID(id string) (*models.Account, error) {
	var account models.Account
	// Важливо підвантажити StorageType, якщо захочеш бачити деталі у формі редагування
	err := r.db.Preload("StorageType").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepo) Update(account *models.Account) error {
	// 🔥 ВИПРАВЛЕНО: Додано повний перелік полів для Select,
	// щоб GORM міг занулити GoalID або StorageTypeID при зміні типу.
	return r.db.Model(account).
		Select(
			"Name",
			"Type",
			"Currency",
			"Balance",
			"InitialBalance",
			"Color",
			"BankName",
			"CardType",
			"CardNumber",
			"CardNumbers",
			"PaymentSystem",
			"UserID",
			"StorageTypeID",
			"GoalID",
			"UpdatedAt",
			"IsArchived",
			"IsSynced", // Додано
			"IsGroup",  // Додано
		).
		Updates(account).Error
}

func (r *accountRepo) Delete(id string) error {
	return r.db.Model(&models.Account{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now().UnixMilli()).Error
}

// Знадобиться для MonobankService
func (r *accountRepo) GetByExternalID(externalID string) (*models.Account, error) {
	var account models.Account
	err := r.db.Preload("StorageType").
		Where("external_id = ? AND deleted_at IS NULL", externalID).
		First(&account).Error
	return &account, err
}
