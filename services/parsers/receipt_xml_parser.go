package parsers

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
	// Embed IANA timezone data in Windows builds, where Europe/Kyiv may not be
	// available from the host OS or a Go installation.
	_ "time/tzdata"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

type ReceiptXMLParser interface {
	Parse(raw []byte) (*ParsedReceipt, error)
}

type silpoXMLParser struct{}

// Fiscal XML receipts contain a local cash-register time without an offset.
// Treat it explicitly as Ukrainian civil time, never as the server's local zone.
var ukrainianTimeLocation = func() *time.Location {
	location, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		panic("Europe/Kyiv timezone must be available: " + err.Error())
	}
	return location
}()

func NewSilpoXMLParser() ReceiptXMLParser {
	return &silpoXMLParser{}
}

type receiptXMLRQ struct {
	XMLName xml.Name      `xml:"RQ"`
	DAT     receiptXMLDAT `xml:"DAT"`
}

type receiptXMLCheck struct {
	XMLName    xml.Name             `xml:"CHECK"`
	CheckHead  receiptXMLCheckHead  `xml:"CHECKHEAD"`
	CheckTotal receiptXMLCheckTotal `xml:"CHECKTOTAL"`
	CheckPay   receiptXMLCheckPay   `xml:"CHECKPAY"`
	CheckBody  receiptXMLCheckBody  `xml:"CHECKBODY"`
}

type receiptXMLDAT struct {
	TS string      `xml:"TS"`
	C  receiptXMLC `xml:"C"`
}

type receiptXMLC struct {
	Lines    []receiptXMLLine    `xml:"L"`
	Products []receiptXMLProduct `xml:"P"`
	Payments []receiptXMLPayment `xml:"M"`
	Ends     []receiptXMLEnd     `xml:"E"`
}

type receiptXMLLine struct {
	Text string `xml:",chardata"`
}

type receiptXMLProduct struct {
	Name         string `xml:"NM,attr"`
	QuantityRaw  int64  `xml:"Q,attr"`
	PricePerUnit int64  `xml:"PRC,attr"`
	Sum          int64  `xml:"SM,attr"`
}

type receiptXMLPayment struct {
	Name          string `xml:"NM,attr"`
	PaymentSystem string `xml:"PSNM,attr"`
	MaskedCard    string `xml:"PD,attr"`
	Sum           int64  `xml:"SM,attr"`
}

type receiptXMLEnd struct {
	Number string `xml:"NO,attr"`
	Sum    int64  `xml:"SM,attr"`
	TS     string `xml:"TS,attr"`
}

type receiptXMLCheckHead struct {
	OrgName   string `xml:"ORGNM"`
	PointName string `xml:"POINTNM"`
	OrderDate string `xml:"ORDERDATE"`
	OrderTime string `xml:"ORDERTIME"`
	OrderNum  string `xml:"ORDERNUM"`
}

type receiptXMLCheckTotal struct {
	Sum         string `xml:"SUM"`
	DiscountSum string `xml:"DISCOUNTSUM"`
}

type receiptXMLCheckPay struct {
	Rows []receiptXMLCheckPayRow `xml:"ROW"`
}

type receiptXMLCheckPayRow struct {
	PayFormName string                `xml:"PAYFORMNM"`
	Sum         string                `xml:"SUM"`
	PaySys      receiptXMLCheckPaySys `xml:"PAYSYS"`
}

type receiptXMLCheckPaySys struct {
	Rows []receiptXMLCheckPaySysRow `xml:"ROW"`
}

type receiptXMLCheckPaySysRow struct {
	Name        string `xml:"NAME"`
	AcquireName string `xml:"ACQUIRENM"`
	EPZDetails  string `xml:"EPZDETAILS"`
	Sum         string `xml:"SUM"`
}

type receiptXMLCheckBody struct {
	Rows []receiptXMLCheckBodyRow `xml:"ROW"`
}

type receiptXMLCheckBodyRow struct {
	Name   string `xml:"NAME"`
	Amount string `xml:"AMOUNT"`
	Price  string `xml:"PRICE"`
	Cost   string `xml:"COST"`
}

