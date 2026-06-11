package services

import (
	"errors"
	"image"
	"image/jpeg"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 🔥 ВИДАЛЕНО CreateAssetOnFlyInput (вона тепер у models)

type TransactionItemInput struct {
	Name         string  `json:"name"`
	Quantity     int64   `json:"quantity"`
	PricePerUnit int64   `json:"price_per_unit"`
	TotalAmount  int64   `json:"total_amount"`
	CategoryID   *string `json:"category_id"`
	
}

type CreateTransactionInput struct {
	AccountID       string
	TargetAccountID string
	CategoryID      string

	CounterpartyID   string
	CounterpartyName string

	Amount       int64
	TargetAmount *int64

	Date int64
	Note string
	Type string // 'expense', 'income', 'loan_give', 'loan_repay', 'debt_take', 'debt_repay'

	Items  []TransactionItemInput
	TagIDs []string

	AssetID *string
	// 🔥 ПОСИЛАННЯ НА MODELS
	NewAsset      *models.CreateAssetOnFlyInput
	Mileage *int `json:"mileage"`
	IsForgiveness bool
	ExternalID    string
}

type TransactionService interface {
	Create(input CreateTransactionInput, files []*multipart.FileHeader, user *models.User) (string, error)
	GetAll(filter repositories.TransactionFilter, user *models.User) ([]models.Transaction, int64, error)
	GetByID(id string, user *models.User) (*models.Transaction, error)
	Delete(id string, user *models.User) error
	Update(id string, input CreateTransactionInput, user *models.User) error

	UploadReceipt(txID string, file multipart.File, header *multipart.FileHeader, user *models.User) (string, error)
	DeleteReceipt(txID string, user *models.User) error
	DeletePhoto(photoID string, user *models.User) error

	BatchCreate(inputs []CreateTransactionInput, user *models.User) (int, error)
	PredictCategory(itemName string, user *models.User) (string, error)
}

type txService struct {
	db     *gorm.DB
	repo   repositories.TransactionRepository
	cpRepo repositories.CounterpartyRepository
	assetRepo repositories.AssetRepository
}

func NewTransactionService(db *gorm.DB, repo repositories.TransactionRepository, cpRepo repositories.CounterpartyRepository, assetRepo repositories.AssetRepository,) TransactionService {
	return &txService{
		db:     db,
		repo:   repo,
		cpRepo: cpRepo,
		assetRepo: assetRepo,
	}
}
func (s *txService) Create(input CreateTransactionInput, files []*multipart.FileHeader, user *models.User) (string, error) {
	now := time.Now().UnixMilli()
	userID := user.ID

	// === ЛОГІКА ПЕРЕКАЗУ ===
	if input.Type == "transfer" {
		if input.TargetAccountID == "" {
			return "", errors.New("target account required for transfer")
		}
		txID := uuid.NewString()
		relatedTxID := uuid.NewString()

		incomingAmount := abs(input.Amount)
		if input.TargetAmount != nil && *input.TargetAmount > 0 {
			incomingAmount = abs(*input.TargetAmount)
		}

		// Списуємо (OUT)
		fromTx := &models.Transaction{
			Base:              models.Base{ID: txID, CreatedAt: now, UpdatedAt: now, IsSynced: true},
			FamilyID:          user.FamilyID,
			AccountID:         input.AccountID,
			UserID:            userID,
			CategoryID:        input.CategoryID,
			Amount:            abs(input.Amount),
			Date:              input.Date,
			Note:              input.Note,
			Type:              "transfer_out",
			TransferRelatedID: &relatedTxID,
		}

		// Зараховуємо (IN)
		toTx := &models.Transaction{
			Base:              models.Base{ID: relatedTxID, CreatedAt: now, UpdatedAt: now, IsSynced: true},
			FamilyID:          user.FamilyID,
			AccountID:         input.TargetAccountID,
			UserID:            userID,
			CategoryID:        input.CategoryID,
			Amount:            incomingAmount,
			Date:              input.Date,
			Note:              input.Note,
			Type:              "transfer_in",
			TransferRelatedID: &txID,
		}
		return txID, s.repo.CreateTransfer(fromTx, toTx)
	}

	// === СТВОРЕННЯ АКТИВУ "НА ЛЬОТУ" ===
	var newAsset *models.Asset
	// 🔥 ДОЗВОЛЯЄМО СТВОРЕННЯ АКТИВУ ДЛЯ ВИТРАТ ТА БОРГІВ
	if input.NewAsset != nil && (input.Type == "expense" || input.Type == "debt_take") {
		newAsset = &models.Asset{
			Base:         models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
			FamilyID:     user.FamilyID,
			UserID:       userID,
			Name:         input.NewAsset.Name,
			Type:         input.NewAsset.Type,
			SerialNumber: input.NewAsset.SerialNumber,
			PurchaseDate: input.Date,
			Price:        abs(input.Amount),
			CurrentPrice: abs(input.Amount),
			WarrantyEnd:  input.NewAsset.WarrantyEnd,
			Note:         input.NewAsset.Note,
			Currency:     "UAH",
			IsSold:       false,
			Photos:       []models.AssetPhoto{},
			// Ініціалізуємо пробіг для нового авто
			Mileage:        input.NewAsset.Mileage,
			InitialMileage: input.NewAsset.Mileage, // Або input.NewAsset.InitialMileage, якщо передали
		}

		if len(files) > 0 {
			for i, fileHeader := range files {
				path, err := s.saveFile(fileHeader, "assets")
				if err == nil {
					newAsset.Photos = append(newAsset.Photos, models.AssetPhoto{
						Base:    models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
						AssetID: newAsset.ID,
						Path:    path,
					})
					if i == 0 {
						newAsset.Photo = path
					}
				}
			}
		}
	}

	// === ЗВИЧАЙНА ТРАНЗАКЦІЯ ===
	txID := uuid.NewString()
	transaction := &models.Transaction{
		Base:           models.Base{ID: txID, CreatedAt: now, UpdatedAt: now, IsSynced: true},
		FamilyID:       user.FamilyID,
		AccountID:      input.AccountID,
		UserID:         userID,
		CategoryID:     input.CategoryID,
		CounterpartyID: input.CounterpartyID,
		Amount:         abs(input.Amount),
		Date:           input.Date,
		Note:           input.Note,
		Type:           input.Type,
		AssetID:        input.AssetID,
		// 🔥 Зберігаємо пробіг в транзакцію
		Mileage:        input.Mileage, 
		Photos:         []models.TransactionPhoto{},
		IsForgiveness:  input.IsForgiveness,
	}

	if newAsset == nil && len(files) > 0 {
		for i, fileHeader := range files {
			path, err := s.saveFile(fileHeader, "receipts")
			if err == nil {
				transaction.Photos = append(transaction.Photos, models.TransactionPhoto{
					Base:          models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
					TransactionID: txID,
					Path:          path,
				})
				if i == 0 {
					transaction.ReceiptImg = path
				}
			}
		}
	}

	var items []models.TransactionItem
	for _, it := range input.Items {
		items = append(items, models.TransactionItem{
			Base:          models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now},
			TransactionID: txID,
			Name:          it.Name,
			Quantity:      it.Quantity,
			PricePerUnit:  it.PricePerUnit,
			TotalAmount:   it.TotalAmount,
			CategoryID:    it.CategoryID,
		})
	}

	err := s.repo.Create(transaction, items, input.TagIDs, newAsset)
	if err != nil {
		return "", err
	}

	// --- 🚗 ЛОГІКА ОНОВЛЕННЯ АКТИВУ (RATCHET MECHANISM) ---
	if input.AssetID != nil && *input.AssetID != "" {
		
		newMileage := 0
		if input.Mileage != nil {
			newMileage = *input.Mileage
		}

		if newMileage > 0 {
			asset, err := s.assetRepo.GetByID(*input.AssetID, user.FamilyID)
			if err == nil {
				// Оновлюємо ТІЛЬКИ якщо новий пробіг БІЛЬШИЙ за існуючий.
				// Це захищає від випадкових помилок або введення старих чеків.
				if newMileage > asset.Mileage {
					// Оновлюємо і пробіг, і дату останнього сервісу (бо це свіжа інформація)
					_ = s.assetRepo.UpdateServiceData(asset.ID, newMileage, input.Date)
				}
			}
		}
	}

	return txID, nil
}

