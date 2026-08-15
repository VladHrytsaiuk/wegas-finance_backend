package parsers

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/ledongthuc/pdf"
)

func parsePrivatBankPDF(reader io.Reader, size int64) ([]ParsedTransaction, error) {
	readerAt, ok := reader.(io.ReaderAt)
	if !ok {
		return nil, fmt.Errorf("reader provided to parsePrivatBankPDF must implement io.ReaderAt")
	}

	pdfReader, err := pdf.NewReader(readerAt, size)
	if err != nil {
		return nil, fmt.Errorf("failed to open pdf: %w", err)
	}

	var transactions []ParsedTransaction

	dateRegex := regexp.MustCompile(`^\s*(\d{2}\.\d{2}\.\d{4})(?:\s?(\d{2}:\d{2}))?`)
	amountRegex := regexp.MustCompile(`(-?[\d\s]+,\d{2})\s*UAH`)
	ignoreRegex := regexp.MustCompile(`(?i)^(Сторінка|Дата операції|Рахунок|Деталі операції|Сума|Залишок|Вклади гарантуються|Інформація про кредитний|№ [A-Z0-9]+|АКЦІОНЕРНЕ|Юридична адреса|Телефони|www\.pb\.ua)`)
	garbageTailRegex := regexp.MustCompile(`(?i)(УКРАЇНА,|адреса терміналу|Коментар до платежу:|Платіжне доручення|Дата валютування|Округлення залишку|код ІД НБУ).*`)

	var contentBuilder strings.Builder
	for pageIndex := 1; pageIndex <= pdfReader.NumPage(); pageIndex++ {
		page := pdfReader.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}
		text, _ := page.GetPlainText(nil)
		contentBuilder.WriteString(text)
		contentBuilder.WriteString("\n")
	}

	lines := strings.Split(contentBuilder.String(), "\n")
	var currentTx *ParsedTransaction

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if ignoreRegex.MatchString(line) {
			continue
		}

		dateMatches := dateRegex.FindStringSubmatch(line)
		if len(dateMatches) > 0 {
			if currentTx != nil {
				finalizePrivatTransaction(currentTx, garbageTailRegex)
				transactions = append(transactions, *currentTx)
			}

			dateStr := dateMatches[1]
			if len(dateMatches) > 2 && dateMatches[2] != "" {
				dateStr += " " + dateMatches[2]
			}

			parsedDate, err := utils.ParseBankDate(dateStr)
			if err != nil {
				continue
			}

			currentTx = &ParsedTransaction{
				Date:    parsedDate,
				RawLine: line,
			}

			if matches := amountRegex.FindStringSubmatch(line); len(matches) > 1 {
				if amt, err := utils.ParseAmountString(matches[1]); err == nil {
					currentTx.Amount = amt
				}
			}

			desc := dateRegex.ReplaceAllString(line, "")
			currentTx.Description = strings.TrimSpace(desc)
			continue
		}

		if currentTx != nil {
			if strings.Contains(line, "******") {
				continue
			}

			currentTx.Description += " " + line
			currentTx.RawLine += " " + line

			matches := amountRegex.FindAllStringSubmatch(line, -1)
			if len(matches) > 0 {
				lastMatch := matches[len(matches)-1]
				if amt, err := utils.ParseAmountString(lastMatch[1]); err == nil {
					currentTx.Amount = amt
				}
			}
		}
	}

	if currentTx != nil {
		finalizePrivatTransaction(currentTx, garbageTailRegex)
		transactions = append(transactions, *currentTx)
	}

	return transactions, nil
}

func finalizePrivatTransaction(tx *ParsedTransaction, cleaner *regexp.Regexp) {
	if tx.Amount < 0 {
		tx.Type = "expense"
	} else {
		tx.Type = "income"
	}

	amountInDescRegex := regexp.MustCompile(`-?[\d\s]+,\d{2}\s*UAH`)
	tx.Description = amountInDescRegex.ReplaceAllString(tx.Description, "")

	lowerDesc := strings.ToLower(tx.Description)
	if strings.Contains(lowerDesc, "скарбничк") {
		tx.Type = "transfer"
		tx.TransferDirection = "in"
		if tx.Amount < 0 {
			tx.TransferDirection = "out"
		}
		tx.CounterpartyName = "Скарбничка"

		if strings.Contains(lowerDesc, "округлення") {
			tx.Description = "Округлення залишку"
		} else if strings.Contains(lowerDesc, "відсотк") {
			tx.Description = "Відсотки на залишок"
		} else {
			tx.Description = "Поповнення"
		}
		return
	}

	rawName := ""
	if tx.Type == "income" && strings.Contains(lowerDesc, "переказ") {
		rawName = tx.Description
	} else {
		parts := strings.Split(tx.Description, ",")
		if len(parts) > 0 {
			rawName = parts[0]
		} else {
			rawName = tx.Description
		}
	}

	tx.CounterpartyName = utils.NormalizeCounterparty(rawName)

	if tx.CounterpartyName == "" {
		if len(rawName) > 50 {
			tx.CounterpartyName = "Операція"
		} else {
			tx.CounterpartyName = strings.TrimSpace(rawName)
		}
	}

	cardMaskRegex := regexp.MustCompile(`\d{2,6}[\*]+\d{2,4}`)
	tx.Description = cardMaskRegex.ReplaceAllString(tx.Description, "")

	tx.Description = cleaner.ReplaceAllString(tx.Description, "")

	tailDigitsRegex := regexp.MustCompile(`\s+-?[\d\s]+([,\.]\d{2})?$`)
	for {
		newDesc := tailDigitsRegex.ReplaceAllString(tx.Description, "")
		if newDesc == tx.Description {
			break
		}
		tx.Description = newDesc
	}

	tx.Description = strings.TrimSpace(strings.Join(strings.Fields(tx.Description), " "))
}