func (p *silpoXMLParser) Parse(raw []byte) (*ParsedReceipt, error) {
	decoded, err := decodeReceiptXML(raw)
	if err != nil {
		return nil, err
	}

	trimmed := bytes.TrimSpace(decoded)
	if bytes.Contains(trimmed, []byte("<CHECK")) {
		return parseCheckXML(decoded)
	}

	var doc receiptXMLRQ
	if err := xml.Unmarshal(decoded, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse receipt xml: %w", err)
	}

	receipt := &ParsedReceipt{
		Currency:   "UAH",
		SourceType: "xml",
		RawSource:  string(decoded),
	}

	receipt.ReceiptDate = parseReceiptTimestamp(doc.DAT.TS)
	if len(doc.DAT.C.Ends) > 0 {
		receipt.ReceiptNumber = doc.DAT.C.Ends[0].Number
		if !doc.DAT.C.Ends[0].TSIsZero() {
			receipt.ReceiptDate = parseReceiptTimestamp(doc.DAT.C.Ends[0].TS)
		}
		receipt.Total = doc.DAT.C.Ends[0].Sum
	}

	for _, product := range doc.DAT.C.Products {
		item := normalizeReceiptXMLItem(product)
		receipt.Items = append(receipt.Items, item)
		receipt.Subtotal += item.TotalAmount
	}

	var paymentTotal int64
	for _, payment := range doc.DAT.C.Payments {
		parsedPayment := ParsedReceiptPayment{
			Type:     strings.TrimSpace(payment.Name),
			Amount:   payment.Sum,
			Provider: strings.TrimSpace(payment.PaymentSystem),
			Mask:     strings.TrimSpace(payment.MaskedCard),
		}
		receipt.Payments = append(receipt.Payments, parsedPayment)
		paymentTotal += payment.Sum
	}

	if receipt.Total == 0 {
		receipt.Total = paymentTotal
	}
	if receipt.Total == 0 {
		receipt.Total = receipt.Subtotal
	}

	if receipt.Subtotal > receipt.Total {
		receipt.DiscountTotal = receipt.Subtotal - receipt.Total
	}

	return receipt, nil
}

func parseCheckXML(decoded []byte) (*ParsedReceipt, error) {
	var doc receiptXMLCheck
	if err := xml.Unmarshal(decoded, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse receipt xml: %w", err)
	}

	receipt := &ParsedReceipt{
		Currency:      "UAH",
		SourceType:    "xml",
		RawSource:     string(decoded),
		Merchant:      firstNonEmpty(doc.CheckHead.OrgName, doc.CheckHead.PointName),
		ReceiptNumber: strings.TrimSpace(doc.CheckHead.OrderNum),
		ReceiptDate:   parseCheckTimestamp(doc.CheckHead.OrderDate, doc.CheckHead.OrderTime),
		Total:         parseMoneyToCents(doc.CheckTotal.Sum),
		DiscountTotal: parseMoneyToCents(doc.CheckTotal.DiscountSum),
	}

	for _, row := range doc.CheckBody.Rows {
		item := ParsedReceiptItem{
			Name:         strings.TrimSpace(row.Name),
			Quantity:     parseDecimalQuantity(row.Amount),
			PricePerUnit: parseMoneyToCents(row.Price),
			TotalAmount:  parseMoneyToCents(row.Cost),
		}
		if item.Quantity == 0 {
			item.Quantity = 1
		}
		if item.TotalAmount == 0 && item.PricePerUnit > 0 && item.Quantity > 0 {
			item.TotalAmount = item.PricePerUnit * item.Quantity
		}
		receipt.Items = append(receipt.Items, item)
		receipt.Subtotal += item.TotalAmount
	}

	for _, row := range doc.CheckPay.Rows {
		payment := ParsedReceiptPayment{
			Type:   strings.TrimSpace(row.PayFormName),
			Amount: parseMoneyToCents(row.Sum),
		}
		if len(row.PaySys.Rows) > 0 {
			payment.Provider = strings.TrimSpace(firstNonEmpty(row.PaySys.Rows[0].Name, row.PaySys.Rows[0].AcquireName))
			payment.Mask = strings.TrimSpace(row.PaySys.Rows[0].EPZDetails)
			if payment.Amount == 0 {
				payment.Amount = parseMoneyToCents(row.PaySys.Rows[0].Sum)
			}
		}
		receipt.Payments = append(receipt.Payments, payment)
	}

	if receipt.Total == 0 {
		receipt.Total = receipt.Subtotal - receipt.DiscountTotal
	}
	if receipt.Total == 0 {
		var paymentTotal int64
		for _, payment := range receipt.Payments {
			paymentTotal += payment.Amount
		}
		receipt.Total = paymentTotal
	}

	return receipt, nil
}

