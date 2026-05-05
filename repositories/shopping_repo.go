package repositories

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type ShoppingRepo struct {
	db *gorm.DB
}

func NewShoppingRepo(db *gorm.DB) *ShoppingRepo {
	return &ShoppingRepo{db: db}
}

// --- РОБОТА З НОТАТКАМИ (LISTS) ---

func (r *ShoppingRepo) CreateList(list *models.ShoppingList) error {
	return r.db.Create(list).Error
}

func (r *ShoppingRepo) GetLists(familyID string, userID string) ([]models.ShoppingList, error) {
	var lists []models.ShoppingList

	// Тягнемо списки з урахуванням видимості + одразу підвантажуємо Items
	err := r.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc") // Пункти сортуємо за часом створення
	}).
		Where("family_id = ?", familyID).
		Where("visibility = 'public' OR "+
			"(visibility = 'private' AND user_id = ?) OR "+
			"(visibility = 'hidden' AND hidden_from NOT LIKE ?)",
			userID, "%"+userID+"%").
		Order("created_at desc"). // Самі нотатки сортуємо від найновіших
		Find(&lists).Error

	return lists, err
}

func (r *ShoppingRepo) UpdateList(id string, familyID string, updates map[string]interface{}) error {
	return r.db.Model(&models.ShoppingList{}).Where("id = ? AND family_id = ?", id, familyID).Updates(updates).Error
}

func (r *ShoppingRepo) DeleteList(id string, familyID string) error {
	// Завдяки OnDelete:CASCADE в моделі, GORM автоматично видалить і всі ShoppingItem для цього списку
	return r.db.Where("id = ? AND family_id = ?", id, familyID).Delete(&models.ShoppingList{}).Error
}

// --- РОБОТА З ПУНКТАМИ (ITEMS) ---

func (r *ShoppingRepo) CreateItem(item *models.ShoppingItem) error {
	return r.db.Create(item).Error
}

func (r *ShoppingRepo) UpdateItem(id string, updates map[string]interface{}) error {
	return r.db.Model(&models.ShoppingItem{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ShoppingRepo) DeleteItem(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.ShoppingItem{}).Error
}

func (r *ShoppingRepo) ClearCompletedInList(listID string) error {
	return r.db.Where("list_id = ? AND is_bought = ?", listID, true).Delete(&models.ShoppingItem{}).Error
}