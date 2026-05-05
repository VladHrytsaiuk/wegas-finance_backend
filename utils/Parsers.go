package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseAmountString парсить суму, зберігаючи знак мінуса.
// Підтримує формати: "-1 234,56", "1234.56-", "1.234,56"
func ParseAmountString(amountStr string) (int64, error) {
	// 1. Нормалізація знаку мінус (всі види тире -> звичайний мінус)
	amountStr = strings.ReplaceAll(amountStr, "–", "-") // En dash
	amountStr = strings.ReplaceAll(amountStr, "—", "-") // Em dash
	amountStr = strings.ReplaceAll(amountStr, "−", "-") // Minus sign
	amountStr = strings.ReplaceAll(amountStr, " ", "")  // Видаляємо пробіли

	// 2. Перевірка на мінус (на початку або в кінці)
	isNegative := strings.Contains(amountStr, "-")
	
	// 3. Чистка від всього, крім цифр, крапок і ком
	re := regexp.MustCompile(`[^\d\.,]`)
	cleanStr := re.ReplaceAllString(amountStr, "")

	if cleanStr == "" {
		return 0, fmt.Errorf("empty amount string")
	}

	// 4. Визначаємо десятковий роздільник
	// Якщо є кома, замінюємо її на крапку, а всі існуючі крапки (роздільники тисяч) видаляємо
	if strings.Contains(cleanStr, ",") {
		// Видаляємо крапки (це роздільники тисяч: 1.250,00 -> 1250,00)
		cleanStr = strings.ReplaceAll(cleanStr, ".", "")
		// Міняємо кому на крапку (1250,00 -> 1250.00)
		cleanStr = strings.ReplaceAll(cleanStr, ",", ".")
	} 
	// Якщо коми немає, вважаємо, що крапка - це десятковий роздільник (формат 1250.50)

	// 5. Парсинг числа
	val, err := strconv.ParseFloat(cleanStr, 64)
	if err != nil {
		return 0, err
	}

	// 6. Відновлюємо знак мінус
	if isNegative {
		val = -val
	}

	// 7. Конвертація в копійки
	return int64(val * 100), nil
}

// ParseBankDate - без змін, він працює
func ParseBankDate(dateStr string) (time.Time, error) {
	re := regexp.MustCompile(`[^\d\.\-\/\:\s]`)
	cleanStr := re.ReplaceAllString(dateStr, "")
	cleanStr = strings.TrimSpace(cleanStr)

	if cleanStr == "" {
		return time.Time{}, fmt.Errorf("порожній рядок дати")
	}

	loc, err := time.LoadLocation("Europe/Kiev")
	if err != nil {
		loc = time.UTC
	}

	formats := []string{
		"02.01.2006 15:04:05",
		"02.01.2006 15:04",
		"02.01.2006",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02/01/2006",
		"2006/01/02",
		"02.01.06 15:04",
		"02.01.06",
	}

	for _, layout := range formats {
		t, err := time.ParseInLocation(layout, cleanStr, loc)
		if err == nil {
			if t.Year() < 2000 {
				continue
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("невідомий формат дати: %s", cleanStr)
}