package services

import (
	"errors"
	"sort"
	"strings"
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

	Merchant        string
	ReceiptNumber   string
	ReceiptDate     *int64
	Subtotal        *int64
	DiscountTotal   *int64
	Total           *int64
	Currency        string
	PaymentProvider string
	PaymentMask     string
	OccurredAt      *int64
	Note            string

	CounterpartyID *string
	CategoryID     *string

	Items []InboxCreateItemInput
}

type InboxService interface {
	Create(input InboxCreateInput, user *models.User) (*models.InboxEntry, error)
	GetAll(filter InboxListFilter, user *models.User) ([]models.InboxEntry, int64, error)
	GetByID(id string, user *models.User) (*models.InboxEntry, error)
	FindAccountCandidates(id string, user *models.User) ([]InboxAccountCandidate, error)
	FindTransactionCandidates(id string, user *models.User) ([]InboxTransactionCandidate, error)
	SelectAccount(id string, accountID string, user *models.User) (*models.InboxEntry, error)
	Link(id string, transactionID string, applyItems bool, user *models.User) (*models.InboxEntry, error)
	Unlink(id string, user *models.User) (*models.InboxEntry, error)
}

type InboxAccountCandidate struct {
	AccountID         string `json:"account_id"`
	AccountName       string `json:"account_name"`
	BankName          string `json:"bank_name"`
	Currency          string `json:"currency"`
	MatchedCardNumber string `json:"matched_card_number"`
	MatchedDigits     int    `json:"matched_digits"`
	Confidence        string `json:"confidence"`
	Score             int    `json:"score"`
	Recommended       bool   `json:"recommended"`
}

type InboxTransactionCandidate struct {
	TransactionID    string   `json:"transaction_id"`
	Amount           int64    `json:"amount"`
	Currency         string   `json:"currency"`
	Date             int64    `json:"date"`
	Note             string   `json:"note"`
	CounterpartyName string   `json:"counterparty_name"`
	Score            int      `json:"score"`
	MatchedBy        []string `json:"matched_by"`
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
		Base:            models.Base{ID: receiptSourceID, CreatedAt: now, UpdatedAt: now, IsSynced: true},
		FamilyID:        user.FamilyID,
		UserID:          user.ID,
		Origin:          input.Origin,
		SourceType:      input.SourceType,
		FilePath:        input.FilePath,
		SourceURL:       input.SourceURL,
		MimeType:        input.MimeType,
		RawPayload:      input.RawPayload,
		ParsedPayload:   input.ParsedPayload,
		Merchant:        input.Merchant,
		ReceiptNumber:   input.ReceiptNumber,
		ReceiptDate:     input.ReceiptDate,
		Subtotal:        input.Subtotal,
		DiscountTotal:   input.DiscountTotal,
		Total:           input.Total,
		Currency:        input.Currency,
		PaymentProvider: input.PaymentProvider,
		PaymentMask:     input.PaymentMask,
		CounterpartyID:  input.CounterpartyID,
		CategoryID:      input.CategoryID,
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

func paymentMaskSuffix(mask string) string {
	mask = strings.TrimSpace(mask)
	end := len(mask)
	start := end
	for start > 0 {
		char := mask[start-1]
		if char < '0' || char > '9' {
			break
		}
		start--
	}

	suffix := mask[start:end]
	if len(suffix) > 4 {
		return suffix[len(suffix)-4:]
	}
	return suffix
}

func accountCardNumbers(account models.Account) []string {
	seen := make(map[string]struct{}, len(account.CardNumbers)+1)
	numbers := make([]string, 0, len(account.CardNumbers)+1)
	for _, number := range append([]string{account.CardNumber}, account.CardNumbers...) {
		number = strings.TrimSpace(number)
		if len(number) != 4 {
			continue
		}
		if _, exists := seen[number]; exists {
			continue
		}
		seen[number] = struct{}{}
		numbers = append(numbers, number)
	}
	return numbers
}

func findAccountCandidates(mask, currency string, accounts []models.Account) []InboxAccountCandidate {
	suffix := paymentMaskSuffix(mask)
	if suffix == "" {
		return []InboxAccountCandidate{}
	}

	candidates := make([]InboxAccountCandidate, 0)
	for _, account := range accounts {
		if account.IsArchived || account.IsGroup || account.Type != "card" {
			continue
		}

		for _, cardNumber := range accountCardNumbers(account) {
			if !strings.HasSuffix(cardNumber, suffix) {
				continue
			}
			matchedDigits := len(suffix)
			score := 0
			if matchedDigits == 4 {
				score = 60
				if currency != "" && account.Currency == currency {
					score += 10
				}
			}
			candidates = append(candidates, InboxAccountCandidate{
				AccountID:         account.ID,
				AccountName:       account.Name,
				BankName:          account.BankName,
				Currency:          account.Currency,
				MatchedCardNumber: cardNumber,
				MatchedDigits:     matchedDigits,
				Confidence:        map[bool]string{true: "exact", false: "partial"}[matchedDigits == 4],
				Score:             score,
			})
			break
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].MatchedDigits != candidates[j].MatchedDigits {
			return candidates[i].MatchedDigits > candidates[j].MatchedDigits
		}
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].AccountName < candidates[j].AccountName
	})

	if len(candidates) == 1 && candidates[0].Score >= 60 {
		candidates[0].Recommended = true
	}
	return candidates
}

