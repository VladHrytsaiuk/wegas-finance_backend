package services

import (
	"errors"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InboxListFilter struct {
	Status []string
	Limit  int
	Offset int
}

type InboxCreateItemInput struct {
	Name         string  `json:"name"`
	Quantity     int64   `json:"quantity"`
	PricePerUnit int64   `json:"price_per_unit"`
	TotalAmount  int64   `json:"total_amount"`
	CategoryID   *string `json:"category_id"`
}

type InboxCreateInput struct {
	Status string
	Reason string

	SelectedAccountID *string
	ReviewRequired    *bool

	SourceType string
	Origin     string

	FilePath  string
	SourceURL string
	MimeType  string

	RawPayload    string
	ParsedPayload string

	Merchant      string
	ReceiptNumber string
	ReceiptDate   *int64
	Subtotal      *int64
	DiscountTotal *int64
	Total         *int64
	Currency      string
	OccurredAt    *int64
	Note          string

	CounterpartyID *string
	CategoryID     *string

	Items []InboxCreateItemInput
}

type InboxService interface {
	Create(input InboxCreateInput, user *models.User) (*models.InboxEntry, error)
	GetAll(filter InboxListFilter, user *models.User) ([]models.InboxEntry, int64, error)
	GetByID(id string, user *models.User) (*models.InboxEntry, error)
	SelectAccount(id string, accountID string, user *models.User) (*models.InboxEntry, error)
	Link(id string, transactionID string, applyItems bool, user *models.User) (*models.InboxEntry, error)
	Unlink(id string, user *models.User) (*models.InboxEntry, error)
}

type inboxService struct {
	repo repositories.InboxRepository
	db   *gorm.DB
}

func NewInboxService(repo repositories.InboxRepository, db *gorm.DB) InboxService {
	return &inboxService{repo: repo, db: db}
}

func (s *inboxService) Create(input InboxCreateInput, user *models.User) (*models.InboxEntry, error) {
	if input.SourceType == "" {
		return nil, errors.New("source_type is required")
	}

	status := input.Status
	if status == "" {
		status = models.InboxEntryStatusNew
	}

	reviewRequired := true
	if input.ReviewRequired != nil {
		reviewRequired = *input.ReviewRequired
	}

	now := time.Now().UnixMilli()
	receiptSourceID := uuid.NewString()
	inboxEntryID := uuid.NewString()

	receiptSource := &models.ReceiptSource{
		Base: models.Base{ID: receiptSourceID, CreatedAt: now, UpdatedAt: now, IsSynced: true},
		FamilyID:            user.FamilyID,
		UserID:              user.ID,
		Origin:              input.Origin,
		SourceType:          input.SourceType,
		FilePath:            input.FilePath,
		SourceURL:           input.SourceURL,
		MimeType:            input.MimeType,
		RawPayload:          input.RawPayload,
		ParsedPayload:       input.ParsedPayload,
		Merchant:            input.Merchant,
		ReceiptNumber:       input.ReceiptNumber,
		ReceiptDate:         input.ReceiptDate,
		Subtotal:            input.Subtotal,
		DiscountTotal:       input.DiscountTotal,
		Total:               input.Total,
		Currency:            input.Currency,
		CounterpartyID:      input.CounterpartyID,
		CategoryID:          input.CategoryID,
	}

	var sourceItems []models.ReceiptSourceItem
	for _, item := range input.Items {
		sourceItems = append(sourceItems, models.ReceiptSourceItem{
			Base:            models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
			ReceiptSourceID: receiptSourceID,
			CategoryID:      item.CategoryID,
			Name:            item.Name,
			Quantity:        item.Quantity,
			PricePerUnit:    item.PricePerUnit,
			TotalAmount:     item.TotalAmount,
		})
	}
	receiptSource.Items = sourceItems

	entry := &models.InboxEntry{
		Base:              models.Base{ID: inboxEntryID, CreatedAt: now, UpdatedAt: now, IsSynced: true},
		FamilyID:          user.FamilyID,
		UserID:            user.ID,
		ReceiptSourceID:   receiptSourceID,
		Status:            status,
		Reason:            input.Reason,
		SelectedAccountID: input.SelectedAccountID,
		ReviewRequired:    reviewRequired,
		SourceType:        input.SourceType,
		Merchant:          input.Merchant,
		Total:             input.Total,
		Currency:          input.Currency,
		OccurredAt:        input.OccurredAt,
		Note:              input.Note,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(receiptSource).Error; err != nil {
			return err
		}
		if err := tx.Create(entry).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(inboxEntryID, user.FamilyID)
}

func (s *inboxService) GetAll(filter InboxListFilter, user *models.User) ([]models.InboxEntry, int64, error) {
	repoFilter := repositories.InboxFilter{
		FamilyID: user.FamilyID,
		Status:   filter.Status,
		Limit:    filter.Limit,
		Offset:   filter.Offset,
	}

	if user.RoleID == "child" {
		repoFilter.UserID = user.ID
	}

	return s.repo.GetAll(repoFilter)
}

func (s *inboxService) GetByID(id string, user *models.User) (*models.InboxEntry, error) {
	entry, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return nil, err
	}

	if user.RoleID == "child" && entry.UserID != user.ID {
		return nil, errors.New("access denied")
	}

	return entry, nil
}

func (s *inboxService) SelectAccount(id string, accountID string, user *models.User) (*models.InboxEntry, error) {
	entry, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return nil, err
	}
	if user.RoleID == "child" && entry.UserID != user.ID {
		return nil, errors.New("access denied")
	}

	entry.SelectedAccountID = &accountID
	if entry.Status == models.InboxEntryStatusNeedsAccount || entry.Status == models.InboxEntryStatusNew {
		entry.Status = models.InboxEntryStatusNeedsLink
	}

	if err := s.repo.Update(entry); err != nil {
		return nil, err
	}

	return s.repo.GetByID(id, user.FamilyID)
}