func (s *txService) Update(id string, input CreateTransactionInput, user *models.User) error {
	oldTx, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return errors.New("transaction not found")
	}
	if user.RoleID == "child" && oldTx.UserID != user.ID {
		return errors.New("access denied")
	}

	// Створюємо об'єкт для оновлення
	updatedTx := &models.Transaction{
		AccountID:      input.AccountID,
		CategoryID:     input.CategoryID,
		CounterpartyID: input.CounterpartyID,
		Amount:         abs(input.Amount),
		Date:           input.Date,
		Note:           input.Note,
		Type:           input.Type,
		ReceiptImg:     oldTx.ReceiptImg,
		AssetID:        input.AssetID,
		IsForgiveness:  input.IsForgiveness,
		
		// 🔥 Не забуваємо оновлювати пробіг в самій транзакції
		Mileage:        input.Mileage, 
	}

	var items []models.TransactionItem
	for _, it := range input.Items {
		items = append(items, models.TransactionItem{
			Name:         it.Name,
			Quantity:     it.Quantity,
			PricePerUnit: it.PricePerUnit,
			TotalAmount:  it.TotalAmount,
			CategoryID:   it.CategoryID,
		})
	}

	// Зберігаємо зміни транзакції
	err = s.repo.Update(id, user.FamilyID, updatedTx, items, input.TagIDs)
	if err != nil {
		return err
	}

	// --- 🚗 ЛОГІКА ОНОВЛЕННЯ АКТИВУ (RATCHET MECHANISM) ---
	if input.AssetID != nil && *input.AssetID != "" {
		
		newMileage := 0
		if input.Mileage != nil {
			newMileage = *input.Mileage
		}

		if newMileage > 0 {
			asset, err := s.assetRepo.GetByID(*input.AssetID, user.FamilyID)
			if err == nil {
				// Та сама логіка: оновлюємо актив, тільки якщо пробіг зростає
				if newMileage > asset.Mileage {
					_ = s.assetRepo.UpdateServiceData(asset.ID, newMileage, input.Date)
				}
			}
		}
	}

	return nil
}

