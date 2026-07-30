package repositories

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type InboxFilter struct {
	FamilyID string
	UserID   string
	Status   []string
	Limit    int
	Offset   int
}

type InboxRepository interface {
	Create(entry *models.InboxEntry) error
	GetAll(filter InboxFilter) ([]models.InboxEntry, int64, error)
	GetByID(id string, familyID string) (*models.InboxEntry, error)
	Update(entry *models.InboxEntry) error
}

type inboxRepo struct {
	db *gorm.DB
}

func NewInboxRepository(db *gorm.DB) InboxRepository {
	return &inboxRepo{db: db}
}

func (r *inboxRepo) Create(entry *models.InboxEntry) error {
	return r.db.Create(entry).Error
}

func (r *inboxRepo) GetAll(filter InboxFilter) ([]models.InboxEntry, int64, error) {
	var entries []models.InboxEntry
	var total int64

	statuses := filter.Status
	if len(statuses) == 0 {
		statuses = []string{
			models.InboxEntryStatusNew,
			models.InboxEntryStatusNeedsAccount,
			models.InboxEntryStatusNeedsLink,
			models.InboxEntryStatusNeedsReview,
			models.InboxEntryStatusUnlinked,
		}
	}

	query := r.db.Model(&models.InboxEntry{}).
		Preload("ReceiptSource").
		Preload("ReceiptSource.Items").
		Preload("SelectedAccount").
		Preload("MatchedTransaction").
		Where("family_id = ? AND deleted_at IS NULL", filter.FamilyID)

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	query = query.Where("status IN ?", statuses)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Order("occurred_at DESC, created_at DESC").Find(&entries).Error
	return entries, total, err
}

func (r *inboxRepo) GetByID(id string, familyID string) (*models.InboxEntry, error) {
	var entry models.InboxEntry
	err := r.db.
		Preload("ReceiptSource").
		Preload("ReceiptSource.Items").
		Preload("ReceiptSource.Counterparty").
		Preload("ReceiptSource.Category").
		Preload("SelectedAccount").
		Preload("MatchedTransaction").
		Where("id = ? AND family_id = ? AND deleted_at IS NULL", id, familyID).
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *inboxRepo) Update(entry *models.InboxEntry) error {
	return r.db.Save(entry).Error
}
