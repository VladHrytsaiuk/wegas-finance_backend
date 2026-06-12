package services

import (
	"io"
	"mime/multipart"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockAccountService
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) Create(input CreateAccountInput, user *models.User) (*models.Account, error) {
	args := m.Called(input, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountService) GetAll(user *models.User) ([]models.Account, error) {
	args := m.Called(user)
	return args.Get(0).([]models.Account), args.Error(1)
}

func (m *MockAccountService) GetByID(id string, user *models.User) (*models.Account, error) {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountService) Update(id string, input CreateAccountInput, user *models.User) (*models.Account, error) {
	args := m.Called(id, input, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountService) Delete(id string, user *models.User) error {
	args := m.Called(id, user)
	return args.Error(0)
}

// MockAssetRepository
type MockAssetRepository struct {
	mock.Mock
}

func (m *MockAssetRepository) Create(asset *models.Asset) error {
	args := m.Called(asset)
	return args.Error(0)
}

func (m *MockAssetRepository) GetAll(familyID string) ([]models.Asset, error) {
	args := m.Called(familyID)
	return args.Get(0).([]models.Asset), args.Error(1)
}

func (m *MockAssetRepository) GetByID(id string, familyID string) (*models.Asset, error) {
	args := m.Called(id, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Asset), args.Error(1)
}

func (m *MockAssetRepository) Update(asset *models.Asset) error {
	args := m.Called(asset)
	return args.Error(0)
}

func (m *MockAssetRepository) Delete(id string, familyID string) error {
	args := m.Called(id, familyID)
	return args.Error(0)
}

func (m *MockAssetRepository) RemovePhoto(assetID string, path string) error {
	args := m.Called(assetID, path)
	return args.Error(0)
}

func (m *MockAssetRepository) UpdatePhoto(assetID string, path string) error {
	args := m.Called(assetID, path)
	return args.Error(0)
}

func (m *MockAssetRepository) AddPhotoToGallery(photo *models.AssetPhoto) error {
	args := m.Called(photo)
	return args.Error(0)
}

func (m *MockAssetRepository) AddDocument(doc *models.AssetDocument) error {
	args := m.Called(doc)
	return args.Error(0)
}

func (m *MockAssetRepository) RemoveDocument(assetID string, documentID string) error {
	args := m.Called(assetID, documentID)
	return args.Error(0)
}

func (m *MockAssetRepository) GetDocumentByID(documentID string) (*models.AssetDocument, error) {
	args := m.Called(documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AssetDocument), args.Error(1)
}

func (m *MockAssetRepository) UpdateServiceData(assetID string, mileage int, date int64) error {
	args := m.Called(assetID, mileage, date)
	return args.Error(0)
}

// MockTransactionRepository
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(tx *models.Transaction, items []models.TransactionItem, tagIDs []string, newAsset *models.Asset) error {
	args := m.Called(tx, items, tagIDs, newAsset)
	return args.Error(0)
}

func (m *MockTransactionRepository) CreateTransfer(from, to *models.Transaction) error {
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

func (m *MockTransactionRepository) Update(id string, familyID string, tx *models.Transaction, items []models.TransactionItem, tagIDs []string) error {
	args := m.Called(id, familyID, tx, items, tagIDs)
	return args.Error(0)
}

func (m *MockTransactionRepository) Delete(tx, relatedTx *models.Transaction) error {
	args := m.Called(tx, relatedTx)
	return args.Error(0)
}

func (m *MockTransactionRepository) BatchCreate(txs []models.Transaction) (int, error) {
	args := m.Called(txs)
	return args.Int(0), args.Error(1)
}

func (m *MockTransactionRepository) UpdateReceiptImage(txID string, path string) error {
	args := m.Called(txID, path)
	return args.Error(0)
}

func (m *MockTransactionRepository) DeleteAllPhotos(txID string) error {
	args := m.Called(txID)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetPhotoByID(id string) (*models.TransactionPhoto, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransactionPhoto), args.Error(1)
}

func (m *MockTransactionRepository) DeletePhotoByID(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetPredictedCategory(familyID, itemName string) (string, error) {
	args := m.Called(familyID, itemName)
	return args.String(0), args.Error(1)
}

// MockCounterpartyRepository
type MockCounterpartyRepository struct {
	mock.Mock
}

func (m *MockCounterpartyRepository) Create(cp *models.Counterparty) error {
	args := m.Called(cp)
	return args.Error(0)
}

func (m *MockCounterpartyRepository) GetAll(familyID string) ([]models.Counterparty, error) {
	args := m.Called(familyID)
	return args.Get(0).([]models.Counterparty), args.Error(1)
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

func (m *MockCounterpartyRepository) Update(cp *models.Counterparty) error {
	args := m.Called(cp)
	return args.Error(0)
}

func (m *MockCounterpartyRepository) Delete(id string, familyID string) error {
	args := m.Called(id, familyID)
	return args.Error(0)
}

func (m *MockCounterpartyRepository) CreateCategory(cat *models.CounterpartyCategory) error {
	args := m.Called(cat)
	return args.Error(0)
}

func (m *MockCounterpartyRepository) GetCategories(familyID string) ([]models.CounterpartyCategory, error) {
	args := m.Called(familyID)
	return args.Get(0).([]models.CounterpartyCategory), args.Error(1)
}

func (m *MockCounterpartyRepository) GetCategoryByID(id string, familyID string) (*models.CounterpartyCategory, error) {
	args := m.Called(id, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CounterpartyCategory), args.Error(1)
}

func (m *MockCounterpartyRepository) UpdateCategory(cat *models.CounterpartyCategory) error {
	args := m.Called(cat)
	return args.Error(0)
}

// MockStorageService
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) SaveImage(file *multipart.FileHeader, folder string) (string, error) {
	args := m.Called(file, folder)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) SaveFile(file *multipart.FileHeader, folder string) (string, error) {
	args := m.Called(file, folder)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) DeleteFile(path string) error {
	args := m.Called(path)
	return args.Error(0)
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

// MockStatsService
type MockStatsService struct {
	mock.Mock
}

func (m *MockStatsService) GetDashboardData(user *models.User, targetCurrency string, from, to int64, accountIDs []string) (*DashboardData, error) {
	args := m.Called(user, targetCurrency, from, to, accountIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DashboardData), args.Error(1)
}

func (m *MockStatsService) GetTopStats(user *models.User, flowType, entityType, targetCurrency string, from, to int64, accountIDs []string) ([]repositories.TopStat, error) {
	args := m.Called(user, flowType, entityType, targetCurrency, from, to, accountIDs)
	return args.Get(0).([]repositories.TopStat), args.Error(1)
}

func (m *MockStatsService) GetTrendStats(user *models.User, flowType, targetCurrency string, from, to int64, accountIDs []string) ([]repositories.TrendStat, error) {
	args := m.Called(user, flowType, targetCurrency, from, to, accountIDs)
	return args.Get(0).([]repositories.TrendStat), args.Error(1)
}

func (m *MockStatsService) GetRecentTransactions(user *models.User, accountIDs []string) ([]models.Transaction, error) {
	args := m.Called(user, accountIDs)
	return args.Get(0).([]models.Transaction), args.Error(1)
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

func (m *MockTransactionService) Update(id string, input CreateTransactionInput, user *models.User) error {
	args := m.Called(id, input, user)
	return args.Error(0)
}

func (m *MockTransactionService) Delete(id string, user *models.User) error {
	args := m.Called(id, user)
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

// MockGoalService
type MockGoalService struct {
	mock.Mock
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

func (m *MockGoalService) UploadGoalPhoto(goalID string, userID string, fileReader io.Reader) (string, error) {
	args := m.Called(goalID, userID, fileReader)
	return args.String(0), args.Error(1)
}

func (m *MockGoalService) CheckOverdueGoals() error {
	args := m.Called()
	return args.Error(0)
}

// MockGoalRepository
type MockGoalRepository struct {
	mock.Mock
}

func (m *MockGoalRepository) Create(goal *models.Goal) error {
	args := m.Called(goal)
	return args.Error(0)
}

func (m *MockGoalRepository) FindAllByFamily(familyID string, userID string) ([]models.Goal, error) {
	args := m.Called(familyID, userID)
	return args.Get(0).([]models.Goal), args.Error(1)
}

func (m *MockGoalRepository) FindOne(id string) (*models.Goal, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Goal), args.Error(1)
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

func (m *MockGoalRepository) GetDB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

// MockAccountRepository
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Create(acc *models.Account) error {
	args := m.Called(acc)
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

func (m *MockAccountRepository) Update(acc *models.Account) error {
	args := m.Called(acc)
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

// MockCategoryRepository
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(cat *models.Category) error {
	args := m.Called(cat)
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

func (m *MockCategoryRepository) Update(cat *models.Category) error {
	args := m.Called(cat)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(id string, familyID string) error {
	args := m.Called(id, familyID)
	return args.Error(0)
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

// MockUserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetMe(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetFamilyMembers(user *models.User) ([]models.User, error) {
	args := m.Called(user)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserService) AddMember(actor *models.User, input CreateUserInput) (*models.User, error) {
	args := m.Called(actor, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateProfile(id string, name string, email string) (*models.User, error) {
	args := m.Called(id, name, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ChangePassword(id string, oldPwd, newPwd string) error {
	args := m.Called(id, oldPwd, newPwd)
	return args.Error(0)
}

func (m *MockUserService) DeleteMember(actor *models.User, targetID string) error {
	args := m.Called(actor, targetID)
	return args.Error(0)
}

func (m *MockUserService) UpdateUser(actor *models.User, targetID string, input CreateUserInput) (*models.User, error) {
	args := m.Called(actor, targetID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) LeaveFamily(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// MockFamilyJoinService
type MockFamilyJoinService struct {
	mock.Mock
}

func (m *MockFamilyJoinService) GenerateCode(familyID string, roleID string) (string, error) {
	args := m.Called(familyID, roleID)
	return args.String(0), args.Error(1)
}

func (m *MockFamilyJoinService) JoinFamily(userID string, code string) (*models.Family, error) {
	args := m.Called(userID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Family), args.Error(1)
}

// MockAssetService
type MockAssetService struct {
	mock.Mock
}

func (m *MockAssetService) GetAll(user *models.User) ([]models.Asset, error) {
	args := m.Called(user)
	return args.Get(0).([]models.Asset), args.Error(1)
}

func (m *MockAssetService) Create(input models.Asset, user *models.User) (string, error) {
	args := m.Called(input, user)
	return args.String(0), args.Error(1)
}

func (m *MockAssetService) GetByID(id string, user *models.User) (*AssetWithStats, error) {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AssetWithStats), args.Error(1)
}

func (m *MockAssetService) Update(id string, input models.Asset, user *models.User) error {
	args := m.Called(id, input, user)
	return args.Error(0)
}

func (m *MockAssetService) UpdateMileage(id string, mileage int, user *models.User) error {
	args := m.Called(id, mileage, user)
	return args.Error(0)
}

func (m *MockAssetService) Delete(id string, user *models.User) error {
	args := m.Called(id, user)
	return args.Error(0)
}

func (m *MockAssetService) UploadPhoto(id string, file multipart.File, header *multipart.FileHeader, user *models.User) (string, error) {
	args := m.Called(id, file, header, user)
	return args.String(0), args.Error(1)
}

func (m *MockAssetService) RemovePhoto(id string, photoPath string, user *models.User) error {
	args := m.Called(id, photoPath, user)
	return args.Error(0)
}

func (m *MockAssetService) UploadDocument(id string, file multipart.File, header *multipart.FileHeader, user *models.User) (*models.AssetDocument, error) {
	args := m.Called(id, file, header, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AssetDocument), args.Error(1)
}

func (m *MockAssetService) RemoveDocument(id string, docID string, user *models.User) error {
	args := m.Called(id, docID, user)
	return args.Error(0)
}

// MockShoppingService
type MockShoppingService struct {
	mock.Mock
}

func (m *MockShoppingService) GetLists(familyID, userID string) ([]models.ShoppingList, error) {
	args := m.Called(familyID, userID)
	return args.Get(0).([]models.ShoppingList), args.Error(1)
}

func (m *MockShoppingService) CreateList(req models.CreateShoppingListRequest, userID, familyID string) (*models.ShoppingList, error) {
	args := m.Called(req, userID, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShoppingList), args.Error(1)
}

func (m *MockShoppingService) UpdateList(id string, req models.UpdateShoppingListRequest, familyID string) error {
	args := m.Called(id, req, familyID)
	return args.Error(0)
}

func (m *MockShoppingService) DeleteList(id, familyID string) error {
	args := m.Called(id, familyID)
	return args.Error(0)
}

func (m *MockShoppingService) AddItemToList(listID string, req models.CreateShoppingItemRequest) (*models.ShoppingItem, error) {
	args := m.Called(listID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShoppingItem), args.Error(1)
}

func (m *MockShoppingService) UpdateItem(id string, req models.UpdateShoppingItemRequest) error {
	args := m.Called(id, req)
	return args.Error(0)
}

func (m *MockShoppingService) DeleteItem(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockShoppingService) ClearCompletedInList(listID string) error {
	args := m.Called(listID)
	return args.Error(0)
}

// MockWishlistService
type MockWishlistService struct {
	mock.Mock
}

func (m *MockWishlistService) CreateGroup(name, color, icon, visibility, hiddenFrom, userID, familyID string) (*models.WishlistGroup, error) {
	args := m.Called(name, color, icon, visibility, hiddenFrom, userID, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WishlistGroup), args.Error(1)
}

func (m *MockWishlistService) GetGroups(familyID, myUserID string) ([]models.WishlistGroup, error) {
	args := m.Called(familyID, myUserID)
	return args.Get(0).([]models.WishlistGroup), args.Error(1)
}

func (m *MockWishlistService) UpdateGroup(id string, req models.UpdateWishlistGroupRequest, familyID string) error {
	args := m.Called(id, req, familyID)
	return args.Error(0)
}

func (m *MockWishlistService) DeleteGroup(id, familyID string) error {
	args := m.Called(id, familyID)
	return args.Error(0)
}

func (m *MockWishlistService) CreateItem(req models.CreateWishlistRequest, userID, familyID string) (*models.WishlistItem, error) {
	args := m.Called(req, userID, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WishlistItem), args.Error(1)
}

func (m *MockWishlistService) ToggleReservation(itemID string, userID string) error {
	args := m.Called(itemID, userID)
	return args.Error(0)
}

func (m *MockWishlistService) GetItems(familyID, userID, groupID, targetUserID string) ([]models.WishlistItem, error) {
	args := m.Called(familyID, userID, groupID, targetUserID)
	return args.Get(0).([]models.WishlistItem), args.Error(1)
}

func (m *MockWishlistService) GetItem(id string, familyID string) (*models.WishlistItem, error) {
	args := m.Called(id, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WishlistItem), args.Error(1)
}

func (m *MockWishlistService) UploadPhoto(id string, familyID string, fileReader io.Reader) (string, error) {
	args := m.Called(id, familyID, fileReader)
	return args.String(0), args.Error(1)
}

func (m *MockWishlistService) UpdateItem(id string, req models.UpdateWishlistRequest, familyID string) (*models.WishlistItem, error) {
	args := m.Called(id, req, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WishlistItem), args.Error(1)
}

func (m *MockWishlistService) DeleteItem(id string, familyID, userID string) error {
	args := m.Called(id, familyID, userID)
	return args.Error(0)
}

func (m *MockWishlistService) RemovePhoto(id string, familyID string) error {
	args := m.Called(id, familyID)
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

// MockExportService
type MockExportService struct {
	mock.Mock
}

func (m *MockExportService) GetTransactions(user *models.User, filter models.ExportFilterDTO) ([]models.Transaction, error) {
	args := m.Called(user, filter)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

// MockFeedbackService
type MockFeedbackService struct {
	mock.Mock
}

func (m *MockFeedbackService) SendFeedback(name, contact, message, priority string, images [][]byte) error {
	args := m.Called(name, contact, message, priority, images)
	return args.Error(0)
}

// MockImportService
type MockImportService struct {
	mock.Mock
}

func (m *MockImportService) ProcessFile(file *multipart.FileHeader, accountID string, bankType string) (*PreviewResult, error) {
	args := m.Called(file, accountID, bankType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PreviewResult), args.Error(1)
}

// MockRoleService
type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) Create(input CreateRoleInput) (*models.Role, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleService) GetAll() ([]models.Role, error) {
	args := m.Called()
	return args.Get(0).([]models.Role), args.Error(1)
}

func (m *MockRoleService) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockUserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetFamilyMembers(familyID string) ([]models.User, error) {
	args := m.Called(familyID)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) CountFamilyMembers(familyID string) (int64, error) {
	args := m.Called(familyID)
	return args.Get(1).(int64), args.Error(1)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) CreateFamily(family *models.Family) error {
	args := m.Called(family)
	return args.Error(0)
}

func (m *MockUserRepository) GetFamilyByID(id string) (*models.Family, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Family), args.Error(1)
}

func (m *MockUserRepository) DeleteFamily(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) GetDB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

// MockStorageTypeService
type MockStorageTypeService struct {
	mock.Mock
}

func (m *MockStorageTypeService) Create(st *models.StorageType) error {
	args := m.Called(st)
	return args.Error(0)
}

func (m *MockStorageTypeService) GetAll(familyID string) ([]models.StorageType, error) {
	args := m.Called(familyID)
	return args.Get(0).([]models.StorageType), args.Error(1)
}

func (m *MockStorageTypeService) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockStorageTypeRepository
type MockStorageTypeRepository struct {
	mock.Mock
}

func (m *MockStorageTypeRepository) Create(st *models.StorageType) error {
	args := m.Called(st)
	return args.Error(0)
}

func (m *MockStorageTypeRepository) FindAvailable(familyID string) ([]models.StorageType, error) {
	args := m.Called(familyID)
	return args.Get(0).([]models.StorageType), args.Error(1)
}

func (m *MockStorageTypeRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStorageTypeRepository) FindBySlug(slug string) (*models.StorageType, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.StorageType), args.Error(1)
}

// MockTagService
type MockTagService struct {
	mock.Mock
}

func (m *MockTagService) Create(name string, color string, user *models.User) (*models.Tag, error) {
	args := m.Called(name, color, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tag), args.Error(1)
}

func (m *MockTagService) GetAll(familyID string) ([]models.Tag, error) {
	args := m.Called(familyID)
	return args.Get(0).([]models.Tag), args.Error(1)
}

func (m *MockTagService) Delete(id string, user *models.User) error {
	args := m.Called(id, user)
	return args.Error(0)
}

// MockUtilityService
type MockUtilityService struct {
	mock.Mock
}

func (m *MockUtilityService) CreateMeter(input models.UtilityMeter, user *models.User) error {
	args := m.Called(input, user)
	return args.Error(0)
}

func (m *MockUtilityService) GetMeters(user *models.User) ([]models.UtilityMeter, error) {
	args := m.Called(user)
	return args.Get(0).([]models.UtilityMeter), args.Error(1)
}

func (m *MockUtilityService) GetMeterByID(id string, user *models.User) (*models.UtilityMeter, error) {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UtilityMeter), args.Error(1)
}

func (m *MockUtilityService) UpdateMeter(id string, input models.UtilityMeter, user *models.User) error {
	args := m.Called(id, input, user)
	return args.Error(0)
}

func (m *MockUtilityService) DeleteMeter(id string, user *models.User) error {
	args := m.Called(id, user)
	return args.Error(0)
}

func (m *MockUtilityService) CreateReading(input models.UtilityReading, user *models.User) error {
	args := m.Called(input, user)
	return args.Error(0)
}

func (m *MockUtilityService) GetReadings(user *models.User, meterID string) ([]models.UtilityReading, error) {
	args := m.Called(user, meterID)
	return args.Get(0).([]models.UtilityReading), args.Error(1)
}

func (m *MockUtilityService) UpdateReading(id string, input models.UtilityReading, user *models.User) error {
	args := m.Called(id, input, user)
	return args.Error(0)
}

func (m *MockUtilityService) DeleteReading(id string, user *models.User) error {
	args := m.Called(id, user)
	return args.Error(0)
}

func (m *MockUtilityService) PayReading(readingID string, accountID string, user *models.User) error {
	args := m.Called(readingID, accountID, user)
	return args.Error(0)
}

func (m *MockUtilityService) GetGlobalStats(user *models.User) ([]models.UtilityStatGlobalDTO, error) {
	args := m.Called(user)
	return args.Get(0).([]models.UtilityStatGlobalDTO), args.Error(1)
}

func (m *MockUtilityService) GetMeterStats(meterID string, user *models.User) ([]models.UtilityStatMeterDTO, error) {
	args := m.Called(meterID, user)
	return args.Get(0).([]models.UtilityStatMeterDTO), args.Error(1)
}
