package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Ловить: "від: Nataliya", "From: John", "з чорної картки", "з картки", "from black card", "від Івана"
	incomingRegex = regexp.MustCompile(`(?i)(з|від|from)\s+(.+)\s*(картки|card|рахунку)|(з картки|від|from card)\s*(.+)|від:|from:`)
	
	// Ловить: "на білу картку", "To white card", "на картку", "to card", "на рахунок"
	outgoingRegex = regexp.MustCompile(`(?i)(на|to)\s+(.+)\s*(картку|card|рахунок)|(на картку|to card)\s+(.+)`)
)

// PredictCategoryID аналізує транзакцію і намагається підібрати ID категорії
func PredictCategoryID(desc string, counterparty string, mcc string, bankCategory string, txType string, catMap map[string]string) string {
	if len(catMap) == 0 {
		return ""
	}

	searchStr := strings.ToLower(desc)
	// Контрагент приходить вже нормалізований (наприклад "АТБ", "McDonald's")
	normalizedCp := strings.ToLower(strings.TrimSpace(counterparty))

	findCat := func(names ...string) string {
		for _, name := range names {
			if id, exists := catMap[name]; exists {
				return id
			}
		}
		return ""
	}

	// --- 0. ТОЧНИЙ ЗБІГ ЗА БАНКІВСЬКОЮ КАТЕГОРІЄЮ ---
	if bankCategory != "" {
		lowerBankCat := strings.ToLower(strings.TrimSpace(bankCategory))
		exactKey := fmt.Sprintf("%s_%s", txType, lowerBankCat)
		if id := findCat(exactKey, lowerBankCat); id != "" {
			return id
		}
	}

	// --- 1. ПЕРЕКАЗИ ТА СПЕЦІАЛЬНА ЛОГІКА ---
	isTransferMCC := (mcc == "4829" || mcc == "6536")

	if txType == "income" {
		isIncomingTransfer := isTransferMCC || incomingRegex.MatchString(searchStr)
		if !isIncomingTransfer {
			transferKw := []string{"поповнення", "переказ", "зарахування", "p2p", "top-up", "replenishment"}
			for _, kw := range transferKw {
				if strings.Contains(searchStr, kw) {
					isIncomingTransfer = true
					break
				}
			}
		}

		if isIncomingTransfer {
			if id := findCat("income_перекази на картку", "income_доходи", "income_інше", "доходи", "інше"); id != "" {
				return id
			}
		}

		incomeRules := map[string]string{
			"зарплата": "зарплата", "salary": "зарплата", "виплата": "зарплата",
			"відсотки": "пасивний дохід", "кешбек": "пасивний дохід", "cashback": "пасивний дохід",
		}
		for kw, catName := range incomeRules {
			if strings.Contains(searchStr, kw) {
				if id := findCat("income_"+catName, catName); id != "" {
					return id
				}
			}
		}

	} else if txType == "expense" {
		isDonation := false
		donationKw := []string{
    "збір", "фонд", "допомога", "зсу", "благодійн", "donation", "банка", "charity",
    "омбр", "оабр", "ошбр", "тро", "ттро", "бригад", "батальйон", "дшв", "авто", "мавік", "fpv", "дрон",
}
		for _, kw := range donationKw {
			if strings.Contains(searchStr, kw) {
				isDonation = true
				break
			}
		}

		if isDonation {
			if id := findCat("expense_благодійність", "expense_фінанси та допомога", "благодійність", "фінанси та допомога"); id != "" {
				return id
			}
		}

		isOutgoingTransfer := isTransferMCC || outgoingRegex.MatchString(searchStr)
		if !isOutgoingTransfer {
			transferKw := []string{"переказ", "p2p", "transfer", "send"}
			for _, kw := range transferKw {
				if strings.Contains(searchStr, kw) {
					isOutgoingTransfer = true
					break
				}
			}
		}

		if isOutgoingTransfer {
			if id := findCat("expense_допомога рідним", "expense_перекази на картку", "expense_фінанси та допомога", "допомога рідним", "фінанси та допомога", "інше"); id != "" {
				return id
			}
		}
	}

	// --- 2. ПОШУК ПО MCC (FALLBACK) ---
	if mcc != "" {
		// Викликаємо функцію GetCategoryByMCC напряму, бо ми в пакеті utils
		if categoryName, found := GetCategoryByMCC(mcc); found {
			exactKey := fmt.Sprintf("%s_%s", txType, strings.ToLower(categoryName))
			if id := findCat(exactKey, strings.ToLower(categoryName)); id != "" {
				return id
			}
		}
	}

	// --- 3. ПОШУК ЗА НОРМАЛІЗОВАНИМ КОНТРАГЕНТОМ ---
	if normalizedCp != "" {
		counterpartyCategories := map[string]string{
			// Продукти (всі значення з utils.NormalizeCounterparty, переведені в нижній регістр)
			"атб": "продукти", "сільпо": "продукти", "фора": "продукти", "novus": "продукти",
			"varus": "продукти", "ашан": "продукти", "metro": "продукти", "roshen": "продукти",
			"сім23": "продукти", "сімі": "продукти", "близенько": "продукти", "м'ясомаркет": "продукти",
			"рукавичка": "продукти", "стовпинські ковбаси": "продукти",

			// Кафе та Ресторани
			"glovo": "кафе та ресторани", "bolt food": "кафе та ресторани", "mcdonald's": "кафе та ресторани",
			"kfc": "кафе та ресторани", "puzata hata": "кафе та ресторани", "fichepizza": "кафе та ресторани",
			"iq pizza": "кафе та ресторани", "вацак": "кафе та ресторани", "перша пекарня твого міста": "кафе та ресторани",

			// Покупки / Дім
			"eva": "покупки", "watsons": "покупки", "prostor": "покупки", "makeup": "покупки",
			"sinsay": "покупки", "readeat": "покупки", "будинок іграшок": "покупки",
			"аврора": "дім та побут", "є таке!": "дім та побут", "копійочка": "дім та побут",
			"епіцентр": "дім та побут", "нова лінія": "дім та побут", "jysk": "дім та побут",
			
			// Техніка / Маркетплейси (теж покупки)
			"rozetka": "покупки", "prom.ua": "покупки", "aliexpress": "покупки", "comfy": "покупки",
			"allo": "покупки", "foxtrot": "покупки", "citrus": "покупки", "moyo": "покупки",

			// Здоров'я
			"аптека анц": "здоров'я", "аптека подорожник": "здоров'я", "аптека 9-1-1": "здоров'я",
			"аптека доброго дня": "здоров'я", "ощад аптека": "здоров'я", "аптека оптових цін": "здоров'я",
			"аптека d.s.": "здоров'я", "аптека": "здоров'я",
			"synevo": "здоров'я", "dila": "здоров'я", "esculab": "здоров'я", "medicover": "здоров'я",

			// Транспорт / Авто
			"wog": "власне авто", "okko": "власне авто", "socar": "власне авто", "upg": "власне авто",
			"klo": "власне авто", "shell": "власне авто", "amic": "власне авто", "bvs": "власне авто",
			"укрзалізниця": "подорожі", "blablacar": "подорожі",
			"uber": "громадський транспорт", "bolt": "громадський транспорт", "uklon": "громадський транспорт",
			"easypay": "громадський транспорт", "city24": "громадський транспорт", // Або інші послуги, залежно як ти хочеш

			// Підписки / Зв'язок
			"kyivstar": "зв'язок та інтернет", "vodafone": "зв'язок та інтернет", "lifecell": "зв'язок та інтернет",
			"megogo": "зв'язок та інтернет", "netflix": "зв'язок та інтернет", "sweet tv": "зв'язок та інтернет",
			"spotify": "зв'язок та інтернет", "youtube": "зв'язок та інтернет", "google": "зв'язок та інтернет",
			"apple services": "зв'язок та інтернет", "microsoft": "зв'язок та інтернет", "openai": "зв'язок та інтернет",
			"adobe": "зв'язок та інтернет", "hostinger": "зв'язок та інтернет",
			
			// Послуги та Сервіс
			"нова пошта": "послуги та сервіс", "укрпошта": "послуги та сервіс", "meest": "послуги та сервіс",
			"дія": "послуги та сервіс", // або фінанси, залежно що це (податки тощо)

			// Розваги / Спорт
			"multiplex": "розваги", "планета кіно": "розваги",
			"sportlife": "спорт", "apollo next": "спорт",
		}

		if targetCat, exists := counterpartyCategories[normalizedCp]; exists {
			exactKey := fmt.Sprintf("%s_%s", txType, targetCat)
			if id := findCat(exactKey, targetCat); id != "" {
				return id
			}
		}
	}

	// --- 4. FALLBACK: ТЕКСТОВИЙ ПОШУК ПО ОПИСУ ---
	descriptionRules := map[string]string{
		"кава": "кафе та ресторани", "хліб": "продукти", "м'ясо": "продукти",
		"парковка": "власне авто", "миття": "власне авто", "квиток": "громадський транспорт",
		"комунал": "житло", "оренда": "житло", "осбб": "житло", "світло": "житло",
	}

	for keyword, targetCat := range descriptionRules {
		if strings.Contains(searchStr, keyword) {
			exactKey := fmt.Sprintf("%s_%s", txType, targetCat)
			if id := findCat(exactKey, targetCat); id != "" {
				return id
			}
		}
	}

	// Дефолтний fallback
	if txType == "income" {
		return findCat("income_інше", "інше")
	}
	return findCat("expense_інше", "інше")
}