package parsers

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/ledongthuc/pdf"
)

type PrivatBankParser struct{}

func NewPrivatBankParser() *PrivatBankParser {
	return &PrivatBankParser{}
}

func (p *PrivatBankParser) Parse(reader io.Reader, size int64, filename string) ([]ParsedTransaction, error) {
	// PDF бібліотека вимагає ReaderAt.
	// ImportService передає *bytes.Reader, тому ми робимо приведення типів.
	readerAt, ok := reader.(io.ReaderAt)
	if !ok {
		return nil, fmt.Errorf("reader provided to PrivatBankParser must implement io.ReaderAt")
	}

	pdfReader, err := pdf.NewReader(readerAt, size)
	if err != nil {
		return nil, fmt.Errorf("failed to open pdf: %w", err)
	}

	var transactions []ParsedTransaction

	// 🔥 РЕГУЛЯРКИ
	// 1. Дата: ^\s* дозволяє пробіли на початку. (?:\s?(\d{2}:\d{2}))? робить час необов'язковим.
	dateRegex := regexp.MustCompile(`^\s*(\d{2}\.\d{2}\.\d{4})(?:\s?(\d{2}:\d{2}))?`)
	
	// 2. Сума: шукаємо числа з комою перед словом UAH
	amountRegex := regexp.MustCompile(`(-?[\d\s]+,\d{2})\s*UAH`)
	
	// 3. Ігнорування службових рядків
	ignoreRegex := regexp.MustCompile(`(?i)^(Сторінка|Дата операції|Рахунок|Деталі операції|Сума|Залишок|Вклади гарантуються|Інформація про кредитний|№ [A-Z0-9]+|АКЦІОНЕРНЕ|Юридична адреса|Телефони|www\.pb\.ua)`)
	
	// 4. Чистка хвостів опису (адреси терміналів і т.д.)
	garbageTailRegex := regexp.MustCompile(`(?i)(УКРАЇНА,|адреса терміналу|Коментар до платежу:|Платіжне доручення|Дата валютування|Округлення залишку|код ІД НБУ).*`)

	// Читаємо весь текст PDF у StringBuilder
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

		// === ПОЧАТОК НОВОЇ ТРАНЗАКЦІЇ (Знайдена дата) ===
		dateMatches := dateRegex.FindStringSubmatch(line)
		if len(dateMatches) > 0 {
			// Якщо ми вже збирали попередню транзакцію, зберігаємо її
			if currentTx != nil {
				finalizeTransaction(currentTx, garbageTailRegex)
				transactions = append(transactions, *currentTx)
			}

			// Формуємо рядок дати для парсера
			dateStr := dateMatches[1]
			if len(dateMatches) > 2 && dateMatches[2] != "" {
				dateStr += " " + dateMatches[2]
			}

			// 🔥 Парсимо дату через твій utils
			parsedDate, err := utils.ParseBankDate(dateStr)
			if err != nil {
				// Якщо дата невалідна (наприклад це просто цифри в тексті), пропускаємо
				continue
			}

			currentTx = &ParsedTransaction{
				Date:    parsedDate,
				RawLine: line,
			}

			// Пробуємо знайти суму в цьому ж рядку
			if matches := amountRegex.FindStringSubmatch(line); len(matches) > 1 {
				if amt, err := utils.ParseAmountString(matches[1]); err == nil {
					currentTx.Amount = amt
				}
			}

			// Опис - це все, що залишилось після видалення дати
			desc := dateRegex.ReplaceAllString(line, "")
			currentTx.Description = strings.TrimSpace(desc)
			continue
		}

		// === ПРОДОВЖЕННЯ ПОТОЧНОЇ ТРАНЗАКЦІЇ ===
		if currentTx != nil {
			// Пропускаємо маскування карток, якщо воно йде окремим рядком
			if strings.Contains(line, "******") {
				continue
			}
			
			currentTx.Description += " " + line
			currentTx.RawLine += " " + line

			// Шукаємо суму (часто вона на другому рядку, якщо опис довгий)
			matches := amountRegex.FindAllStringSubmatch(line, -1)
			if len(matches) > 0 {
				lastMatch := matches[len(matches)-1]
				if amt, err := utils.ParseAmountString(lastMatch[1]); err == nil {
					currentTx.Amount = amt
				}
			}
		}
	}

	// Зберігаємо останню транзакцію після виходу з циклу
	if currentTx != nil {
		finalizeTransaction(currentTx, garbageTailRegex)
		transactions = append(transactions, *currentTx)
	}

	return transactions, nil
}

// finalizeTransaction - фінальна обробка та очистка даних
func finalizeTransaction(tx *ParsedTransaction, cleaner *regexp.Regexp) {
	// 1. Визначаємо тип
	if tx.Amount < 0 {
		tx.Type = "expense"
	} else {
		tx.Type = "income"
	}

	// 2. Видаляємо суму з тексту опису ("-100.00 UAH")
	amountInDescRegex := regexp.MustCompile(`-?[\d\s]+,\d{2}\s*UAH`)
	tx.Description = amountInDescRegex.ReplaceAllString(tx.Description, "")

	// 3. Обробка "Скарбнички" (це внутрішній переказ)
	lowerDesc := strings.ToLower(tx.Description)
	if strings.Contains(lowerDesc, "скарбничк") {
		tx.Type = "transfer"
		tx.CounterpartyName = "Скарбничка"
		
		if strings.Contains(lowerDesc, "округлення") {
			tx.Description = "Округлення залишку"
		} else if strings.Contains(lowerDesc, "відсотк") {
			tx.Description = "Відсотки на залишок"
		} else {
			tx.Description = "Поповнення"
		}
		// Для скарбнички далі нормалізацію робити не треба
		return 
	}

	// 4. Виділення "Сирої назви" для пошуку бренду
	rawName := ""
	if tx.Type == "income" && strings.Contains(lowerDesc, "переказ") {
		rawName = tx.Description // Для P2P переказів беремо весь текст
	} else {
		// Зазвичай назва магазину йде до коми
		parts := strings.Split(tx.Description, ",")
		if len(parts) > 0 {
			rawName = parts[0]
		} else {
			rawName = tx.Description
		}
	}

	// 5. Нормалізація контрагента через твій utils
	tx.CounterpartyName = utils.NormalizeCounterparty(rawName)
	
	// Фолбек, якщо назва не нормалізувалась
	if tx.CounterpartyName == "" {
		if len(rawName) > 50 {
			tx.CounterpartyName = "Операція"
		} else {
			tx.CounterpartyName = strings.TrimSpace(rawName)
		}
	}

	// 6. Фінальна зачистка опису
	// Прибираємо маски карток типу "5168******9747"
	cardMaskRegex := regexp.MustCompile(`\d{2,6}[\*]+\d{2,4}`)
	tx.Description = cardMaskRegex.ReplaceAllString(tx.Description, "")

	// Використовуємо cleaner (адреси терміналів і т.д.)
	tx.Description = cleaner.ReplaceAllString(tx.Description, "")
	
	// Чистка цифрових хвостів в кінці опису ("ATB ... -9")
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