func (s *inboxService) FindAccountCandidates(id string, user *models.User) ([]InboxAccountCandidate, error) {
	entry, err := s.GetByID(id, user)
	if err != nil {
		return nil, err
	}

	var accounts []models.Account
	query := s.db.Where("family_id = ? AND deleted_at IS NULL", user.FamilyID)
	if user.RoleID == "child" {
		query = query.Where("user_id = ?", user.ID)
	}
	if err := query.Find(&accounts).Error; err != nil {
		return nil, err
	}

	return findAccountCandidates(entry.ReceiptSource.PaymentMask, entry.Currency, accounts), nil
}

func absoluteDifference(first, second int64) int64 {
	if first >= second {
		return first - second
	}
	return second - first
}

func receiptMerchantMatchesTransaction(source models.ReceiptSource, transaction models.Transaction) bool {
	merchant := strings.ToLower(strings.TrimSpace(source.Merchant))
	if merchant == "" {
		return false
	}

	for _, value := range []string{transaction.Note, transaction.Counterparty.Name} {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" && (strings.Contains(value, merchant) || strings.Contains(merchant, value)) {
			return true
		}
	}
	return false
}

func scoreTransactionCandidate(entry *models.InboxEntry, transaction models.Transaction) InboxTransactionCandidate {
	score := 40 // The user explicitly confirmed this account for the receipt.
	matchedBy := []string{"рахунок"}

	total := entry.Total
	if total == nil {
		total = entry.ReceiptSource.Total
	}
	if total != nil && transaction.Amount == *total {
		score += 25
		matchedBy = append(matchedBy, "сума")
	}
	if entry.Currency != "" && entry.Currency == transaction.Currency {
		score += 10
		matchedBy = append(matchedBy, "валюта")
	}

	receiptDate := entry.OccurredAt
	if receiptDate == nil {
		receiptDate = entry.ReceiptSource.ReceiptDate
	}
	if receiptDate != nil {
		diff := absoluteDifference(*receiptDate, transaction.Date)
		switch {
		case diff <= int64(2*time.Hour/time.Millisecond):
			score += 10
			matchedBy = append(matchedBy, "час")
		case diff <= int64(24*time.Hour/time.Millisecond):
			score += 6
			matchedBy = append(matchedBy, "дата")
		case diff <= int64(72*time.Hour/time.Millisecond):
			score += 2
			matchedBy = append(matchedBy, "дата")
		}
	}
	if receiptMerchantMatchesTransaction(entry.ReceiptSource, transaction) {
		score += 5
		matchedBy = append(matchedBy, "контрагент")
	}

	return InboxTransactionCandidate{
		TransactionID:    transaction.ID,
		Amount:           transaction.Amount,
		Currency:         transaction.Currency,
		Date:             transaction.Date,
		Note:             transaction.Note,
		CounterpartyName: transaction.Counterparty.Name,
		Score:            score,
		MatchedBy:        matchedBy,
	}
}

