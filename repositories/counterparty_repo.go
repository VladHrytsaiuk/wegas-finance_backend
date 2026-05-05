package repositories

import (
	"errors" // ✅ Додано для повернення помилки
	"fmt"
	"strings"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type CounterpartyRepository interface {
	// --- Categories ---
	CreateCategory(category *models.CounterpartyCategory) error
	GetCategories(familyID string) ([]models.CounterpartyCategory, error)
	UpdateCategory(category *models.CounterpartyCategory) error
	GetCategoryByID(id string, familyID string) (*models.CounterpartyCategory, error)

	// --- Counterparties ---
	Create(cp *models.Counterparty) error
	GetAll(familyID string) ([]models.Counterparty, error)
	Update(cp *models.Counterparty) error
	Delete(id string, familyID string) error
	GetByID(id string, familyID string) (*models.Counterparty, error)
	GetByName(name string, familyID string) (*models.Counterparty, error)
}

type cpRepo struct {
	db *gorm.DB
}

func NewCounterpartyRepository(db *gorm.DB) CounterpartyRepository {
	return &cpRepo{db: db}
}

// === CATEGORIES ===

func (r *cpRepo) CreateCategory(category *models.CounterpartyCategory) error {
	return r.db.Create(category).Error
}

func (r *cpRepo) GetCategories(familyID string) ([]models.CounterpartyCategory, error) {
	var list []models.CounterpartyCategory
	err := r.db.Where("family_id = ? AND deleted_at IS NULL", familyID).Find(&list).Error
	return list, err
}

func (r *cpRepo) GetCategoryByID(id string, familyID string) (*models.CounterpartyCategory, error) {
	var cat models.CounterpartyCategory
	err := r.db.Where("id = ? AND family_id = ?", id, familyID).First(&cat).Error
	return &cat, err
}

func (r *cpRepo) UpdateCategory(category *models.CounterpartyCategory) error {
	return r.db.Save(category).Error
}

// === COUNTERPARTIES ===

func (r *cpRepo) Create(cp *models.Counterparty) error {
	// GORM автоматично створить записи в таблиці balances,
	// якщо вони передані в структурі cp.Balances
	return r.db.Create(cp).Error
}

func (r *cpRepo) GetAll(familyID string) ([]models.Counterparty, error) {
	var list []models.Counterparty
	// 🔥 ЗМІНА: Додано Preload("Balances"), щоб завантажити мульти-валютні борги
	err := r.db.
		Preload("Category").
		Preload("Balances").
		Where("family_id = ? AND deleted_at IS NULL", familyID).
		Find(&list).Error
	return list, err
}

func (r *cpRepo) GetByID(id string, familyID string) (*models.Counterparty, error) {
	var cp models.Counterparty
	// 🔥 ЗМІНА: Додано Preload("Balances")
	err := r.db.
		Preload("Category").
		Preload("Balances").
		Where("id = ? AND family_id = ?", id, familyID).
		First(&cp).Error
	return &cp, err
}

func (r *cpRepo) GetByName(name string, familyID string) (*models.Counterparty, error) {
    // 🔍 DEBUG PRINT (Потім видалите)
    fmt.Printf("🕵️ SEARCHING DB: Name='%s' (Len: %d) | FamilyID='%s'\n", name, len(name), familyID)

    var cp models.Counterparty
    
    // Спробуємо знайти (додав TrimSpace про всяк випадок)
    cleanName := strings.TrimSpace(name)
    
    err := r.db.Where("family_id = ? AND name = ?", familyID, cleanName).First(&cp).Error

    if err != nil {
        fmt.Printf("❌ FAILED to find: %v\n", err) // Покаже, чому не знайшло
        return nil, err
    }

    fmt.Printf("✅ FOUND: ID=%s Name='%s'\n", cp.ID, cp.Name)
    return &cp, nil
}

func (r *cpRepo) Update(cp *models.Counterparty) error {
	return r.db.Save(cp).Error
}

// 🔥 ОНОВЛЕНИЙ МЕТОД DELETE
func (r *cpRepo) Delete(id string, familyID string) error {
	// 1. ПЕРЕВІРКА БАЛАНСУ
	// Перевіряємо, чи є у цього контрагента записи в balances, де сума не 0 (або > 10 копійок для похибки)
	var count int64
	err := r.db.Model(&models.CounterpartyBalance{}).
		Where("counterparty_id = ? AND ABS(balance) > 10", id).
		Count(&count).Error

	if err != nil {
		return err
	}

	// Якщо знайшли хоч один активний баланс — забороняємо видалення
	if count > 0 {
		return errors.New("cannot delete counterparty with active debt")
	}

	// 2. Якщо боргів немає — виконуємо Soft Delete
	return r.db.Model(&models.Counterparty{}).
		Where("id = ? AND family_id = ?", id, familyID).
		Update("deleted_at", time.Now().UnixMilli()).Error
}