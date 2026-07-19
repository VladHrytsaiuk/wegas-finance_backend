package parsers

import "time"

type ParsedReceipt struct {
	Merchant      string
	ReceiptNumber string
	Currency      string
	ReceiptDate   time.Time
	Subtotal      int64
	DiscountTotal int64
	Total         int64
	Items         []ParsedReceiptItem
	Payments      []ParsedReceiptPayment
	SourceType    string
	RawSource     string
}

type ParsedReceiptItem struct {
	Name         string
	Quantity     int64
	PricePerUnit int64
	TotalAmount  int64
}

type ParsedReceiptPayment struct {
	Type     string
	Amount   int64
	Provider string
	Mask     string
}
