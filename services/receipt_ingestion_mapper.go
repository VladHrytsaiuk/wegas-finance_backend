package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services/parsers"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
)

func (s *receiptIngestionService) parseReceiptXMLWithMerchantFallback(raw []byte, sourceURL string) (*parsers.ParsedReceipt, error) {
	receipt, err := s.xmlParser.Parse(raw)
	if err != nil {
		return nil, err
	}
	if receipt.Merchant == "" && strings.Contains(strings.ToLower(sourceURL), "silpo") {
		receipt.Merchant = "Сільпо"
	}
	return receipt, nil
}

func (s *receiptIngestionService) createInboxFromParsedReceipt(
	receipt *parsers.ParsedReceipt,
	origin string,
	sourceURL string,
	user *models.User,
) (*models.InboxEntry, error) {
	if existing, err := s.findExistingInboxEntry(receipt, user); err == nil && existing != nil {
		return existing, nil
	}

	counterpartyID := s.findCounterpartyID(user.FamilyID, receipt.Merchant)
	categoryMap := s.loadCategoryMap(user.FamilyID)

	var items []InboxCreateItemInput
	categoryTotals := make(map[string]int64)
	for _, item := range receipt.Items {
		catID := utils.PredictCategoryID(item.Name, receipt.Merchant, "", "", "expense", categoryMap)
		var categoryID *string
		if catID != "" {
			categoryID = &catID
			categoryTotals[catID] += item.TotalAmount
		}
		items = append(items, InboxCreateItemInput{
			Name:         item.Name,
			Quantity:     item.Quantity,
			PricePerUnit: item.PricePerUnit,
			TotalAmount:  item.TotalAmount,
			CategoryID:   categoryID,
		})
	}

	var dominantCategoryID *string
	var dominantTotal int64
	for catID, total := range categoryTotals {
		if total > dominantTotal {
			id := catID
			dominantCategoryID = &id
			dominantTotal = total
		}
	}

	parsedPayloadBytes, _ := json.Marshal(receipt)
	var receiptDate *int64
	if !receipt.ReceiptDate.IsZero() {
		ts := receipt.ReceiptDate.UnixMilli()
		receiptDate = &ts
	}

	status := models.InboxEntryStatusNeedsAccount
	if len(items) == 0 {
		status = models.InboxEntryStatusNeedsReview
	}
	reviewRequired := true
	paymentProvider, paymentMask := extractPrimaryPayment(receipt)

	return s.inbox.Create(InboxCreateInput{
		Status:         status,
		Reason:         "parsed_receipt_pending_account",
		ReviewRequired: &reviewRequired,
		SourceType:     receipt.SourceType,
		Origin:         origin,
		SourceURL:      sourceURL,
		RawPayload:     receipt.RawSource,
		ParsedPayload:  string(parsedPayloadBytes),
		Merchant:       receipt.Merchant,
		ReceiptNumber:  receipt.ReceiptNumber,
		ReceiptDate:    receiptDate,
		Subtotal:       int64PtrIfNonZero(receipt.Subtotal),
		DiscountTotal:  int64PtrIfNonZero(receipt.DiscountTotal),
		Total:          int64PtrIfNonZero(receipt.Total),
		Currency:       fallbackCurrency(receipt.Currency),
		PaymentProvider: paymentProvider,
		PaymentMask:     paymentMask,
		OccurredAt:     receiptDate,
		Note:           buildReceiptNote(receipt),
		CounterpartyID: counterpartyID,
		CategoryID:     dominantCategoryID,
		Items:          items,
	}, user)
}

func (s *receiptIngestionService) findExistingInboxEntry(
	receipt *parsers.ParsedReceipt,
	user *models.User,
) (*models.InboxEntry, error) {
	rawPayload := strings.TrimSpace(receipt.RawSource)

	var entry models.InboxEntry
	query := s.db.
		Model(&models.InboxEntry{}).
		Joins("JOIN receipt_sources ON receipt_sources.id = inbox_entries.receipt_source_id").
		Preload("ReceiptSource").
		Preload("ReceiptSource.Items").
		Preload("ReceiptSource.Counterparty").
		Preload("ReceiptSource.Category").
		Preload("SelectedAccount").
		Preload("MatchedTransaction").
		Where("inbox_entries.family_id = ? AND inbox_entries.user_id = ? AND inbox_entries.deleted_at IS NULL", user.FamilyID, user.ID).
		Where("receipt_sources.deleted_at IS NULL")

	if rawPayload != "" {
		query = query.Where("receipt_sources.raw_payload = ?", rawPayload)
	} else {
		query = query.
			Where("receipt_sources.source_type = ?", receipt.SourceType).
			Where("receipt_sources.merchant = ?", receipt.Merchant).
			Where("receipt_sources.receipt_number = ?", receipt.ReceiptNumber).
			Where("receipt_sources.total = ?", receipt.Total)

		if !receipt.ReceiptDate.IsZero() {
			ts := receipt.ReceiptDate.UnixMilli()
			query = query.Where("receipt_sources.receipt_date = ?", ts)
		}
	}

	err := query.Order("inbox_entries.created_at DESC").First(&entry).Error
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (s *receiptIngestionService) loadCategoryMap(familyID string) map[string]string {
	var categories []models.Category
	mapping := make(map[string]string)
	s.db.Where("family_id = ?", familyID).Find(&categories)

	for _, cat := range categories {
		exactKey := fmt.Sprintf("%s_%s", cat.Type, strings.ToLower(cat.Name))
		mapping[exactKey] = cat.ID
		mapping[strings.ToLower(cat.Name)] = cat.ID
	}
	return mapping
}

func (s *receiptIngestionService) findCounterpartyID(familyID string, merchant string) *string {
	normalized := strings.TrimSpace(merchant)
	if normalized == "" {
		return nil
	}

	var cp models.Counterparty
	err := s.db.Where("family_id = ? AND LOWER(name) = ? AND deleted_at IS NULL", familyID, strings.ToLower(normalized)).First(&cp).Error
	if err != nil {
		return nil
	}
	return &cp.ID
}

func int64PtrIfNonZero(v int64) *int64 {
	if v == 0 {
		return nil
	}
	value := v
	return &value
}

func fallbackCurrency(currency string) string {
	if strings.TrimSpace(currency) == "" {
		return "UAH"
	}
	return currency
}

func buildReceiptNote(receipt *parsers.ParsedReceipt) string {
	var parts []string
	if receipt.Merchant != "" {
		parts = append(parts, receipt.Merchant)
	}
	if receipt.ReceiptNumber != "" {
		parts = append(parts, "receipt #"+receipt.ReceiptNumber)
	}
	if len(receipt.Payments) > 0 && receipt.Payments[0].Provider != "" {
		parts = append(parts, receipt.Payments[0].Provider)
	}
	if len(receipt.Payments) > 0 && receipt.Payments[0].Mask != "" {
		parts = append(parts, receipt.Payments[0].Mask)
	}
	return strings.Join(parts, " | ")
}

func extractPrimaryPayment(receipt *parsers.ParsedReceipt) (string, string) {
	if receipt == nil || len(receipt.Payments) == 0 {
		return "", ""
	}

	return strings.TrimSpace(receipt.Payments[0].Provider), strings.TrimSpace(receipt.Payments[0].Mask)
}
