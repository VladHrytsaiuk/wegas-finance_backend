package parsers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
)

type MonobankParser struct{}

func NewMonobankParser() *MonobankParser {
	return &MonobankParser{}
}

func (p *MonobankParser) Parse(reader io.Reader, size int64, fileName string) ([]ParsedTransaction, error) {
	lowerName := strings.ToLower(fileName)

	if strings.HasSuffix(lowerName, ".csv") {
		return p.parseCSV(reader)
	}
	if strings.HasSuffix(lowerName, ".xls") || strings.HasSuffix(lowerName, ".xlsx") {
		return p.parseXLS(reader)
	}
	if strings.HasSuffix(lowerName, ".pdf") {
		return p.parsePDF(reader, size)
	}
	return nil, fmt.Errorf("непідтримуваний формат файлу: %s", fileName)
}

// --- CSV ---
func (p *MonobankParser) parseCSV(r io.Reader) ([]ParsedTransaction, error) {
	csvReader := csv.NewReader(r)
	csvReader.Comma = ','
	csvReader.LazyQuotes = true

	lines, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("помилка читання CSV: %w", err)
	}

	var transactions []ParsedTransaction
	for i, row := range lines {
		// Пропускаємо заголовок або биті рядки
		if i == 0 || len(row) < 3 {
			continue
		}

		// 1. Дата
		t, err := utils.ParseBankDate(row[0])
		if err != nil {
			continue
		}

		// 2. Опис
		description := row[1]

		// 3. MCC (Зазвичай це 3-й стовпчик у Mono CSV: Date, Description, MCC, Amount...)
		var mcc string
		if len(row) > 2 {
			// Перевіряємо, чи це дійсно схоже на код (число)
			val := strings.TrimSpace(row[2])
			if len(val) == 4 { // MCC завжди 4 цифри
				mcc = val
			}
		}

		// 4. Сума
		amountIdx := 3
		if len(row) <= amountIdx {
			amountIdx = 2 // Фолбек, якщо MCC немає в файлі
		}

		amountSigned, err := utils.ParseAmountString(row[amountIdx])
		if err != nil {
			continue
		}

		txType := "income"
		if amountSigned < 0 {
			txType = "expense"
		}

		transactions = append(transactions, ParsedTransaction{
			Date:             t,
			Amount:           abs(amountSigned),
			Description:      description,
			CounterpartyName: utils.NormalizeCounterparty(description),
			Type:             txType,
			MCC:              mcc, // 🔥 Зберігаємо MCC
		})
	}
	return transactions, nil
}

// --- XLS / XLSX ---
func (p *MonobankParser) parseXLS(r io.Reader) ([]ParsedTransaction, error) {
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
		if len(row) < 3 {
			continue
		}

		t, err := utils.ParseBankDate(row[0])
		if err != nil {
			continue
		}

		description := row[1]

		// MCC в XLS (спробуємо взяти з 3-ї колонки)
		var mcc string
		if len(row) > 2 {
			val := strings.TrimSpace(row[2])
			// У Excel число може бути з плаваючою точкою типу "5411.0" або просто "5411"
			// Очистимо від зайвого
			if len(val) >= 4 {
				mcc = val[:4]
			}
		}

		amountIdx := 3
		if len(row) <= amountIdx {
			amountIdx = 2
		}

		amountSigned, err := utils.ParseAmountString(row[amountIdx])
		if err != nil {
			continue
		}

		txType := "income"
		if amountSigned < 0 {
			txType = "expense"
		}

		transactions = append(transactions, ParsedTransaction{
			Date:             t,
			Amount:           abs(amountSigned),
			Description:      description,
			CounterpartyName: utils.NormalizeCounterparty(description),
			Type:             txType,
			MCC:              mcc, // 🔥 Зберігаємо MCC
		})
	}
	return transactions, nil
}

// --- PDF ---
func (p *MonobankParser) parsePDF(r io.Reader, size int64) ([]ParsedTransaction, error) {
	readerAt, ok := r.(io.ReaderAt)
	if !ok {
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, r); err != nil {
			return nil, fmt.Errorf("failed to buffer pdf: %w", err)
		}
		readerAt = bytes.NewReader(buf.Bytes())
	}

	pdfReader, err := pdf.NewReader(readerAt, size)
	if err != nil {
		return nil, fmt.Errorf("failed to open pdf: %w", err)
	}

	var textBuilder strings.Builder
	for i := 1; i <= pdfReader.NumPage(); i++ {
		if page := pdfReader.Page(i); !page.V.IsNull() {
			txt, _ := page.GetPlainText(nil)
			textBuilder.WriteString(txt + "\n")
		}
	}

	lines := strings.Split(textBuilder.String(), "\n")
	var transactions []ParsedTransaction

	dateRegex := regexp.MustCompile(`^"?(\d{2}\.\d{2}\.\d{4})`)
	timeRegex := regexp.MustCompile(`(\d{2}:\d{2}:\d{2})`)

	// У PDF MCC коди зазвичай відсутні або їх дуже важко дістати парсингом тексту
	// Тому для PDF поле MCC залишаємо пустим

	for i := 0; i < len(lines); i++ {
		line := strings.Trim(lines[i], " \"\n\r\t")

		if dateRegex.MatchString(line) {
			matches := dateRegex.FindStringSubmatch(line)
			dateStr := matches[1]
			timeStr := "00:00:00"

			if tm := timeRegex.FindString(line); tm != "" {
				timeStr = tm
			} else if i+1 < len(lines) {
				if tm := timeRegex.FindString(lines[i+1]); tm != "" {
					timeStr = tm
					i++
				}
			}

			fullDateStr := dateStr + " " + timeStr
			t, err := utils.ParseBankDate(fullDateStr)
			if err != nil {
				continue
			}

			var desc string
			if i+1 < len(lines) {
				desc = strings.Trim(lines[i+1], " \"\n\r\t")
				i++
			}

			var amountSigned int64
			foundAmount := false

			for k := 1; k <= 4 && i+k < len(lines); k++ {
				val := strings.Trim(lines[i+k], " \"")
				if len(val) == 4 && !strings.Contains(val, ".") {
					continue
				}

				if amt, err := utils.ParseAmountString(val); err == nil {
					amountSigned = amt
					foundAmount = true
					i += k
					break
				}
			}

			if foundAmount {
				txType := "income"
				if amountSigned < 0 {
					txType = "expense"
				}

				cleanDesc := strings.TrimPrefix(desc, "Від: ")

				transactions = append(transactions, ParsedTransaction{
					Date:             t,
					Amount:           abs(amountSigned),
					Description:      desc,
					CounterpartyName: utils.NormalizeCounterparty(cleanDesc),
					Type:             txType,
					MCC:              "", // В PDF зазвичай немає MCC
				})
			}
		}
	}
	return transactions, nil
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}