func (s *txService) Delete(id string, user *models.User) error {
	tx, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return errors.New("transaction not found")
	}
	if user.RoleID == "child" && tx.UserID != user.ID {
		return errors.New("access denied")
	}

	s.DeleteReceipt(id, user)

	var relatedTx *models.Transaction
	if strings.Contains(tx.Type, "transfer") && tx.TransferRelatedID != nil {
		rel, err := s.repo.GetByID(*tx.TransferRelatedID, user.FamilyID)
		if err == nil {
			relatedTx = rel
		}
	}
	return s.repo.Delete(tx, relatedTx)
}

func (s *txService) saveFile(fileHeader *multipart.FileHeader, folder string) (string, error) {
  file, err := fileHeader.Open()
  if err != nil {
    return "", err
  }
  defer file.Close()

  uploadDir := "./uploads/" + folder
  if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
    return "", err
  }

  // Оскільки ми примусово зберігаємо як JPEG, відкидаємо старе розширення
  filename := uuid.NewString() + ".jpg"
  filePath := filepath.Join(uploadDir, filename)

  out, err := os.Create(filePath)
  if err != nil {
    return "", err
  }
  defer out.Close()

  // 1. Декодуємо картинку (щоб Go прочитав пікселі)
  img, _, err := image.Decode(file)
  if err != nil {
    return "", errors.New("invalid image format")
  }

  // 2. Зберігаємо у файл зі стисненням якості 80%
  if err := jpeg.Encode(out, img, &jpeg.Options{Quality: 80}); err != nil {
    return "", err
  }

  return "/uploads/" + folder + "/" + filename, nil
}

