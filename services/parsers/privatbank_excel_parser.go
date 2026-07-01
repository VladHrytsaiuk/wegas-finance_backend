package parsers

import (
	"fmt"
	"io"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/xuri/excelize/v2"
)

func parsePrivatBankXLS(r io.Reader) ([]ParsedTransaction, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("помилка відкриття Excel: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("файл не містить аркушів")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	var transactions []ParsedTransaction
	startIndex := 0

	// Шукаємо заголовок таблиці
	for i, row := range rows {
		if len(row) > 0 && (strings.Contains(row[0], "Дата") || strings.Contains(row[0], "Date")) {
			startIndex = i + 1
			break
		}
	}

	for i := startIndex; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 5 {
			continue // Рядок має містити хоча б 5 колонок (Дата, Категорія, Картка, Опис, Сума)
		}

		t, err := utils.ParseBankDate(row[0])
		if err != nil {
			continue
		}

		bankCategory := strings.TrimSpace(row[1])
		description := strings.TrimSpace(row[3])
		amountStr := row[4]

		// Очищення суми (наприклад, "-328.89 UAH" -> "-328.89")
		amountStr = strings.ReplaceAll(amountStr, " UAH", "")
		amountStr = strings.ReplaceAll(amountStr, " EUR", "")
		amountStr = strings.ReplaceAll(amountStr, " USD", "")
		
		amountSigned, err := utils.ParseAmountString(amountStr)
		if err != nil {
			continue
		}

		txType := "income"
		if amountSigned < 0 {
			txType = "expense"
		}

		rawName := ""
		lowerDesc := strings.ToLower(description)
		
		if strings.Contains(lowerDesc, "скарбничк") {
			txType = "transfer"
			rawName = "Скарбничка"
			if strings.Contains(lowerDesc, "округлення") {
				description = "Округлення залишку"
			} else if strings.Contains(lowerDesc, "відсотк") {
				description = "Відсотки на залишок"
			} else {
				description = "Поповнення"
			}
		} else if txType == "income" && strings.Contains(lowerDesc, "переказ") {
			rawName = description
		} else {
			parts := strings.Split(description, ",")
			if len(parts) > 0 {
				rawName = parts[0]
			} else {
				rawName = description
			}
		}

		counterpartyName := utils.NormalizeCounterparty(rawName)
		if counterpartyName == "" {
			if len(rawName) > 50 {
				counterpartyName = "Операція"
			} else {
				counterpartyName = strings.TrimSpace(rawName)
			}
		}

		transactions = append(transactions, ParsedTransaction{
			Date:             t,
			Amount:           abs(amountSigned), // Використовуємо існуючу функцію abs з monobank_parser.go
			Description:      description,
			CounterpartyName: counterpartyName,
			Type:             txType,
			BankCategory:     bankCategory,
			RawLine:          strings.Join(row, " "),
		})
	}
	return transactions, nil
}