func (s *inboxService) FindTransactionCandidates(id string, user *models.User) ([]InboxTransactionCandidate, error) {
	entry, err := s.GetByID(id, user)
	if err != nil {
		return nil, err
	}
	if entry.SelectedAccountID == nil || *entry.SelectedAccountID == "" {
		return []InboxTransactionCandidate{}, nil
	}

	total := entry.Total
	if total == nil {
		total = entry.ReceiptSource.Total
	}
	if total == nil {
		return []InboxTransactionCandidate{}, nil
	}

	query := s.db.
		Preload("Counterparty").
		Where("family_id = ? AND account_id = ? AND amount = ? AND external_id <> '' AND deleted_at IS NULL", user.FamilyID, *entry.SelectedAccountID, *total).
		Where("NOT EXISTS (SELECT 1 FROM receipt_sources WHERE receipt_sources.linked_transaction_id = transactions.id AND receipt_sources.deleted_at IS NULL)")
	if user.RoleID == "child" {
		query = query.Where("user_id = ?", user.ID)
	}

	receiptDate := entry.OccurredAt
	if receiptDate == nil {
		receiptDate = entry.ReceiptSource.ReceiptDate
	}
	if receiptDate != nil {
		window := int64(72 * time.Hour / time.Millisecond)
		query = query.Where("date BETWEEN ? AND ?", *receiptDate-window, *receiptDate+window)
	}

	var transactions []models.Transaction
	if err := query.Order("date DESC").Limit(10).Find(&transactions).Error; err != nil {
		return nil, err
	}

	candidates := make([]InboxTransactionCandidate, 0, len(transactions))
	for _, transaction := range transactions {
		candidates = append(candidates, scoreTransactionCandidate(entry, transaction))
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].Date > candidates[j].Date
	})
	return candidates, nil
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

func transactionItemsFromReceipt(source models.ReceiptSource, transactionID string, now int64) []models.TransactionItem {
	items := make([]models.TransactionItem, 0, len(source.Items)+2)
	itemsTotal := int64(0)
	for _, item := range source.Items {
		itemsTotal += item.TotalAmount
		items = append(items, models.TransactionItem{
			Base:          models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
			TransactionID: transactionID,
			CategoryID:    item.CategoryID,
			Name:          item.Name,
			Quantity:      item.Quantity,
			PricePerUnit:  item.PricePerUnit,
			TotalAmount:   item.TotalAmount,
		})
	}

	if source.DiscountTotal == nil || *source.DiscountTotal == 0 {
		return items
	}
	if source.Total != nil && absoluteDifference(itemsTotal, *source.Total) <= 1 {
		// Some retailers already distribute the discount across item prices.
		return items
	}

	// Keep the itemized total equal to the receipt and the bank transaction.
	if len(items) == 0 && source.Subtotal != nil {
		items = append(items, models.TransactionItem{
			Base:          models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
			TransactionID: transactionID,
			Name:          "Проміжна сума за чеком",
			Quantity:      1,
			PricePerUnit:  *source.Subtotal,
			TotalAmount:   *source.Subtotal,
		})
	}

	discount := *source.DiscountTotal
	if discount < 0 {
		discount = -discount
	}
	items = append(items, models.TransactionItem{
		Base:          models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
		TransactionID: transactionID,
		Name:          "Знижка за чеком",
		Quantity:      1,
		PricePerUnit:  -discount,
		TotalAmount:   -discount,
	})

	return items
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
	if entry.SelectedAccountID != nil && *entry.SelectedAccountID != "" && txModel.AccountID != *entry.SelectedAccountID {
		return nil, errors.New("transaction account does not match selected receipt account")
	}

	receiptTotal := entry.Total
	if receiptTotal == nil {
		receiptTotal = entry.ReceiptSource.Total
	}
	if receiptTotal != nil && txModel.Amount != *receiptTotal {
		return nil, errors.New("transaction amount does not match receipt total")
	}

	now := time.Now().UnixMilli()
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if applyItems {
			if err := tx.Where("transaction_id = ?", txModel.ID).Delete(&models.TransactionItem{}).Error; err != nil {
				return err
			}

			for _, newItem := range transactionItemsFromReceipt(entry.ReceiptSource, txModel.ID, now) {
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