func (s *inboxService) Link(id string, transactionID string, applyItems bool, user *models.User) (*models.InboxEntry, error) {
	entry, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return nil, err
	}
	if user.RoleID == "child" && entry.UserID != user.ID {
		return nil, errors.New("access denied")
	}

	var txModel models.Transaction
	if err := s.db.Where("id = ? AND family_id = ? AND deleted_at IS NULL", transactionID, user.FamilyID).First(&txModel).Error; err != nil {
		return nil, err
	}
	if user.RoleID == "child" && txModel.UserID != user.ID {
		return nil, errors.New("access denied")
	}

	now := time.Now().UnixMilli()
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if applyItems {
			if err := tx.Where("transaction_id = ?", txModel.ID).Delete(&models.TransactionItem{}).Error; err != nil {
				return err
			}

			for _, item := range entry.ReceiptSource.Items {
				newItem := models.TransactionItem{
					Base:          models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
					TransactionID: txModel.ID,
					CategoryID:    item.CategoryID,
					Name:          item.Name,
					Quantity:      item.Quantity,
					PricePerUnit:  item.PricePerUnit,
					TotalAmount:   item.TotalAmount,
				}
				if err := tx.Create(&newItem).Error; err != nil {
					return err
				}
			}
		}

		updates := map[string]interface{}{}
		if txModel.CounterpartyID == "" && entry.ReceiptSource.CounterpartyID != nil {
			updates["counterparty_id"] = *entry.ReceiptSource.CounterpartyID
		}
		if txModel.CategoryID == "" && entry.ReceiptSource.CategoryID != nil {
			updates["category_id"] = *entry.ReceiptSource.CategoryID
		}
		if txModel.ReceiptImg == "" && entry.ReceiptSource.FilePath != "" {
			updates["receipt_img"] = entry.ReceiptSource.FilePath
		}
		if len(updates) > 0 {
			if err := tx.Model(&models.Transaction{}).Where("id = ?", txModel.ID).Updates(updates).Error; err != nil {
				return err
			}
		}

		if entry.ReceiptSource.FilePath != "" {
			var existing int64
			if err := tx.Model(&models.TransactionPhoto{}).
				Where("transaction_id = ? AND path = ?", txModel.ID, entry.ReceiptSource.FilePath).
				Count(&existing).Error; err != nil {
				return err
			}
			if existing == 0 {
				photo := models.TransactionPhoto{
					Base:          models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
					TransactionID: txModel.ID,
					Path:          entry.ReceiptSource.FilePath,
				}
				if err := tx.Create(&photo).Error; err != nil {
					return err
				}
			}
		}

		if err := tx.Model(&models.ReceiptSource{}).Where("id = ?", entry.ReceiptSourceID).Updates(map[string]interface{}{
			"linked_transaction_id": txModel.ID,
			"linked_at":             now,
		}).Error; err != nil {
			return err
		}

		entry.MatchedTransactionID = &txModel.ID
		entry.Status = models.InboxEntryStatusLinked
		if err := tx.Save(entry).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(id, user.FamilyID)
}

func (s *inboxService) Unlink(id string, user *models.User) (*models.InboxEntry, error) {
	entry, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return nil, err
	}
	if user.RoleID == "child" && entry.UserID != user.ID {
		return nil, errors.New("access denied")
	}

	now := time.Now().UnixMilli()
	linkedTxID := ""
	if entry.MatchedTransactionID != nil {
		linkedTxID = *entry.MatchedTransactionID
	} else if entry.ReceiptSource.LinkedTransactionID != nil {
		linkedTxID = *entry.ReceiptSource.LinkedTransactionID
	}
	if linkedTxID == "" {
		return nil, errors.New("receipt is not linked")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("transaction_id = ?", linkedTxID).Delete(&models.TransactionItem{}).Error; err != nil {
			return err
		}

		if entry.ReceiptSource.FilePath != "" {
			if err := tx.Where("transaction_id = ? AND path = ?", linkedTxID, entry.ReceiptSource.FilePath).Delete(&models.TransactionPhoto{}).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.Transaction{}).
				Where("id = ? AND receipt_img = ?", linkedTxID, entry.ReceiptSource.FilePath).
				Update("receipt_img", "").Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&models.ReceiptSource{}).Where("id = ?", entry.ReceiptSourceID).Updates(map[string]interface{}{
			"linked_transaction_id": nil,
			"linked_at":             nil,
		}).Error; err != nil {
			return err
		}

		entry.MatchedTransactionID = nil
		entry.Status = models.InboxEntryStatusUnlinked
		entry.Reason = "unlinked_by_user"
		entry.UpdatedAt = now
		if err := tx.Save(entry).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(id, user.FamilyID)
}
