package utils

import (
	"regexp"
	"strings"
)

	func NormalizeCounterparty(rawName string) string {
		cleanName := strings.TrimSpace(rawName)
		upperName := strings.ToUpper(cleanName)

		// 1. СЛОВНИК БРЕНДІВ (Пріоритетний пошук)
		// Ключ (UPPER CASE) -> Офіційна назва в БД (як у SeedDefaultCounterparties)
		rules := map[string]string{

// === СУПЕРМАРКЕТИ ===
		"ATB": "АТБ",
		"АТБ": "АТБ",
		"SILPO": "Сільпо",
		"СIЛЬПО": "Сільпо",
		"CILPO": "Сільпо",
		"FORA": "Фора",
		"ФОРА": "Фора",
		"NOVUS": "Novus",
		"НОВУС": "Novus",
		"VARUS": "Varus",
		"ВАРУС": "Varus",
		"ASHAN": "Ашан",
		"АШАН": "Ашан",
		"AUCHAN": "Ашан",
		"METRO": "Metro",
		"МЕТРО": "Metro",
		"ROSHEN": "Roshen",
		"РОШЕН": "Roshen",
		"SIM23": "Сім23",
		"SIM 23": "Сім23",
		"СІМ23": "Сім23",
		"Сім-23": "Сім23",
		"SIMI": "Сімі",
		"СІМІ": "Сімі",
		"BLYZENKO": "Близенько",
		"БЛИЗЕНЬКО": "Близенько",
		"MYASOMARKET": "М'ясомаркет",
		"MIRA MEAT": "М'ясомаркет",
		"М'ЯСОМАРКЕТ": "М'ясомаркет",
		"RUKAVYCHKA": "Рукавичка",
		"РУКАВИЧКА": "Рукавичка",
		"STOVPYNSKI": "Стовпинські ковбаси",
			
		// === ДРОГЕРІ ТА АПТЕКИ ===
		"EVA": "Eva",
		"ЄВА": "Eva",
		"WATSONS": "Watsons",
		"ВАТСОНС": "Watsons",
		"PROSTOR": "Prostor",
		"ПРОСТОР": "Prostor",
		"AVRORA": "Аврора",
		"АВРОРА": "Аврора",
		"YE TAKE": "Є Таке!",
		"Є ТАКЕ": "Є Таке!",
		"E-TAKE": "Є Таке!",
		"KOPIYOCHKA": "Копійочка",
		"КОПІЙОЧКА": "Копійочка",
		"SINSAY": "Sinsay",
		"MAKEUP": "Makeup",
		"МЕЙКАП": "Makeup",
		"EPITSENTR": "Епіцентр",
		"EPICENTR": "Епіцентр",
		"ЕПІЦЕНТР": "Епіцентр",
		"NOVA LINIYA": "Нова Лінія",
		"НОВА ЛІНІЯ": "Нова Лінія",
		"JYSK": "Jysk",
		"ЮСК": "Jysk",
		"BUDYNOK IGRASHOK": "Будинок іграшок",
		"БУДИНОК ІГРАШОК": "Будинок іграшок",
		"Бужинок Іграшок": "Будинок іграшок",
		"Readeat": "Readeat",

		// Аптеки
		"ANC": "Аптека АНЦ",
		"АНЦ": "Аптека АНЦ",
		"Аптека низьких цін": "Аптека АНЦ",
		"PODOROZHNYK": "Аптека Подорожник",
		"ПОДОРОЖНИК": "Аптека Подорожник",
		"Аптека Подорожник": "Аптека Подорожник",
		"9-1-1": "Аптека 9-1-1",
		"911": "Аптека 9-1-1",
		"DOBROHO DNYA": "Аптека Доброго Дня",
		"ДОБРОГО ДНЯ": "Аптека Доброго Дня",
		"ADD": "Аптека Доброго Дня",
		"Аптека доброго дня": "Аптека Доброго Дня",
		"OSHCHAD APTEKA": "Ощад Аптека",
		"ОЩАД АПТЕКА": "Ощад Аптека",
		"Ощад аптека": "Ощад Аптека",
		"APTEKA OPTOVYKH": "Аптека оптових цін",
		"PHARMACY": "Аптека",
		"APTEKA": "Аптека",

		// === АЗС ===
		"WOG": "WOG",
		"ВОГ": "WOG",
		"OKKO": "OKKO",
		"ОККО": "OKKO",
		"SOCAR": "SOCAR",
		"СОКАР": "SOCAR",
		"UPG": "UPG",
		"УПГ": "UPG",
		"KLO": "KLO",
		"КЛО": "KLO",
		"SHELL": "Shell",
		"ШЕЛЛ": "Shell",
		"AMIC": "Amic",
		"BVS": "BVS",

		// === ТРАНСПОРТ ===
		"UZ GOV": "Укрзалізниця",
		"UKRZALIZNYTSIA": "Укрзалізниця",
		"УКРЗАЛІЗНИЦЯ": "Укрзалізниця",
		"UZ.GOV": "Укрзалізниця",
		"UBER": "Uber",
		"УБЕР": "Uber",
		"BOLT": "Bolt",
		"БОЛТ": "Bolt", // Конфлікт з Bolt Food, але зазвичай Bolt Food має "Food" в назві
		"UKLON": "Uklon",
		"УКЛОН": "Uklon",
		"BLABLACAR": "Blablacar",
		"BlaBlaCar": "Blablacar",
		"EASYPAY": "EasyPay", // Транспорт/Квитки
		"City24": "City24",

		// === ПОШТА ===
		"NOVAPAY": "NovaPay",
		"NOVA POSHTA": "Нова Пошта",
		"НОВА ПОШТА": "Нова Пошта",
		"UKRPOSHTA": "Укрпошта",
		"УКРПОШТА": "Укрпошта",
		"MEEST": "Meest",
		"МІСТ": "Meest",
			
		// === ТЕХНІКА ТА МАРКЕТПЛЕЙСИ ===
		"ROZETKA": "Rozetka",
		"РОЗЕТКА": "Rozetka",
		"RozetkaPay": "Rozetka",
		"PROM.UA": "Prom.ua",
		"ПРОМ": "Prom.ua",
		"ALIEXPRESS": "AliExpress",
		"COMFY": "Comfy",
		"КОМФІ": "Comfy",
		"ALLO": "Allo",
		"АЛЛО": "Allo",
		"Алло": "Allo",
		"FOXTROT": "Foxtrot",
		"ФОКСТРОТ": "Foxtrot",
		"CITRUS": "Citrus",
		"ЦИТРУС": "Citrus",
		"MOYO": "Moyo",
		"МОЙО": "Moyo",
			
		// === КАФЕ ТА ЇЖА ===
		"GLOVO": "Glovo",
		"ГЛОВО": "Glovo",
		"BOLT FOOD": "Bolt Food",
		"MCDONALDS": "McDonald's",
		"MAKDONALDS": "McDonald's",
		"МАКДОНАЛЬДЗ": "McDonald's",
		"KFC": "KFC",
		"PUZATA": "Puzata Hata",
		"ПУЗАТА ХАТА": "Puzata Hata",
		"FICHEPIZZA": "fichepizza",
		"IQ PIZZA": "iq pizza",
		"VATSAK": "Вацак",
		"PPTM": "Перша пекарня твого міста",
		"PERSHA PEKARNYA": "Перша пекарня твого міста",

			// === ЗВ'ЯЗОК ТА ПІДПИСКИ ===
			"KYIVSTAR": "Kyivstar",
			"КИЇВСТАР": "Kyivstar",
			"VODAFONE": "Vodafone",
			"ВОДАФОН": "Vodafone",
			"LIFECELL": "Lifecell",
			"ЛАЙФСЕЛЛ": "Lifecell",
			"LifeCell": "Lifecell",
			"MEGOGO": "Megogo",
			"МЕГОГО": "Megogo",
			"NETFLIX": "Netflix",
			"SWEET.TV": "Sweet TV",
			"SWEET TV": "Sweet TV",
			"SPOTIFY": "Spotify",
			"YOUTUBE": "Youtube",
			"YouTube Premium": "Youtube",
			"YouTube": "Youtube",
			"GOOGLE": "Google",
			"APPLE": "Apple Services",
			"MICROSOFT": "Microsoft",
			"OPENAI": "OpenAI",
			"ADOBE": "Adobe",
			"HOSTINGER": "Hostinger",
			"DIYA": "Дія",
			"ДІЯ": "Дія",
			"Дія | Військові облігації": "Дія",
			"Виплата кешбеку від держави": "Дія",

			// === РОЗВАГИ ТА СПОРТ ===
			"MULTIPLEX": "Multiplex",
			"PLANETA KINO": "Планета кіно",
			"ПЛАНЕТА КІНО": "Планета кіно",
			"Планета Кіно": "Планета кіно",
			"SPORTLIFE": "Sportlife",
			"APOLLO NEXT1": "Apollo next",
			"APOLLO.NEXT1": "Apollo next",
			"APOLLO NEXT": "Apollo next",

			// === ЗДОРОВ'Я ===
			"SYNEVO": "Synevo",
			"СІНЕВО": "Synevo",
			"DILA": "Dila",
			"ДІЛА": "Dila",
			"ESCULAB": "Esculab",
			"MEDICOVER": "Medicover",

			"РІДІТ": "Readeat",
			"Рідіт": "Readeat",
			"ReadEAT": "Readeat",
			"ReadEat": "Readeat",
			"readeat": "Readeat",
			"READEAT": "Readeat",
			"Книгарня Readeat": "Readeat",
}
// 🔥 ГОЛОВНЕ ВИПРАВЛЕННЯ: Шукаємо НАЙДОВШЕ співпадіння
	var bestMatch string
	var longestKeyLen int

	for key, officialName := range rules {
		if strings.Contains(upperName, strings.ToUpper(key)) {
			// Якщо знайшли співпадіння, перевіряємо чи воно довше за попереднє
			// Наприклад: "UKLON" (5) > "KLO" (3). Перемагає UKLON.
			if len(key) > longestKeyLen {
				bestMatch = officialName
				longestKeyLen = len(key)
			}
		}
	}

	if bestMatch != "" {
		return bestMatch
	}

	// 2. ЧИСТКА СМІТТЯ (Якщо в словнику не знайшли)
	prefixes := regexp.MustCompile(`(?i)^(SHOP|MAGAZYN|STORE|MARKET|FOP|ФОП|ТОВ|TOV|PP|ПП|PAYMENT|Оплата|POS|TRANZAKTSIYA|ID платежу)\s*[:\.]?`)
	cleanName = prefixes.ReplaceAllString(cleanName, "")

	gateways := regexp.MustCompile(`(?i)^[A-Z0-9\.]+\s*[\*]\s*`)
	cleanName = gateways.ReplaceAllString(cleanName, "")

	if idx := strings.Index(cleanName, ","); idx != -1 {
		cleanName = cleanName[:idx]
	}

	numericTail := regexp.MustCompile(`\s+[:\d\s\.,]{4,}$`)
	cleanName = numericTail.ReplaceAllString(cleanName, "")

	idGarbage := regexp.MustCompile(`(?i)^ID\s*платежу\s*[:\d]+`)
	cleanName = idGarbage.ReplaceAllString(cleanName, "")

	return strings.Trim(cleanName, " ,.-*")
}