func (s *txService) GetAll(filter repositories.TransactionFilter, user *models.User) ([]models.Transaction, int64, error) {
	if user.RoleID == "child" {
		filter.UserID = user.ID
	}
	return s.repo.GetAll(filter)
}

func (s *txService) GetByID(id string, user *models.User) (*models.Transaction, error) {
	tx, err := s.repo.GetByID(id, user.FamilyID)
	if err != nil {
		return nil, err
	}
	if user.RoleID == "child" && tx.UserID != user.ID {
		return nil, errors.New("access denied")
	}
	return tx, nil
}



func (s *txService) UploadReceipt(txID string, file multipart.File, header *multipart.FileHeader, user *models.User) (string, error) {
	tx, err := s.repo.GetByID(txID, user.FamilyID)
	if err != nil {
		return "", err
	}
	if user.RoleID == "child" && tx.UserID != user.ID {
		return "", errors.New("access denied")
	}
	path, err := s.saveFile(header, "receipts")
	if err != nil {
		return "", err
	}
	if err := s.repo.UpdateReceiptImage(txID, path); err != nil {
		return "", err
	}
	return path, nil
}

func (s *txService) DeleteReceipt(txID string, user *models.User) error {
	tx, err := s.repo.GetByID(txID, user.FamilyID)
	if err != nil {
		return errors.New("transaction not found")
	}
	if user.RoleID == "child" && tx.UserID != user.ID {
		return errors.New("access denied")
	}
	for _, photo := range tx.Photos {
		if photo.Path != "" {
			_ = os.Remove("." + photo.Path)
		}
	}
	if tx.ReceiptImg != "" {
		_ = os.Remove("." + tx.ReceiptImg)
	}
	return s.repo.DeleteAllPhotos(txID)
}

func (s *txService) DeletePhoto(photoID string, user *models.User) error {
	photo, err := s.repo.GetPhotoByID(photoID)
	if err != nil {
		return errors.New("photo not found")
	}
	tx, err := s.repo.GetByID(photo.TransactionID, user.FamilyID)
	if err != nil {
		return errors.New("transaction access denied or not found")
	}
	if user.RoleID == "child" && tx.UserID != user.ID {
		return errors.New("access denied")
	}
	if photo.Path != "" {
		_ = os.Remove("." + photo.Path)
	}
	return s.repo.DeletePhotoByID(photoID)
}

func (s *txService) BatchCreate(inputs []CreateTransactionInput, user *models.User) (int, error) {
	var txs []models.Transaction
	now := time.Now().UnixMilli()
	cpCache := make(map[string]string)

	for _, input := range inputs {
		txID := uuid.NewString()
		finalCounterpartyID := input.CounterpartyID
		if finalCounterpartyID == "" && input.CounterpartyName != "" {
			name := input.CounterpartyName
			if cachedID, ok := cpCache[name]; ok {
				finalCounterpartyID = cachedID
			} else {
				existingCP, err := s.cpRepo.GetByName(name, user.FamilyID)
				if err == nil && existingCP != nil {
					finalCounterpartyID = existingCP.ID
					cpCache[name] = existingCP.ID
				}
			}
		}

		amount := abs(input.Amount)

		tx := models.Transaction{
			Base:           models.Base{ID: txID, CreatedAt: now, UpdatedAt: now, IsSynced: true},
			FamilyID:       user.FamilyID,
			AccountID:      input.AccountID,
			UserID:         user.ID,
			CategoryID:     input.CategoryID,
			CounterpartyID: finalCounterpartyID,
			Amount:         amount,
			Date:           input.Date,
			Note:           input.Note,
			Type:           input.Type,
			AssetID:        input.AssetID,
			IsForgiveness:  input.IsForgiveness,
			ExternalID:     input.ExternalID,
		}
		txs = append(txs, tx)
	}
	return s.repo.BatchCreate(txs)
}

func (s *txService) PredictCategory(itemName string, user *models.User) (string, error) {
	if itemName == "" {
		return "", nil
	}
	return s.repo.GetPredictedCategory(user.FamilyID, itemName)
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}