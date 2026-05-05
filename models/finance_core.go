package models

// Account - гаманець/рахунок
type Account struct {
	Base
	UserID   string `json:"user_id" gorm:"index"`
	FamilyID string `json:"family_id" gorm:"index"`
	ParentID string `json:"parent_id"` // Для вкладеності
	IsGroup  bool   `json:"is_group"`

	// Зв'язок з ціллю
	GoalID *string `json:"goal_id" gorm:"index;default:null"`
	Goal   *Goal   `json:"goal" gorm:"foreignKey:GoalID"`

	// Типи: 'card', 'cash', 'crypto', 'piggy_bank'
	Type string `json:"type"`

	// Для скарбничок (Piggy Bank)
	StorageTypeID *string      `json:"storage_type_id" gorm:"index;default:null"`
	StorageType   *StorageType `json:"storage_type" gorm:"foreignKey:StorageTypeID"`

	// Банківські деталі
	CardNumber    string `json:"card_number"`
	Name          string `json:"name"`
	PaymentSystem string `json:"payment_system"` // Visa/Mastercard
	Currency      string `json:"currency"`
	BankName      string `json:"bank_name"`
	CardType      string `json:"card_type"` // Credit/Debit

	// Баланси
	InitialBalance int64  `json:"initial_balance"`
	Balance        int64  `json:"balance" gorm:"column:balance"` // Поточний кешований баланс
	
	Color      string `json:"color"`
	IsArchived bool   `json:"is_archived"`
}

// Goal - фінансова ціль
type Goal struct {
	Base
	UserID   string `json:"user_id" gorm:"index"` // <-- ДОДАНО: Власник цілі
	FamilyID string `json:"family_id" gorm:"index"`

	Name         string `json:"name"`
	Description  string `json:"description"`
	TargetAmount int64  `json:"target_amount"`
	Currency     string `json:"currency"`

	DateStart    int64  `json:"date_start"`
	DateDeadline *int64 `json:"date_deadline"`

	Color string `json:"color"`
	Icon  string `json:"icon"`

	PhotoURL     string `json:"photo_url" gorm:"type:text"`
	ExternalLink string `json:"external_link"`
	Status       string `json:"status" gorm:"default:'active'"` // active, reached, failed
	RemovePhoto  bool   `json:"remove_photo" gorm:"-"`

	// <-- ДОДАНО: Налаштування приватності
	Visibility string `json:"visibility" gorm:"default:'public'"` // public / private
	HiddenFrom string `json:"hidden_from"`                        // ID юзера, від якого приховано

	// Relations
	Accounts      []Account `json:"accounts" gorm:"foreignKey:GoalID"`
	CurrentAmount int64     `json:"current_amount" gorm:"-"` // Вираховується динамічно
}

// StorageType - типи зберігання готівки (Конверт, Сейф тощо)
type StorageType struct {
	Base
	FamilyID *string `json:"family_id" gorm:"index"` // Pointer для NULL (системні типи)

	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Icon     string `json:"icon"`
	IsSystem bool   `json:"is_system" gorm:"default:false"`
}

// ShoppingList - сама картка-нотатка (наприклад, "Продукти Сільпо", "Для авто")
type ShoppingList struct {
	Base
	UserID   string `json:"user_id" gorm:"index"`
	FamilyID string `json:"family_id" gorm:"index"`

	Title      string `json:"title"`
	Color      string `json:"color" gorm:"default:'#fff7d6'"` // Можна фарбувати нотатки як у Keep
	Visibility string `json:"visibility" gorm:"default:'public'"`
	HiddenFrom string `json:"hidden_from"`

	// Зв'язок: один список має багато покупок. 
	// OnDelete:CASCADE означає, що якщо видалити список, всі покупки в ньому теж видаляться.
	Items []ShoppingItem `json:"items" gorm:"foreignKey:ListID;constraint:OnDelete:CASCADE;"`
}

// ShoppingItem - конкретний пункт у списку (наприклад, "Молоко")
type ShoppingItem struct {
	Base
	ListID string `json:"list_id" gorm:"index"` // Прив'язка до батьківської нотатки

	Name     string `json:"name"`
	IsBought bool   `json:"is_bought" gorm:"default:false"`
}

// WishlistGroup - кастомна папка для бажань (напр. "День народження", "Ремонт")
type WishlistGroup struct {
	Base
	UserID   string `json:"user_id" gorm:"index"`
	FamilyID string `json:"family_id" gorm:"index"`

	Name  string `json:"name"`
	Color string `json:"color"` // Наприклад, щоб фарбувати папку
	Icon  string `json:"icon"`  // Emoji або назва іконки

	Visibility string `json:"visibility" gorm:"default:'public'"`
	HiddenFrom string `json:"hidden_from"`
}

// WishlistItem - оновлений елемент
type WishlistItem struct {
	Base
	UserID   string `json:"user_id" gorm:"index"`
	FamilyID string `json:"family_id" gorm:"index"`

	// Зв'язок з групою (Pointer, бо може бути без групи - "Загальне")
	GroupID *string        `json:"group_id" gorm:"index;default:null"`
	Group   *WishlistGroup `json:"group" gorm:"foreignKey:GroupID"`

	Name     string `json:"name"`
	URL      string `json:"url"`
	Price    int64  `json:"price"`
	Currency string `json:"currency"`

	Priority int    `json:"priority" gorm:"default:1"`
	Status   string `json:"status" gorm:"default:'planning'"`

	PhotoURL string `json:"photo_url" gorm:"type:text"`

	Visibility string `json:"visibility" gorm:"default:'public'"`
	HiddenFrom string `json:"hidden_from"`

	ReservedByUserID *string `json:"reserved_by"`

	GoalID *string `json:"goal_id" gorm:"index;default:null"`
	Goal   *Goal   `json:"goal" gorm:"foreignKey:GoalID"`
}