func decodeReceiptXML(raw []byte) ([]byte, error) {
	lower := strings.ToLower(string(raw))
	if strings.Contains(lower, `encoding="windows-1251"`) || strings.Contains(lower, `encoding='windows-1251'`) {
		if utf8.Valid(raw) {
			return normalizeXMLDeclaration(raw), nil
		}
		decoded, err := charmap.Windows1251.NewDecoder().Bytes(raw)
		if err != nil {
			return nil, err
		}
		return normalizeXMLDeclaration(decoded), nil
	}
	if strings.Contains(lower, `encoding="cp1251"`) || strings.Contains(lower, `encoding='cp1251'`) {
		if utf8.Valid(raw) {
			return normalizeXMLDeclaration(raw), nil
		}
		decoded, err := charmap.Windows1251.NewDecoder().Bytes(raw)
		if err != nil {
			return nil, err
		}
		return normalizeXMLDeclaration(decoded), nil
	}
	if bytes.HasPrefix(bytes.TrimSpace(raw), []byte("<?xml")) {
		return normalizeXMLDeclaration(raw), nil
	}
	return raw, nil
}

func normalizeXMLDeclaration(raw []byte) []byte {
	value := string(raw)
	value = strings.ReplaceAll(value, `encoding="windows-1251"`, `encoding="utf-8"`)
	value = strings.ReplaceAll(value, `encoding='windows-1251'`, `encoding="utf-8"`)
	value = strings.ReplaceAll(value, `encoding="cp1251"`, `encoding="utf-8"`)
	value = strings.ReplaceAll(value, `encoding='cp1251'`, `encoding="utf-8"`)
	return []byte(value)
}

func normalizeReceiptXMLItem(product receiptXMLProduct) ParsedReceiptItem {
	name := strings.TrimSpace(product.Name)
	quantityRaw := product.QuantityRaw
	price := product.PricePerUnit
	total := product.Sum

	if quantityRaw == 0 {
		quantityRaw = 1000
	}

	if quantityRaw%1000 == 0 {
		quantity := quantityRaw / 1000
		if quantity <= 0 {
			quantity = 1
		}
		if price == 0 && total > 0 {
			price = total / quantity
		}
		return ParsedReceiptItem{
			Name:         name,
			Quantity:     quantity,
			PricePerUnit: price,
			TotalAmount:  total,
		}
	}

	// The current transaction item model does not safely support fractional
	// quantity round-tripping through the whole app yet. For weighted goods we
	// preserve the financially correct total and price the item as a single unit.
	displayQty := float64(quantityRaw) / 1000
	if price == 0 && total > 0 {
		price = total
	}
	return ParsedReceiptItem{
		Name:         fmt.Sprintf("%s (%.3f)", name, displayQty),
		Quantity:     1,
		PricePerUnit: price,
		TotalAmount:  total,
	}
}

func parseReceiptTimestamp(value string) time.Time {
	if len(value) != 14 {
		return time.Time{}
	}
	t, err := time.ParseInLocation("20060102150405", value, ukrainianTimeLocation)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseCheckTimestamp(dateValue string, timeValue string) time.Time {
	combined := strings.TrimSpace(dateValue) + strings.TrimSpace(timeValue)
	if len(combined) != 14 {
		return time.Time{}
	}
	t, err := time.ParseInLocation("02012006150405", combined, ukrainianTimeLocation)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseMoneyToCents(value string) int64 {
	cleaned := strings.TrimSpace(strings.ReplaceAll(value, ",", "."))
	if cleaned == "" {
		return 0
	}
	floatValue, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	return int64(floatValue*100 + 0.5)
}

func parseDecimalQuantity(value string) int64 {
	cleaned := strings.TrimSpace(strings.ReplaceAll(value, ",", "."))
	if cleaned == "" {
		return 0
	}
	floatValue, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	return int64(floatValue + 0.000001)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (e receiptXMLEnd) TSIsZero() bool {
	return strings.TrimSpace(e.TS) == ""
}
