package repositories

import (
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type GoalRepository struct {
	db *gorm.DB
}

func NewGoalRepository(db *gorm.DB) *GoalRepository {
	return &GoalRepository{db: db}
}

// GetDB повертає екземпляр DB (потрібен сервісу для транзакцій або специфічних оновлень)
func (r *GoalRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *GoalRepository) Create(goal *models.Goal) error {
	return r.db.Create(goal).Error
}

// FindOne шукає ціль за ID, ігноруючи видалені.
// Виправлено: враховує і 0, і NULL для deleted_at
func (r *GoalRepository) FindOne(id string) (*models.Goal, error) {
	var goal models.Goal
	err := r.db.Preload("Accounts").
		Where("id = ? AND (deleted_at = 0 OR deleted_at IS NULL)", id).
		First(&goal).Error
	
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

// FindAllByFamily повертає цілі сім'ї з урахуванням прав доступу
func (r *GoalRepository) FindAllByFamily(familyID string, userID string) ([]models.Goal, error) {
	var goals []models.Goal

	// Логіка:
	// 1. Сім'я співпадає
	// 2. Не видалено (0 або NULL)
	// 3. ПРИВАТНІСТЬ: (Це моя ціль) АБО (Вона публічна І не прихована від мене)
	err := r.db.Preload("Accounts").
		Where("family_id = ? AND (deleted_at = 0 OR deleted_at IS NULL)", familyID).
		Where("user_id = ? OR (visibility = 'public' AND (hidden_from IS NULL OR hidden_from != ?))", userID, userID).
		Order("created_at DESC").
		Find(&goals).Error

	return goals, err
}

func (r *GoalRepository) Update(goal *models.Goal) error {
	return r.db.Save(goal).Error
}

func (r *GoalRepository) Delete(id string) error {
	now := time.Now().UnixMilli()

	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Відв'язуємо рахунки
		err := tx.Model(&models.Account{}).
			Where("goal_id = ?", id).
			Update("goal_id", nil).Error
		if err != nil {
			return err
		}

		// 2. Soft Delete цілі
		err = tx.Model(&models.Goal{}).
			Where("id = ?", id).
			Update("deleted_at", now).Error
		if err != nil {
			return err
		}

		return nil
	})
}

func (r *GoalRepository) FindAllActive() ([]models.Goal, error) {
	var goals []models.Goal
	err := r.db.Preload("Accounts").
		Where("status = ? AND (deleted_at = 0 OR deleted_at IS NULL)", "active").
		Find(&goals).Error
	return goals, err
}