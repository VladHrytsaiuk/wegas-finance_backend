package parsers

import (
	"bytes"
	"errors"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ReceiptHTMLParser interface {
	Parse(raw []byte) (*ParsedReceipt, error)
}

type silpoHTMLParser struct{}

func NewSilpoHTMLParser() ReceiptHTMLParser {
	return &silpoHTMLParser{}
}

var (
	reCenteredDiv      = regexp.MustCompile(`(?s)<div class="centered">(.*?)</div>`)
	reReceiptPageTbody = regexp.MustCompile(`(?s)<tbody>(.*?)</tbody>`)
	reTableCell        = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
	reTag              = regexp.MustCompile(`(?s)<[^>]+>`)
	reReceiptNumber    = regexp.MustCompile(`(?s)ЧЕК №\s*:</td>\s*<td[^>]*>(.*?)</td>`)
	reReceiptDate      = regexp.MustCompile(`(?s)ЧАС\s*:</td>\s*<td[^>]*>(.*?)</td>`)
	rePaymentSystem    = regexp.MustCompile(`(?s)ПЛАТІЖНА СИСТЕМА</td>\s*<td[^>]*>(.*?)</td>`)
	reCardMask         = regexp.MustCompile(`(?s)ЕПЗ</td>\s*<td[^>]*>(.*?)</td>`)
	reDiscountAmount   = regexp.MustCompile(`(?s)# В т\.ч\. знижка\s*([0-9]+(?:[.,][0-9]+)?)`)
	reAmountDueBold    = regexp.MustCompile(`(?s)ДО СПЛАТИ:</b></td>\s*<td[^>]*><b[^>]*>\s*([0-9]+(?:[.,][0-9]+)?)`)
	reAmountDuePlain   = regexp.MustCompile(`(?s)>СУМА</td>\s*<td[^>]*>\s*([0-9]+(?:[.,][0-9]+)?)`)
	reQtyTimesPrice    = regexp.MustCompile(`^\s*([0-9]+(?:[.,][0-9]+)?)\s*[Xx]\s*([0-9]+(?:[.,][0-9]+)?)\s*$`)
)

func (p *silpoHTMLParser) Parse(raw []byte) (*ParsedReceipt, error) {
	content := string(bytes.TrimSpace(raw))
	if !strings.Contains(content, `cheque-goods`) {
		return nil, errors.New("unsupported receipt html format")
	}

	receipt := &ParsedReceipt{
		Currency:   "UAH",
		SourceType: "url",
		RawSource:  content,
	}

	receipt.Merchant = parseHTMLMerchant(content)
	receipt.ReceiptNumber = extractSingleText(reReceiptNumber, content)
	receipt.ReceiptDate = parseHTMLReceiptDate(extractSingleText(reReceiptDate, content))
	receipt.Total = extractFirstMoney(content, reAmountDueBold, reAmountDuePlain)
	receipt.DiscountTotal = parseMoneyToCentsFromString(extractSingleText(reDiscountAmount, content))

	paymentProvider := extractSingleText(rePaymentSystem, content)
	cardMask := extractSingleText(reCardMask, content)
	if paymentProvider != "" || cardMask != "" {
		receipt.Payments = append(receipt.Payments, ParsedReceiptPayment{
			Type:     "Безготівкова",
			Amount:   receipt.Total,
			Provider: paymentProvider,
			Mask:     cardMask,
		})
	}

	for _, block := range reReceiptPageTbody.FindAllStringSubmatch(content, -1) {
		item, ok := parseHTMLItemBlock(block[1])
		if !ok {
			continue
		}
		receipt.Items = append(receipt.Items, item)
		receipt.Subtotal += item.TotalAmount
	}

	if receipt.Total == 0 {
		receipt.Total = receipt.Subtotal - receipt.DiscountTotal
	}

	if len(receipt.Items) == 0 && receipt.Total == 0 {
		return nil, errors.New("unsupported receipt html format")
	}

	return receipt, nil
}

func parseHTMLMerchant(content string) string {
	matches := reCenteredDiv.FindAllStringSubmatch(content, 3)
	for _, match := range matches {
		value := sanitizeHTMLText(match[1])
		if value == "" {
			continue
		}
		if strings.Contains(value, "Магазин") {
			continue
		}
		if strings.Contains(value, "ПН ") {
			continue
		}
		return value
	}
	return ""
}

func parseHTMLItemBlock(block string) (ParsedReceiptItem, bool) {
	cellMatches := reTableCell.FindAllStringSubmatch(block, -1)
	if len(cellMatches) == 0 {
		return ParsedReceiptItem{}, false
	}

	var cells []string
	for _, match := range cellMatches {
		value := sanitizeHTMLText(match[1])
		if value == "" || strings.HasPrefix(value, "ШК ") {
			continue
		}
		cells = append(cells, value)
	}

	if len(cells) < 2 {
		return ParsedReceiptItem{}, false
	}

	last := cells[len(cells)-1]
	if len(last) == 1 && isUppercaseTaxMarker(last) {
		cells = cells[:len(cells)-1]
	}
	if len(cells) < 2 {
		return ParsedReceiptItem{}, false
	}

	if len(cells) >= 3 {
		if qty, price, fractional, ok := parseQtyTimesPrice(cells[len(cells)-2]); ok {
			total := parseMoneyToCentsFromString(cells[len(cells)-1])
			name := cells[len(cells)-3]
			if fractional {
				name = name + " (" + strings.ReplaceAll(cells[len(cells)-2], " ", "") + ")"
				return ParsedReceiptItem{
					Name:         name,
					Quantity:     1,
					PricePerUnit: total,
					TotalAmount:  total,
				}, true
			}
			return ParsedReceiptItem{
				Name:         name,
				Quantity:     qty,
				PricePerUnit: price,
				TotalAmount:  total,
			}, true
		}
	}

	name := cells[len(cells)-2]
	total := parseMoneyToCentsFromString(cells[len(cells)-1])
	if total == 0 && len(cells) >= 3 {
		total = parseMoneyToCentsFromString(cells[len(cells)-2])
		name = cells[len(cells)-3]
	}

	return ParsedReceiptItem{
		Name:         name,
		Quantity:     1,
		PricePerUnit: total,
		TotalAmount:  total,
	}, total > 0
}

func sanitizeHTMLText(value string) string {
	decoded := html.UnescapeString(value)
	decoded = reTag.ReplaceAllString(decoded, "")
	decoded = strings.ReplaceAll(decoded, "\u00a0", " ")
	decoded = strings.ReplaceAll(decoded, "\n", " ")
	decoded = strings.ReplaceAll(decoded, "\t", " ")
	return strings.TrimSpace(decoded)
}

func extractSingleText(re *regexp.Regexp, content string) string {
	match := re.FindStringSubmatch(content)
	if len(match) < 2 {
		return ""
	}
	return sanitizeHTMLText(match[1])
}

func extractFirstMoney(content string, patterns ...*regexp.Regexp) int64 {
	for _, pattern := range patterns {
		value := extractSingleText(pattern, content)
		if amount := parseMoneyToCentsFromString(value); amount > 0 {
			return amount
		}
	}
	return 0
}

func parseHTMLReceiptDate(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.ParseInLocation("15:04:05 02.01.2006", value, time.Local)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func parseQtyTimesPrice(value string) (int64, int64, bool, bool) {
	match := reQtyTimesPrice.FindStringSubmatch(value)
	if len(match) != 3 {
		return 0, 0, false, false
	}
	qtyRaw, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", "."), 64)
	if err != nil {
		return 0, 0, false, false
	}
	price := parseMoneyToCentsFromString(match[2])
	if price == 0 {
		return 0, 0, false, false
	}
	if qtyRaw != float64(int64(qtyRaw)) {
		return int64(qtyRaw * 1000), price, true, true
	}
	return int64(qtyRaw), price, false, true
}

func parseMoneyToCentsFromString(value string) int64 {
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

func isUppercaseTaxMarker(value string) bool {
	return len(value) == 1 && value[0] >= 'A' && value[0] <= 'Z'
}
