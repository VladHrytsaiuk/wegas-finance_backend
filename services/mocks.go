package services

import (
	"io"
	"mime/multipart"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockTransactionRepository
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(t *models.Transaction, items []models.TransactionItem, tagIDs []string, newAsset *models.Asset) error {
	args := m.Called(t, items, tagIDs, newAsset)
	return args.Error(0)
}

func (m *MockTransactionRepository) CreateTransfer(from *models.Transaction, to *models.Transaction) error {
	args := m.Called(from, to)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetAll(filter repositories.TransactionFilter) ([]models.Transaction, int64, error) {
	args := m.Called(filter)
	return args.Get(0).([]models.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionRepository) GetByID(id string, familyID string) (*models.Transaction, error) {
	args := m.Called(id, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) Update(id string, familyID string, t *models.Transaction, items []models.TransactionItem, tagIDs []string) error {
	args := m.Called(id, familyID, t, items, tagIDs)
	return args.Error(0)
}

func (m *MockTransactionRepository) UpdateReceiptImage(id string, imagePath string) error {
	args := m.Called(id, imagePath)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetPhotoByID(photoID string) (*models.TransactionPhoto, error) {
	args := m.Called(photoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransactionPhoto), args.Error(1)
}

func (m *MockTransactionRepository) DeletePhotoByID(photoID string) error {
	args := m.Called(photoID)
	return args.Error(0)
}

func (m *MockTransactionRepository) DeleteAllPhotos(txID string) error {
	args := m.Called(txID)
	return args.Error(0)
}

func (m *MockTransactionRepository) Delete(tx *models.Transaction, relatedTx *models.Transaction) error {
	args := m.Called(tx, relatedTx)
	return args.Error(0)
}

func (m *MockTransactionRepository) BatchCreate(transactions []models.Transaction) (int, error) {
	args := m.Called(transactions)
	return args.Int(0), args.Error(1)
}

func (m *MockTransactionRepository) GetPredictedCategory(familyID string, itemName string) (string, error) {
	args := m.Called(familyID, itemName)
	return args.String(0), args.Error(1)
}

// MockAccountRepository
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Create(account *models.Account) error {
	args := m.Called(account)
	return args.Error(0)
}

func (m *MockAccountRepository) GetAllByFamilyID(familyID string) ([]models.Account, error) {
	args := m.Called(familyID)
	return args.Get(0).([]models.Account), args.Error(1)
}

func (m *MockAccountRepository) GetAllByUserID(userID string) ([]models.Account, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByID(id string) (*models.Account, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountRepository) Update(account *models.Account) error {
	args := m.Called(account)
	return args.Error(0)
}

func (m *MockAccountRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAccountRepository) GetByExternalID(externalID string) (*models.Account, error) {
	args := m.Called(externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

// MockCounterpartyRepository
type MockCounterpartyRepository struct {
	mock.Mock
}

func (m *MockCounterpartyRepository) CreateCategory(category *models.CounterpartyCategory) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockCounterpartyRepository) GetCategories(familyID string) ([]models.CounterpartyCategory, error) {
	args := m.Called(familyID)
	return args.Get(0).([]models.CounterpartyCategory), args.Error(1)
}

func (m *MockCounterpartyRepository) UpdateCategory(category *models.CounterpartyCategory) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockCounterpartyRepository) GetCategoryByID(id string, familyID string) (*models.CounterpartyCategory, error) {
	args := m.Called(id, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CounterpartyCategory), args.Error(1)
}

func (m *MockCounterpartyRepository) Create(cp *models.Counterparty) error {
	args := m.Called(cp)
	return args.Error(0)
}

func (m *MockCounterpartyRepository) GetAll(familyID string) ([]models.Counterparty, error) {
	args := m.Called(familyID)
	return args.Get(0).([]models.Counterparty), args.Error(1)
}

func (m *MockCounterpartyRepository) Update(cp *models.Counterparty) error {
	args := m.Called(cp)
	return args.Error(0)
}

func (m *MockCounterpartyRepository) Delete(id string, familyID string) error {
	args := m.Called(id, familyID)
	return args.Error(0)
}

func (m *MockCounterpartyRepository) GetByID(id string, familyID string) (*models.Counterparty, error) {
	args := m.Called(id, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Counterparty), args.Error(1)
}

func (m *MockCounterpartyRepository) GetByName(name string, familyID string) (*models.Counterparty, error) {
	args := m.Called(name, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Counterparty), args.Error(1)
}

// MockStorageService
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) SaveImage(fileHeader *multipart.FileHeader, folder string) (string, error) {
	args := m.Called(fileHeader, folder)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) SaveFile(fileHeader *multipart.FileHeader, folder string) (string, error) {
	args := m.Called(fileHeader, folder)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) DeleteFile(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

// MockTransactionService
type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) Create(input CreateTransactionInput, files []*multipart.FileHeader, user *models.User) (string, error) {
	args := m.Called(input, files, user)
	return args.String(0), args.Error(1)
}

func (m *MockTransactionService) GetAll(filter repositories.TransactionFilter, user *models.User) ([]models.Transaction, int64, error) {
	args := m.Called(filter, user)
	return args.Get(0).([]models.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionService) GetByID(id string, user *models.User) (*models.Transaction, error) {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionService) Delete(id string, user *models.User) error {
	args := m.Called(id, user)
	return args.Error(0)
}

func (m *MockTransactionService) Update(id string, input CreateTransactionInput, user *models.User) error {
	args := m.Called(id, input, user)
	return args.Error(0)
}

func (m *MockTransactionService) UploadReceipt(txID string, file multipart.File, header *multipart.FileHeader, user *models.User) (string, error) {
	args := m.Called(txID, file, header, user)
	return args.String(0), args.Error(1)
}

func (m *MockTransactionService) DeleteReceipt(txID string, user *models.User) error {
	args := m.Called(txID, user)
	return args.Error(0)
}

func (m *MockTransactionService) DeletePhoto(photoID string, user *models.User) error {
	args := m.Called(photoID, user)
	return args.Error(0)
}

func (m *MockTransactionService) BatchCreate(inputs []CreateTransactionInput, user *models.User) (int, error) {
	args := m.Called(inputs, user)
	return args.Int(0), args.Error(1)
}

func (m *MockTransactionService) PredictCategory(itemName string, user *models.User) (string, error) {
	args := m.Called(itemName, user)
	return args.String(0), args.Error(1)
}

// MockAuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(input RegisterInput) (*LoginResponse, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LoginResponse), args.Error(1)
}

func (m *MockAuthService) Login(input LoginInput) (*LoginResponse, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LoginResponse), args.Error(1)
}

// MockMonobankService
type MockMonobankService struct {
	mock.Mock
}

func (m *MockMonobankService) Connect(userID, familyID, rawToken string) ([]MonoAccount, error) {
	args := m.Called(userID, familyID, rawToken)
	return args.Get(0).([]MonoAccount), args.Error(1)
}

func (m *MockMonobankService) GetUserData(userID string) ([]MonoAccount, []models.BankAccountMapping, error) {
	args := m.Called(userID)
	return args.Get(0).([]MonoAccount), args.Get(1).([]models.BankAccountMapping), args.Error(2)
}

func (m *MockMonobankService) RefreshClientInfo(userID string) ([]MonoAccount, []models.BankAccountMapping, error) {
	args := m.Called(userID)
	return args.Get(0).([]MonoAccount), args.Get(1).([]models.BankAccountMapping), args.Error(2)
}

func (m *MockMonobankService) SaveSettings(userID, familyID string, accounts []models.BankAccountMapping) error {
	args := m.Called(userID, familyID, accounts)
	return args.Error(0)
}

func (m *MockMonobankService) Sync(userID string, targetAccountID string) (int, error) {
	args := m.Called(userID, targetAccountID)
	return args.Int(0), args.Error(1)
}

func (m *MockMonobankService) Disconnect(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockMonobankService) GetSyncStatus(userID string) SyncStatus {
	args := m.Called(userID)
	return args.Get(0).(SyncStatus)
}

func (m *MockMonobankService) GlobalResyncCounterparties() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *MockMonobankService) ProcessWebhook(payload MonoWebhookPayload) error {
	args := m.Called(payload)
	return args.Error(0)
}

// MockCategoryRepository
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetAll(familyID string) ([]models.Category, error) {
	args := m.Called(familyID)
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByID(id string, familyID string) (*models.Category, error) {
	args := m.Called(id, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(id string, familyID string) error {
	args := m.Called(id, familyID)
	return args.Error(0)
}

// MockGoalRepository
type MockGoalRepository struct {
	mock.Mock
}

func (m *MockGoalRepository) GetDB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockGoalRepository) Create(goal *models.Goal) error {
	args := m.Called(goal)
	return args.Error(0)
}

func (m *MockGoalRepository) FindOne(id string) (*models.Goal, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Goal), args.Error(1)
}

func (m *MockGoalRepository) FindAllByFamily(familyID string, userID string) ([]models.Goal, error) {
	args := m.Called(familyID, userID)
	return args.Get(0).([]models.Goal), args.Error(1)
}

func (m *MockGoalRepository) Update(goal *models.Goal) error {
	args := m.Called(goal)
	return args.Error(0)
}

func (m *MockGoalRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockGoalRepository) FindAllActive() ([]models.Goal, error) {
	args := m.Called()
	return args.Get(0).([]models.Goal), args.Error(1)
}

// MockCategoryService
type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) Create(input CategoryInput, user *models.User) (*models.Category, error) {
	args := m.Called(input, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryService) GetAll(user *models.User) ([]models.Category, error) {
	args := m.Called(user)
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryService) GetByID(id string, user *models.User) (*models.Category, error) {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryService) Update(id string, input CategoryInput, user *models.User) (*models.Category, error) {
	args := m.Called(id, input, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryService) Delete(id string, user *models.User) error {
	args := m.Called(id, user)
	return args.Error(0)
}

// MockGoalService
type MockGoalService struct {
	mock.Mock
}

func (m *MockGoalService) UploadGoalPhoto(goalID string, userID string, fileReader io.Reader) (string, error) {
	args := m.Called(goalID, userID, fileReader)
	return args.String(0), args.Error(1)
}

func (m *MockGoalService) Create(goal *models.Goal, userID string) error {
	args := m.Called(goal, userID)
	return args.Error(0)
}

func (m *MockGoalService) GetAll(familyID string, userID string) ([]models.Goal, error) {
	args := m.Called(familyID, userID)
	return args.Get(0).([]models.Goal), args.Error(1)
}

func (m *MockGoalService) GetOne(id string, userID string) (*models.Goal, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Goal), args.Error(1)
}

func (m *MockGoalService) Update(incomingGoal *models.Goal, userID string) error {
	args := m.Called(incomingGoal, userID)
	return args.Error(0)
}

func (m *MockGoalService) Delete(id string, userID string) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *MockGoalService) LinkAccount(goalID string, accountID string) error {
	args := m.Called(goalID, accountID)
	return args.Error(0)
}

func (m *MockGoalService) UnlinkAccount(accountID string) error {
	args := m.Called(accountID)
	return args.Error(0)
}

func (m *MockGoalService) CheckOverdueGoals() error {
	args := m.Called()
	return args.Error(0)
}

// MockCounterpartyService
type MockCounterpartyService struct {
	mock.Mock
}

func (m *MockCounterpartyService) CreateCategory(input CpCategoryInput, user *models.User) (*models.CounterpartyCategory, error) {
	args := m.Called(input, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CounterpartyCategory), args.Error(1)
}

func (m *MockCounterpartyService) GetCategories(user *models.User) ([]models.CounterpartyCategory, error) {
	args := m.Called(user)
	return args.Get(0).([]models.CounterpartyCategory), args.Error(1)
}

func (m *MockCounterpartyService) UpdateCategory(id string, input CpCategoryInput, user *models.User) (*models.CounterpartyCategory, error) {
	args := m.Called(id, input, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CounterpartyCategory), args.Error(1)
}

func (m *MockCounterpartyService) GetCategoryByID(id string, user *models.User) (*models.CounterpartyCategory, error) {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CounterpartyCategory), args.Error(1)
}

func (m *MockCounterpartyService) Create(input CounterpartyInput, user *models.User) (*models.Counterparty, error) {
	args := m.Called(input, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Counterparty), args.Error(1)
}

func (m *MockCounterpartyService) GetAll(user *models.User) ([]models.Counterparty, error) {
	args := m.Called(user)
	return args.Get(0).([]models.Counterparty), args.Error(1)
}

func (m *MockCounterpartyService) Update(id string, input CounterpartyInput, user *models.User) (*models.Counterparty, error) {
	args := m.Called(id, input, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Counterparty), args.Error(1)
}

func (m *MockCounterpartyService) Delete(id string, user *models.User) error {
	args := m.Called(id, user)
	return args.Error(0)
}

func (m *MockCounterpartyService) GetByID(id string, user *models.User) (*models.Counterparty, error) {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Counterparty), args.Error(1)
}

// MockCurrencyService
type MockCurrencyService struct {
	mock.Mock
}

func (m *MockCurrencyService) SyncRates() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCurrencyService) Convert(amount int64, fromCurrency, toCurrency string) (int64, error) {
	args := m.Called(amount, fromCurrency, toCurrency)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCurrencyService) GetAllRates() ([]models.ExchangeRate, error) {
	args := m.Called()
	return args.Get(0).([]models.ExchangeRate), args.Error(1)
}
