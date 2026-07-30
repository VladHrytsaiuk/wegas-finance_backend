package models

const (
	InboxEntryStatusNew          = "new"
	InboxEntryStatusNeedsAccount = "needs_account"
	InboxEntryStatusNeedsLink    = "needs_link"
	InboxEntryStatusNeedsReview  = "needs_review"
	InboxEntryStatusLinked       = "linked"
	InboxEntryStatusUnlinked     = "unlinked"

	ReceiptSourceTypePhoto = "photo"
	ReceiptSourceTypeXML   = "xml"
	ReceiptSourceTypePDF   = "pdf"
	ReceiptSourceTypeURL   = "url"

	ReceiptOriginManualPhoto = "manual_photo"
	ReceiptOriginTelegramXML = "telegram_xml"
	ReceiptOriginTelegramPDF = "telegram_pdf"
	ReceiptOriginTelegramURL = "telegram_url"
)

// InboxEntry stores temporary receipt-driven work that is not yet attached
// to a final transaction.
type InboxEntry struct {
	Base
	FamilyID string `json:"family_id" gorm:"index"`
	UserID   string `json:"user_id" gorm:"index"`

	ReceiptSourceID string `json:"receipt_source_id" gorm:"index;not null"`

	Status string `json:"status" gorm:"index;default:'new'"`
	Reason string `json:"reason" gorm:"index"`

	SelectedAccountID    *string `json:"selected_account_id" gorm:"index"`
	MatchedTransactionID *string `json:"matched_transaction_id" gorm:"index"`

	ReviewRequired bool `json:"review_required" gorm:"default:true"`

	SourceType string `json:"source_type" gorm:"index"`
	Merchant   string `json:"merchant"`
	Total      *int64 `json:"total"`
	Currency   string `json:"currency"`
	OccurredAt *int64 `json:"occurred_at" gorm:"index"`
	Note       string `json:"note"`

	ReceiptSource      ReceiptSource `json:"receipt_source" gorm:"foreignKey:ReceiptSourceID"`
	SelectedAccount    *Account      `json:"selected_account,omitempty" gorm:"foreignKey:SelectedAccountID"`
	MatchedTransaction *Transaction  `json:"matched_transaction,omitempty" gorm:"foreignKey:MatchedTransactionID"`
}

// ReceiptSource stores the raw and normalized receipt payload regardless of
// whether it was already linked to a final transaction.
type ReceiptSource struct {
	Base
	FamilyID string `json:"family_id" gorm:"index"`
	UserID   string `json:"user_id" gorm:"index"`

	Origin     string `json:"origin" gorm:"index"`
	SourceType string `json:"source_type" gorm:"index"`

	FilePath  string `json:"file_path"`
	FilePaths string `json:"file_paths" gorm:"type:text"`
	SourceURL string `json:"source_url"`
	MimeType  string `json:"mime_type"`

	RawPayload    string `json:"raw_payload" gorm:"type:text"`
	ParsedPayload string `json:"parsed_payload" gorm:"type:text"`

	Merchant        string `json:"merchant"`
	ReceiptNumber   string `json:"receipt_number"`
	ReceiptDate     *int64 `json:"receipt_date" gorm:"index"`
	Subtotal        *int64 `json:"subtotal"`
	DiscountTotal   *int64 `json:"discount_total"`
	Total           *int64 `json:"total"`
	Currency        string `json:"currency"`
	PaymentProvider string `json:"payment_provider"`
	PaymentMask     string `json:"payment_mask"`

	CounterpartyID *string `json:"counterparty_id" gorm:"index"`
	CategoryID     *string `json:"category_id" gorm:"index"`

	LinkedTransactionID *string `json:"linked_transaction_id" gorm:"index"`
	LinkedAt            *int64  `json:"linked_at"`

	Counterparty      *Counterparty       `json:"counterparty,omitempty" gorm:"foreignKey:CounterpartyID"`
	Category          *Category           `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	LinkedTransaction *Transaction        `json:"linked_transaction,omitempty" gorm:"foreignKey:LinkedTransactionID"`
	Items             []ReceiptSourceItem `json:"items" gorm:"foreignKey:ReceiptSourceID;constraint:OnDelete:CASCADE;"`
	InboxEntries      []InboxEntry        `json:"inbox_entries,omitempty" gorm:"foreignKey:ReceiptSourceID;constraint:OnDelete:CASCADE;"`
}

// ReceiptSourceItem stores parsed line items before they are applied to a
// final transaction.
type ReceiptSourceItem struct {
	Base
	ReceiptSourceID string  `json:"receipt_source_id" gorm:"index"`
	CategoryID      *string `json:"category_id" gorm:"index"`

	Name         string `json:"name"`
	Quantity     int64  `json:"quantity"`
	PricePerUnit int64  `json:"price_per_unit"`
	TotalAmount  int64  `json:"total_amount"`

	Category Category `json:"category" gorm:"foreignKey:CategoryID"`
}
