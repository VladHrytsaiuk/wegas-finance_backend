package services

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateAccountInput struct {
	Name           string
	Type           string
	Currency       string
	InitialBalance int64
	Color          string
	CardNumber     string
	PaymentSystem  string
	OwnerID        string
	BankName       string // Для карток (mono, privat...)
	CardType       string // Для карток (black, gold...)
	StorageTypeID  *string
	GoalID         *string
}

type AccountService interface {
	Create(input CreateAccountInput, user *models.User) (*models.Account, error)
	GetAll(user *models.User) ([]models.Account, error)
	GetByID(id string, user *models.User) (*models.Account, error)
	Update(id string, input CreateAccountInput, user *models.User) (*models.Account, error)
	Delete(id string, user *models.User) error
	UpdateMobileOrder(accountIDs []string, user *models.User) error
}

type accountService struct {
	repo repositories.AccountRepository
	db   *gorm.DB // Для доступу до транзакцій та маппінгів
}

func NewAccountService(repo repositories.AccountRepository, db *gorm.DB) AccountService {
	return &accountService{repo: repo, db: db}
}

func (s *accountService) Create(input CreateAccountInput, user *models.User) (*models.Account, error) {
	ownerID := input.OwnerID

	// ЗАХИСТ: Дитина створює рахунок тільки для себе
	if user.RoleID == "child" {
		ownerID = user.ID
	} else if ownerID == "" {
		ownerID = user.ID
	}

	paymentSystem := input.PaymentSystem
	if input.Type == "card" && paymentSystem == "" {
		paymentSystem = ""
	}

	account := &models.Account{
		Base: models.Base{
			ID:        uuid.NewString(),
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
			IsSynced:  false, // Ручні рахунки за замовчуванням
		},
		UserID:         ownerID,
		FamilyID:       user.FamilyID,
		Name:           input.Name,
		Type:           input.Type,
		Currency:       input.Currency,
		InitialBalance: input.InitialBalance,
		Balance:        input.InitialBalance,
		Color:          input.Color,
		BankName:       input.BankName,
		CardType:       input.CardType,
		CardNumber:     input.CardNumber,
		PaymentSystem:  paymentSystem,
		IsArchived:     false,
		IsGroup:        false,
		StorageTypeID:  input.StorageTypeID,
		GoalID:         input.GoalID,
	}

	if err := s.repo.Create(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *accountService) GetAll(user *models.User) ([]models.Account, error) {
	// ЗАХИСТ: Дитина бачить тільки свої, батьки — всієї сім'ї
	if user.RoleID == "child" {
		return s.repo.GetAllByUserID(user.ID)
	}
	return s.repo.GetAllByFamilyID(user.FamilyID)
}

func (s *accountService) GetByID(id string, user *models.User) (*models.Account, error) {
	account, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Перевірка сімейного доступу
	if account.FamilyID != user.FamilyID {
		return nil, errors.New("access denied")
	}

	// ЗАХИСТ: Дитина не бачить чужі рахунки навіть по прямим посиланням
	if user.RoleID == "child" && account.UserID != user.ID {
		return nil, errors.New("access denied: child cannot view this account")
	}

	return account, nil
}

func (s *accountService) Update(id string, input CreateAccountInput, user *models.User) (*models.Account, error) {
	account, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("account not found")
	}

	// Перевірка прав
	if account.FamilyID != user.FamilyID {
		return nil, errors.New("access denied")
	}
	if user.RoleID == "child" && account.UserID != user.ID {
		return nil, errors.New("access denied: cannot update others' accounts")
	}

	// 🔥 НОВА ЛОГІКА ТИПУ: Якщо тип змінюється на будь-який, крім "скарбнички",
	// ми занулюємо посилання на ціль.
	finalGoalID := input.GoalID
	if input.Type != "piggy_bank" {
		finalGoalID = nil
	}

	// ЛОГІКА БАЛАНСУ ТА СИСТЕМНИХ ПОЛІВ
	if !account.IsSynced {
		// Для ручних рахунків дозволяємо міняти фінанси
		if input.InitialBalance != account.InitialBalance {
			diff := input.InitialBalance - account.InitialBalance
			account.Balance += diff
		}
		account.InitialBalance = input.InitialBalance
		account.Currency = input.Currency
		account.PaymentSystem = input.PaymentSystem
		account.BankName = input.BankName
		account.CardType = input.CardType
	}
	// Якщо IsSynced == true, GORM проігнорує зміни балансу/валюти (вони не зміняться в об'єкті)

	// Поля, які можна редагувати завжди
	account.Name = input.Name
	account.Type = input.Type
	account.Color = input.Color
	account.CardNumber = input.CardNumber
	account.StorageTypeID = input.StorageTypeID
	account.GoalID = finalGoalID

	// Зміна власника (тільки для батьків)
	if user.RoleID != "child" && input.OwnerID != "" {
		account.UserID = input.OwnerID
	}

	account.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.Update(account); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *accountService) Delete(id string, user *models.User) error {
	account, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("account not found")
	}

	if account.FamilyID != user.FamilyID {
		return errors.New("access denied")
	}

	if user.RoleID == "child" && account.UserID != user.ID {
		return errors.New("access denied")
	}

	// Використовуємо транзакцію тільки для того, що реально треба оновити разом
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. ВІД'ЄДНУЄМО ВІД ЦІЛІ
		// Ми занулюємо посилання, щоб ціль більше не рахувала цей баланс у свій прогрес
		if err := tx.Model(&models.Account{}).
			Where("id = ?", id).
			Update("goal_id", nil).Error; err != nil {
			return err
		}

		// 2. ЛОГІКА ВИДАЛЕННЯ СИНХРОНІЗОВАНИХ (Monobank)
		if account.IsSynced {
			if err := tx.Where("internal_account_id = ?", id).
				Delete(&models.BankAccountMapping{}).Error; err != nil {
				return err
			}
		}

		// 3. SOFT DELETE РАХУНКУ (Робимо через tx, щоб не було database is locked)
		// Ми оновлюємо deleted_at прямо тут, не викликаючи repo.Delete
		return tx.Model(&models.Account{}).
			Where("id = ?", id).
			Update("deleted_at", time.Now().UnixMilli()).Error
	})
}

func (s *accountService) UpdateMobileOrder(accountIDs []string, user *models.User) error {
	accessibleAccounts, err := s.GetAll(user)
	if err != nil {
		return err
	}

	accessibleIDs := make(map[string]struct{}, len(accessibleAccounts))
	for _, account := range accessibleAccounts {
		accessibleIDs[account.ID] = struct{}{}
	}

	seen := make(map[string]struct{}, len(accountIDs))
	for _, accountID := range accountIDs {
		if _, ok := accessibleIDs[accountID]; !ok {
			return errors.New("account is not accessible")
		}
		if _, duplicate := seen[accountID]; duplicate {
			return errors.New("duplicate account id in order payload")
		}
		seen[accountID] = struct{}{}
	}

	payload, err := json.Marshal(accountIDs)
	if err != nil {
		return err
	}

	return s.db.Model(&models.User{}).
		Where("id = ? AND deleted_at IS NULL", user.ID).
		Update("mobile_accounts_order", string(payload)).Error
}
