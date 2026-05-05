package utils

import (
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- CATEGORIES SEEDING ---

type CategorySeed struct {
	Name     string
	Type     string
	Color    string
	Icon     string 
	Children []CategorySeed
}

func SeedFamilyDefaults(db *gorm.DB, familyID string) error {
	categories := []CategorySeed{
		// 1. ПРОДУКТИ
		{
			Name:  "Продукти",
			Type:  "expense",
			Color: "#f97316", // Orange
			Icon:  "HiShoppingCart",
			Children: []CategorySeed{
				{Name: "М'ясо та риба", Type: "expense", Color: "#fdba74", Icon: "HiFire"},
				{Name: "Молочка та яйця", Type: "expense", Color: "#fed7aa", Icon: "HiCube"},
				{Name: "Овочі та фрукти", Type: "expense", Color: "#86efac", Icon: "HiSun"},
				{Name: "Хліб та випічка", Type: "expense", Color: "#fb923c", Icon: "HiShoppingBag"}, // ⬅️ ЗАМІНА (був Супермаркет)
				{Name: "Бакалія", Type: "expense", Color: "#fde047", Icon: "HiArchiveBox"}, // Крупи, макарони, олія
				{Name: "Солодощі", Type: "expense", Color: "#f472b6", Icon: "HiCake"},
				{Name: "Напої (додому)", Type: "expense", Color: "#a5f3fc", Icon: "HiBeaker"},
				{Name: "Алкоголь", Type: "expense", Color: "#ef4444", Icon: "HiNoSymbol"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 2. КАФЕ ТА РЕСТОРАНИ
		{
			Name:  "Кафе та Ресторани",
			Type:  "expense",
			Color: "#f59e0b", // Amber
			Icon:  "HiBuildingStorefront",
			Children: []CategorySeed{
				{Name: "Ресторани", Type: "expense", Color: "#fbbf24", Icon: "HiStar"},
				{Name: "Кава та перекуси", Type: "expense", Color: "#fcd34d", Icon: "HiBeaker"}, // Заміна HiMug
				{Name: "Фастфуд", Type: "expense", Color: "#d97706", Icon: "HiBolt"}, // Заміна HiLightningBolt
				{Name: "Доставка їжі", Type: "expense", Color: "#b45309", Icon: "HiTruck"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 3. ВЛАСНЕ АВТО
		{
			Name:  "Власне Авто",
			Type:  "expense",
			Color: "#3b82f6", // Blue
			Icon:  "HiKey",
			Children: []CategorySeed{
				{Name: "Пальне", Type: "expense", Color: "#60a5fa", Icon: "HiFunnel"},
				{Name: "СТО та Ремонт", Type: "expense", Color: "#2563eb", Icon: "HiWrenchScrewdriver"},
				{Name: "Страхування (Авто)", Type: "expense", Color: "#1d4ed8", Icon: "HiShieldCheck"},
				{Name: "Мийка та догляд", Type: "expense", Color: "#93c5fd", Icon: "HiSparkles"},
				{Name: "Платні дороги/Парковка", Type: "expense", Color: "#1e40af", Icon: "HiMap"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 4. МІСЬКИЙ ТРАНСПОРТ
		{
			Name:  "Громадський Транспорт",
			Type:  "expense",
			Color: "#0ea5e9", // Sky
			Icon:  "HiMapPin",
			Children: []CategorySeed{
				{Name: "Таксі", Type: "expense", Color: "#38bdf8", Icon: "HiMap"},
				{Name: "Метро / Автобус", Type: "expense", Color: "#7dd3fc", Icon: "HiTicket"},
				{Name: "Каршерінг / Самокати", Type: "expense", Color: "#0284c7", Icon: "HiBolt"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 5. ЖИТЛО
		{
			Name:  "Житло",
			Type:  "expense",
			Color: "#6366f1", // Indigo
			Icon:  "HiHome",
			Children: []CategorySeed{
				{Name: "Оренда / Іпотека", Type: "expense", Color: "#818cf8", Icon: "HiCurrencyDollar"},
				{Name: "Електроенергія", Type: "expense", Color: "#facc15", Icon: "HiBolt"},
				{Name: "Газ та опалення", Type: "expense", Color: "#fb923c", Icon: "HiFire"},
				{Name: "Вода", Type: "expense", Color: "#60a5fa", Icon: "HiCloud"},
				{Name: "Квартплата", Type: "expense", Color: "#a5b4fc", Icon: "HiBuildingOffice"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 6. ЗВ'ЯЗОК
		{
			Name:  "Зв'язок та Інтернет",
			Type:  "expense",
			Color: "#8b5cf6", // Violet
			Icon:  "HiWifi",
			Children: []CategorySeed{
				{Name: "Мобільний зв'язок", Type: "expense", Color: "#a78bfa", Icon: "HiDevicePhoneMobile"},
				{Name: "Домашній інтернет", Type: "expense", Color: "#c4b5fd", Icon: "HiGlobeAlt"},
				{Name: "Підписки (TV/Soft)", Type: "expense", Color: "#ddd6fe", Icon: "HiPlay"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 7. ПОКУПКИ
		{
			Name:  "Покупки",
			Type:  "expense",
			Color: "#ec4899", // Pink
			Icon:  "HiShoppingBag",
			Children: []CategorySeed{
				{Name: "Одяг та взуття", Type: "expense", Color: "#f472b6", Icon: "HiTag"},
				{Name: "Техніка та аксесуари", Type: "expense", Color: "#d946ef", Icon: "HiDeviceTablet"},
				{Name: "Ювелірні вироби", Type: "expense", Color: "#fbcfe8", Icon: "HiSparkles"},
				{Name: "Косметика", Type: "expense", Color: "#f9a8d4", Icon: "HiFaceSmile"},
				{Name: "Подарунки", Type: "expense", Color: "#db2777", Icon: "HiGift"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 8. ПОСЛУГИ ТА СЕРВІС
		{
			Name:  "Послуги та Сервіс",
			Type:  "expense",
			Color: "#14b8a6", // Teal
			Icon:  "HiClipboardDocumentList",
			Children: []CategorySeed{
				{Name: "Поштові послуги", Type: "expense", Color: "#2dd4bf", Icon: "HiInbox"},
				{Name: "Хімчистка / Ремонт одягу", Type: "expense", Color: "#5eead4", Icon: "HiTag"},
				{Name: "Юридичні послуги", Type: "expense", Color: "#0d9488", Icon: "HiScale"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 9. ДІМ ТА ПОБУТ
		{
			Name:  "Дім та Побут",
			Type:  "expense",
			Color: "#10b981", // Emerald
			Icon:  "HiHomeModern",
			Children: []CategorySeed{
				{Name: "Побутова хімія", Type: "expense", Color: "#34d399", Icon: "HiSparkles"},
				{Name: "Дрібниці для дому", Type: "expense", Color: "#6ee7b7", Icon: "HiArchiveBox"},
				{Name: "Меблі та інтер'єр", Type: "expense", Color: "#059669", Icon: "HiSwatch"}, // Заміна HiTable
				{Name: "Ремонт (матеріали)", Type: "expense", Color: "#047857", Icon: "HiWrench"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 10. ЗДОРОВ'Я
		{
			Name:  "Здоров'я",
			Type:  "expense",
			Color: "#ef4444", // Red
			Icon:  "HiHeart",
			Children: []CategorySeed{
				{Name: "Аптека", Type: "expense", Color: "#fca5a5", Icon: "HiPlusCircle"},
				{Name: "Лікарі", Type: "expense", Color: "#f87171", Icon: "HiUser"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 11. ОСВІТА
		{
			Name:  "Освіта",
			Type:  "expense",
			Color: "#059669", // Emerald darker
			Icon:  "HiAcademicCap",
			Children: []CategorySeed{
				{Name: "Книги", Type: "expense", Color: "#34d399", Icon: "HiBookOpen"},
				{Name: "Курси", Type: "expense", Color: "#10b981", Icon: "HiPresentationChartLine"},
				{Name: "Мови", Type: "expense", Color: "#059669", Icon: "HiChatBubbleLeftRight"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 12. РОЗВАГИ
		{
			Name:  "Розваги",
			Type:  "expense",
			Color: "#d946ef", // Fuchsia
			Icon:  "HiTicket",
			Children: []CategorySeed{
				{Name: "Кіно / Театр", Type: "expense", Color: "#e879f9", Icon: "HiFilm"},
				{Name: "Хобі", Type: "expense", Color: "#f0abfc", Icon: "HiMusicalNote"},
				{Name: "Спорт", Type: "expense", Color: "#fae8ff", Icon: "HiTrophy"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 13. ПОДОРОЖІ
		{
			Name: "Подорожі",
			Type: "expense",
			Color: "#06b6d4", // Cyan
			Icon: "HiGlobeAmericas",
			Children: []CategorySeed{
				{Name: "Квитки (Потяг)", Type: "expense", Color: "#22d3ee", Icon: "HiTicket"},
				{Name: "Квитки (Літак)", Type: "expense", Color: "#67e8f9", Icon: "HiPaperAirplane"},
				{Name: "Готелі", Type: "expense", Color: "#a5f3fc", Icon: "HiBuildingOffice"},
				{Name: "Відпустка", Type: "expense", Color: "#0891b2", Icon: "HiSun"},
				{Name: "Ділові поїздки", Type: "expense", Color: "#155e75", Icon: "HiBriefcase"},
				{Name: "Дальні поїздки", Type: "expense", Color: "#0e7490", Icon: "HiMap"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 14. ЗОБОВ'ЯЗАННЯ
		{
			Name:  "Зобов'язання",
			Type:  "expense",
			Color: "#64748b", // Slate
			Icon:  "HiScale",
			Children: []CategorySeed{
				{Name: "Податки", Type: "expense", Color: "#94a3b8", Icon: "HiDocumentText"},
				{Name: "Штрафи", Type: "expense", Color: "#ef4444", Icon: "HiExclamationTriangle"},
				{Name: "Аліменти", Type: "expense", Color: "#cbd5e1", Icon: "HiUsers"},
				{Name: "Комісії банку", Type: "expense", Color: "#e2e8f0", Icon: "HiReceiptPercent"},
				{Name: "Кредити / Борги", Type: "expense", Color: "#475569", Icon: "HiBanknotes"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// 15. ФІНАНСИ
		{
			Name:  "Фінанси та Допомога",
			Type:  "expense",
			Color: "#84cc16", // Lime
			Icon:  "HiBanknotes",
			Children: []CategorySeed{
				{Name: "Благодійність", Type: "expense", Color: "#a3e635", Icon: "HiHandRaised"},
				{Name: "Допомога рідним", Type: "expense", Color: "#bef264", Icon: "HiUserGroup"},
				{Name: "Втрачені кошти", Type: "expense", Color: "#d9f99d", Icon: "HiTrash"},
				{Name: "Інше", Type: "expense", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
		// --- ДОХОДИ ---
		{
			Name:  "Доходи",
			Type:  "income",
			Color: "#22c55e", // Green
			Icon:  "HiBriefcase",
			Children: []CategorySeed{
				{Name: "Зарплата", Type: "income", Color: "#4ade80", Icon: "HiBanknotes"},
				{Name: "Пасивний дохід", Type: "income", Color: "#86efac", Icon: "HiChartBar"},
				{Name: "Подарунки", Type: "income", Color: "#bbf7d0", Icon: "HiGift"},
				{Name: "Інше", Type: "income", Color: "#d1d5db", Icon: "HiEllipsisHorizontalCircle"},
			},
		},
	}
    
    // ... createCategory код залишається без змін
    var createCategory func(seed CategorySeed, parentID string) error
	createCategory = func(seed CategorySeed, parentID string) error {
		cat := models.Category{
			Base: models.Base{
				ID:        uuid.NewString(),
				CreatedAt: time.Now().UnixMilli(),
				UpdatedAt: time.Now().UnixMilli(),
				IsSynced:  true,
			},
			FamilyID: familyID,
			ParentID: parentID,
			Name:     seed.Name,
			Type:     seed.Type,
			Color:    seed.Color,
			Icon:     seed.Icon,
		}
		if err := db.Create(&cat).Error; err != nil { return err }
		for _, child := range seed.Children {
			if err := createCategory(child, cat.ID); err != nil { return err }
		}
		return nil
	}
	for _, rootCat := range categories {
		if err := createCategory(rootCat, ""); err != nil { return err }
	}
	return nil
}

// SeedDefaultCounterparties також оновлюємо під нові категорії
// --- COUNTERPARTIES SEEDING (ОНОВЛЕНО) ---

func SeedDefaultCounterparties(db *gorm.DB, familyID string) error {
	now := time.Now().UnixMilli()

	// 1. СТВОРЮЄМО КАТЕГОРІЇ КОНТРАГЕНТІВ
	// Додав нові, щоб розвантажити "Інше" та "Дрогері"
	categoriesData := []models.CounterpartyCategory{
		// Магазини (Type: shop)
		{Name: "Супермаркети", Type: "shop", Icon: "HiShoppingCart", Color: "#f97316"},
		{Name: "Дрогері та Аптеки", Type: "shop", Icon: "HiSparkles", Color: "#ec4899"},
		{Name: "Дім та Ремонт", Type: "shop", Icon: "HiHome", Color: "#10b981"},         // <--- НОВЕ (Епіцентр, Юск)
		{Name: "Одяг та Аксесуари", Type: "shop", Icon: "HiTag", Color: "#f472b6"},      // <--- НОВЕ (Sinsay, Staff)
		{Name: "Техніка та Електроніка", Type: "shop", Icon: "HiComputerDesktop", Color: "#3b82f6"},
		{Name: "Маркетплейси", Type: "shop", Icon: "HiShoppingBag", Color: "#84cc16"},
		{Name: "АЗС", Type: "shop", Icon: "HiFunnel", Color: "#eab308"},
		{Name: "Кафе та Ресторани", Type: "shop", Icon: "HiCake", Color: "#f59e0b"},

		// Сервіси (Type: other)
		{Name: "Доставка їжі", Type: "other", Icon: "HiTruck", Color: "#ef4444"}, // Glovo - це послуга
		{Name: "Таксі та Транспорт", Type: "other", Icon: "HiMapPin", Color: "#0ea5e9"},
		{Name: "Поштові служби", Type: "other", Icon: "HiInbox", Color: "#6366f1"},
		{Name: "Зв'язок та Інтернет", Type: "other", Icon: "HiWifi", Color: "#8b5cf6"},
		{Name: "Підписки та Сервіси", Type: "other", Icon: "HiPlay", Color: "#a855f7"},
		{Name: "Розваги та Спорт", Type: "other", Icon: "HiTicket", Color: "#d946ef"},   // <--- НОВЕ (Кіно, Зал)
		{Name: "Здоров'я", Type: "other", Icon: "HiHeart", Color: "#ef4444"},
		{Name: "Комунальні послуги", Type: "other", Icon: "HiLightningBolt", Color: "#f59e0b"}, // Або HiHome
	}

	catMap := make(map[string]string)
	for _, catData := range categoriesData {
		cat := models.CounterpartyCategory{
			Base:     models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
			FamilyID: familyID, Name: catData.Name, Type: catData.Type, Icon: catData.Icon, Color: catData.Color,
		}
		if err := db.Create(&cat).Error; err != nil {
			return err
		}
		catMap[cat.Name] = cat.ID
	}

	// 2. СТВОРЮЄМО КОНТРАГЕНТІВ
	type CpSeed struct {
		Name, Type, CatName, Logo string
	}

	cpDefaults := []CpSeed{
		// === СУПЕРМАРКЕТИ (Продукти) ===
		{Name: "Сільпо", Type: "shop", CatName: "Супермаркети", Logo: "silpo.svg"},
		{Name: "АТБ", Type: "shop", CatName: "Супермаркети", Logo: "atb.svg"},
		{Name: "Novus", Type: "shop", CatName: "Супермаркети", Logo: "novus.svg"},
		{Name: "Varus", Type: "shop", CatName: "Супермаркети", Logo: "varus.svg"},
		{Name: "Ашан", Type: "shop", CatName: "Супермаркети", Logo: "auchan.svg"},
		{Name: "Metro", Type: "shop", CatName: "Супермаркети", Logo: "metro.svg"},
		{Name: "Фора", Type: "shop", CatName: "Супермаркети", Logo: "fora.svg"},
		{Name: "Roshen", Type: "shop", CatName: "Супермаркети", Logo: "roshen.svg"},
		{Name: "Стовпинські ковбаси", Type: "shop", CatName: "Супермаркети", Logo: "stovpynski_kovbasy.svg"},
		{Name: "Близенько", Type: "shop", CatName: "Супермаркети", Logo: "blyzenko.svg"},
		{Name: "Сімі", Type: "shop", CatName: "Супермаркети", Logo: "simi.svg"},
		{Name: "Сім23", Type: "shop", CatName: "Супермаркети", Logo: "sim23.svg"},
		{Name: "М'ясомаркет", Type: "shop", CatName: "Супермаркети", Logo: "myasomarket.svg"},
		{Name: "Рукавичка", Type: "shop", CatName: "Супермаркети", Logo: "rukavychka.svg"},

		// === ДРОГЕРІ ТА АПТЕКИ (Хімія + Ліки) ===
		{Name: "Eva", Type: "shop", CatName: "Дрогері та Аптеки", Logo: "eva.svg"},
		{Name: "Watsons", Type: "shop", CatName: "Дрогері та Аптеки", Logo: "watsons.svg"},
		{Name: "Prostor", Type: "shop", CatName: "Дрогері та Аптеки", Logo: "prostor.svg"},
		{Name: "Аврора", Type: "shop", CatName: "Дрогері та Аптеки", Logo: "avrora.svg"},
		{Name: "Копійочка", Type: "shop", CatName: "Дрогері та Аптеки", Logo: "kopiyochka.svg"},
		{Name: "Makeup", Type: "shop", CatName: "Дрогері та Аптеки", Logo: "makeup.svg"},

		// Аптеки
		{Name: "Аптека АНЦ", Type: "shop", CatName: "Дрогері та Аптеки", Logo: "anc.svg"},
		{Name: "Аптека Подорожник", Type: "shop", CatName: "Дрогері та Аптеки", Logo: "podorozhnyk.svg"},
		{Name: "Аптека 9-1-1", Type: "shop", CatName: "Дрогері та Аптеки", Logo: "911.svg"},
		{Name: "Аптека Доброго Дня", Type: "shop", CatName: "Дрогері та Аптеки", Logo: "add.svg"},
		{Name: "Ощад Аптека", Type: "shop", CatName: "Дрогері та Аптеки", Logo: "oshchad_apteka.svg"},
		{Name: "Аптека оптових цін", Type: "shop", CatName: "Дрогері та Аптеки", Logo: "apteka_optovykh_tsin.svg"},

		// === ДІМ ТА РЕМОНТ (Все для дому) ===
		{Name: "Епіцентр", Type: "shop", CatName: "Дім та Ремонт", Logo: "epitsentr.svg"},
		{Name: "Нова Лінія", Type: "shop", CatName: "Дім та Ремонт", Logo: "nova_liniya.svg"},
		{Name: "Jysk", Type: "shop", CatName: "Дім та Ремонт", Logo: "jysk.svg"},
		{Name: "Є Таке!", Type: "shop", CatName: "Дім та Ремонт", Logo: "ye_take.svg"},

		// === ОДЯГ ТА АКСЕСУАРИ ===
		{Name: "Sinsay", Type: "shop", CatName: "Одяг та Аксесуари", Logo: "sinsay.svg"},
		{Name: "Staff", Type: "shop", CatName: "Одяг та Аксесуари", Logo: "staff.svg"},
		{Name: "Будинок іграшок", Type: "shop", CatName: "Одяг та Аксесуари", Logo: "bi.svg"}, // Або окремо, але тут ок
		{Name: "Intertop", Type: "shop", CatName: "Одяг та Аксесуари", Logo: "intertop.svg"},

		// === АЗС ===
		{Name: "OKKO", Type: "shop", CatName: "АЗС", Logo: "okko.svg"},
		{Name: "WOG", Type: "shop", CatName: "АЗС", Logo: "wog.svg"},
		{Name: "SOCAR", Type: "shop", CatName: "АЗС", Logo: "socar.svg"},
		{Name: "UPG", Type: "shop", CatName: "АЗС", Logo: "upg.svg"},
		{Name: "KLO", Type: "shop", CatName: "АЗС", Logo: "klo.svg"},
		{Name: "Shell", Type: "shop", CatName: "АЗС", Logo: "shell.svg"},
		{Name: "Amic", Type: "shop", CatName: "АЗС", Logo: "amic.svg"},
		{Name: "BVS", Type: "shop", CatName: "АЗС", Logo: "bvs.svg"},

		// === ТЕХНІКА ТА ЕЛЕКТРОНІКА ===
		{Name: "Comfy", Type: "shop", CatName: "Техніка та Електроніка", Logo: "comfy.svg"},
		{Name: "Allo", Type: "shop", CatName: "Техніка та Електроніка", Logo: "allo.svg"},
		{Name: "Foxtrot", Type: "shop", CatName: "Техніка та Електроніка", Logo: "foxtrot.svg"},
		{Name: "Citrus", Type: "shop", CatName: "Техніка та Електроніка", Logo: "citrus.svg"},
		{Name: "Moyo", Type: "shop", CatName: "Техніка та Електроніка", Logo: "moyo.svg"},

		// === МАРКЕТПЛЕЙСИ ===
		{Name: "Rozetka", Type: "shop", CatName: "Маркетплейси", Logo: "rozetka.svg"},
		{Name: "Prom.ua", Type: "shop", CatName: "Маркетплейси", Logo: "prom.svg"},
		{Name: "AliExpress", Type: "shop", CatName: "Маркетплейси", Logo: "aliexpress.svg"}, // Переніс сюди з підписок
			
		// === КАФЕ ТА РЕСТОРАНИ ===
		{Name: "McDonald's", Type: "shop", CatName: "Кафе та Ресторани", Logo: "mcdonalds.svg"},
		{Name: "KFC", Type: "shop", CatName: "Кафе та Ресторани", Logo: "kfc.svg"},
		{Name: "Puzata Hata", Type: "shop", CatName: "Кафе та Ресторани", Logo: "puzatahata.svg"},
		{Name: "fichepizza", Type: "shop", CatName: "Кафе та Ресторани", Logo: "fichepizza.svg"},
		{Name: "iq pizza", Type: "shop", CatName: "Кафе та Ресторани", Logo: "iq_pizza.svg"},
		{Name: "Вацак", Type: "shop", CatName: "Кафе та Ресторани", Logo: "vatsak.svg"}, 

		// Кондитерська = кафе/магазин
		{Name: "Перша пекарня твого міста", Type: "shop", CatName: "Кафе та Ресторани", Logo: "pptm.svg"},

		// === ПОШТА ===
		{Name: "Нова Пошта", Type: "other", CatName: "Поштові служби", Logo: "novaposhta.svg"},
		{Name: "Укрпошта", Type: "other", CatName: "Поштові служби", Logo: "ukrposhta.svg"},
		{Name: "Meest", Type: "other", CatName: "Поштові служби", Logo: "meest.svg"},
			
		// === ТАКСІ ТА ТРАНСПОРТ ===
		{Name: "Uklon", Type: "other", CatName: "Таксі та Транспорт", Logo: "uklon.svg"},
		{Name: "Bolt", Type: "other", CatName: "Таксі та Транспорт", Logo: "bolt.svg"},
		{Name: "Uber", Type: "other", CatName: "Таксі та Транспорт", Logo: "uber.svg"},
		{Name: "Укрзалізниця", Type: "other", CatName: "Таксі та Транспорт", Logo: "uz.svg"},
		{Name: "Blablacar", Type: "other", CatName: "Таксі та Транспорт", Logo: "blablacar.svg"},

		// === ДОСТАВКА ЇЖІ (Це сервіс, тому 'other') ===
		{Name: "Glovo", Type: "other", CatName: "Доставка їжі", Logo: "glovo.svg"},
		{Name: "Bolt Food", Type: "other", CatName: "Доставка їжі", Logo: "bolt_food.svg"},
			
		// === ЗВ'ЯЗОК ===
		{Name: "Kyivstar", Type: "other", CatName: "Зв'язок та Інтернет", Logo: "kyivstar.svg"},
		{Name: "Vodafone", Type: "other", CatName: "Зв'язок та Інтернет", Logo: "vodafone.svg"},
		{Name: "Lifecell", Type: "other", CatName: "Зв'язок та Інтернет", Logo: "lifecell.svg"},

		// === ПІДПИСКИ ТА СЕРВІСИ ===
		{Name: "Megogo", Type: "other", CatName: "Підписки та Сервіси", Logo: "megogo.svg"},
		{Name: "Netflix", Type: "other", CatName: "Підписки та Сервіси", Logo: "netflix.svg"},
		{Name: "Sweet TV", Type: "other", CatName: "Підписки та Сервіси", Logo: "sweet_tv.svg"},
		{Name: "Spotify", Type: "other", CatName: "Підписки та Сервіси", Logo: "spotify.svg"},
		{Name: "Youtube", Type: "other", CatName: "Підписки та Сервіси", Logo: "youtube.svg"},
		{Name: "Kyivstar TV", Type: "other", CatName: "Підписки та Сервіси", Logo: "kyivstar_tv.svg"},
		{Name: "Google", Type: "other", CatName: "Підписки та Сервіси", Logo: "google.svg"},
		{Name: "Apple Services", Type: "other", CatName: "Підписки та Сервіси", Logo: "apple.svg"},
		{Name: "Microsoft", Type: "other", CatName: "Підписки та Сервіси", Logo: "microsoft.svg"},
		{Name: "Hostinger", Type: "other", CatName: "Підписки та Сервіси", Logo: "hostinger.svg"},
		{Name: "EasyPay", Type: "other", CatName: "Підписки та Сервіси", Logo: "easypay.svg"},
		{Name: "OpenAI", Type: "other", CatName: "Підписки та Сервіси", Logo: "openai.svg"},
		{Name: "Adobe", Type: "other", CatName: "Підписки та Сервіси", Logo: "adobe.svg"},
		{Name: "Дія", Type: "other", CatName: "Підписки та Сервіси", Logo: "diya.svg"},
		{Name: "NovaPay", Type: "other", CatName: "Підписки та Сервіси", Logo: "novapay.svg"}, // Фін. сервіс
		{Name: "City24", Type: "other", CatName: "Підписки та Сервіси", Logo: "city24.svg"},

		// === РОЗВАГИ ТА СПОРТ ===
		{Name: "Multiplex", Type: "other", CatName: "Розваги та Спорт", Logo: "multiplex.svg"},
		{Name: "Планета кіно", Type: "other", CatName: "Розваги та Спорт", Logo: "planeta_kino.svg"},
		{Name: "Sportlife", Type: "other", CatName: "Розваги та Спорт", Logo: "sportlife.svg"},
		{Name: "Apollo next", Type: "other", CatName: "Розваги та Спорт", Logo: "apollo_next.svg"},

		// === ЗДОРОВ'Я ===
		{Name: "Synevo", Type: "other", CatName: "Здоров'я", Logo: "synevo.svg"},
		{Name: "Dila", Type: "other", CatName: "Здоров'я", Logo: "dila.svg"},
		{Name: "Esculab", Type: "other", CatName: "Здоров'я", Logo: "esculab.svg"},
		{Name: "Medicover", Type: "other", CatName: "Здоров'я", Logo: "medicover.svg"},
		{Name: "Dental clinic", Type: "other", CatName: "Здоров'я", Logo: "dental_clinic.svg"},

		//Комунальні послуги
		{Name: "YASNO", Type: "other", CatName: "Комунальні послуги", Logo: "yasno.svg"},
		{Name: "ДТЕК", Type: "other", CatName: "Комунальні послуги", Logo: "dtek.svg"},
		{Name: "Нафтогаз", Type: "other", CatName: "Комунальні послуги", Logo: "naftogaz.svg"},
		{Name: "Водоканал", Type: "other", CatName: "Комунальні послуги", Logo: "vodokanal.svg"},
		{Name: "Київтеплоенерго", Type: "other", CatName: "Комунальні послуги", Logo: "kte.svg"},
		{Name: "Львівтеплоенерго", Type: "other", CatName: "Комунальні послуги", Logo: "lte.svg"},
		{Name: "Житомиробленерго", Type: "other", CatName: "Комунальні послуги", Logo: "zhte.svg"},
		{Name: "ОСББ / ЖЕК", Type: "other", CatName: "Комунальні послуги", Logo: "osbb.svg"},
		{Name: "Укртелеком", Type: "other", CatName: "Комунальні послуги", Logo: "ukrtelecom.svg"},
		{Name: "Volia", Type: "other", CatName: "Комунальні послуги", Logo: "volia.svg"},
	}

	var counterparties []models.Counterparty
	for _, d := range cpDefaults {
		var catID *string
		if id, ok := catMap[d.CatName]; ok {
			catID = &id
		}
		counterparties = append(counterparties, models.Counterparty{
			Base:       models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
			FamilyID:   familyID,
			Name:       d.Name,
			Type:       d.Type,
			CategoryID: catID,
			Logo:       d.Logo,
		})
	}

	return db.Create(&counterparties).Error
}


// ... (твій попередній код Counterparties)

// --- STORAGE TYPES SEEDING (СИСТЕМНІ ТИПИ СКАРБНИЧОК) ---

func SeedSystemStorageTypes(db *gorm.DB) error {
	// Список системних типів (FamilyID = nil)
	types := []models.StorageType{
		{Name: "Конверт", Slug: "envelope", Icon: "HiEnvelope", IsSystem: true},
		{Name: "Сейф", Slug: "safe", Icon: "HiLockClosed", IsSystem: true},
		{Name: "Банка", Slug: "jar", Icon: "HiArchiveBox", IsSystem: true}, // Для Монобанку
		{Name: "Скарбничка", Slug: "piggy", Icon: "HiCurrencyDollar", IsSystem: true}, // Класична свинка
		{Name: "Готівка (Схов)", Slug: "cash_stash", Icon: "HiBanknotes", IsSystem: true},
		{Name: "Крипто-гаманець", Slug: "crypto_wallet", Icon: "HiCpuChip", IsSystem: true},
	}

	for _, t := range types {
		// 1. Перевіряємо, чи існує такий тип за Slug
		var count int64
		// Шукаємо тільки серед системних (family_id IS NULL)
		err := db.Model(&models.StorageType{}).
			Where("slug = ? AND family_id IS NULL", t.Slug).
			Count(&count).Error
		
		if err != nil {
			return err
		}

		// 2. Якщо немає — створюємо
		if count == 0 {
			newType := models.StorageType{
				Base: models.Base{
					ID:        uuid.NewString(),
					CreatedAt: time.Now().UnixMilli(),
					UpdatedAt: time.Now().UnixMilli(),
					IsSynced:  true,
				},
				FamilyID: nil, // Це важливо для системних типів
				Name:     t.Name,
				Slug:     t.Slug,
				Icon:     t.Icon,
				IsSystem: true,
			}
			if err := db.Create(&newType).Error; err != nil {
				return err
			}
		}
	}
	return nil
}