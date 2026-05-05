package repositories

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type WishlistRepo struct {
	db *gorm.DB
}

func NewWishlistRepo(db *gorm.DB) *WishlistRepo {
	return &WishlistRepo{db: db}
}

// --- GROUPS ---

func (r *WishlistRepo) CreateGroup(group *models.WishlistGroup) error {
	return r.db.Create(group).Error
}

func (r *WishlistRepo) GetGroups(familyID, myUserID string) ([]models.WishlistGroup, error) {
	var groups []models.WishlistGroup
	err := r.db.Where("family_id = ?", familyID).
		Where("user_id = ? OR (visibility != 'private' AND hidden_from != ?)", myUserID, myUserID).
		Order("created_at desc").Find(&groups).Error
	return groups, err
}

func (r *WishlistRepo) UpdateGroup(id string, familyID string, updates map[string]interface{}) error {
	return r.db.Model(&models.WishlistGroup{}).
		Where("id = ? AND family_id = ?", id, familyID).
		Updates(updates).Error
}

func (r *WishlistRepo) DeleteGroup(id string, familyID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Спочатку відв'язуємо всі айтеми від цієї групи (GroupID = NULL)
		if err := tx.Model(&models.WishlistItem{}).
			Where("group_id = ? AND family_id = ?", id, familyID).
			Update("group_id", nil).Error; err != nil {
			return err
		}

		// 2. Видаляємо саму групу
		if err := tx.Where("id = ? AND family_id = ?", id, familyID).Delete(&models.WishlistGroup{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// --- ITEMS ---

func (r *WishlistRepo) Create(item *models.WishlistItem) error {
	return r.db.Create(item).Error
}

// GetAll тепер підтримує фільтрацію по Групі та Користувачу
func (r *WishlistRepo) GetAll(familyID, myUserID, groupID, targetUserID string) ([]models.WishlistItem, error) {
	var items []models.WishlistItem

	// Початковий запит: тільки сім'я
	query := r.db.Where("family_id = ?", familyID)

	// Фільтр приватності: (Мої) АБО (Публічні І не сховані від мене)
	query = query.Where("user_id = ? OR (visibility = 'public' AND hidden_from != ?)", myUserID, myUserID)

	// 1. Фільтр по ГРУПІ
	if groupID != "" {
		if groupID == "null" {
			// Отримати елементи БЕЗ групи ("Загальні")
			query = query.Where("group_id IS NULL")
		} else {
			// Отримати елементи конкретної папки
			query = query.Where("group_id = ?", groupID)
		}
	}

	// 2. Фільтр по ЮЗЕРУ (якщо клікнули на аватарку дружини)
	if targetUserID != "" {
		query = query.Where("user_id = ?", targetUserID)
	}

	// Завантажуємо зв'язки і сортуємо
	err := query.
		Preload("Goal").
		Preload("Group"). // Підтягуємо інфо про групу
		Order("priority desc, created_at desc").
		Find(&items).Error

	return items, err
}

func (r *WishlistRepo) GetByID(id string, familyID string) (*models.WishlistItem, error) {
	var item models.WishlistItem
	err := r.db.
		Where("id = ? AND family_id = ?", id, familyID).
		Preload("Goal").
		Preload("Group").
		First(&item).Error
	return &item, err
}

func (r *WishlistRepo) Update(id string, familyID string, updates map[string]interface{}) error {
	return r.db.Model(&models.WishlistItem{}).
		Where("id = ? AND family_id = ?", id, familyID).
		Updates(updates).Error
}

func (r *WishlistRepo) Delete(id string, familyID string) error {
	return r.db.Where("id = ? AND family_id = ?", id, familyID).Delete(&models.WishlistItem{}).Error
}
func (r *WishlistRepo) GetDB() *gorm.DB {
	return r.db
}