package services

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services/parsers"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"gorm.io/gorm"
)

type ImportService struct {
	DB *gorm.DB
}

func NewImportService(db *gorm.DB) *ImportService {
	return &ImportService{DB: db}
}

type PreviewTransaction struct {
	Date                 int64               `json:"date"`
	Amount               int64               `json:"amount"`
	Description          string              `json:"description"`
	CounterpartyName     string              `json:"counterparty_name"`
	CounterpartyID       string              `json:"counterparty_id"`
	Type                 string              `json:"type"`
	PredictedCategory    string              `json:"predicted_category"`
	IsDuplicate          bool                `json:"is_duplicate"`
	IsPotentialDuplicate bool                `json:"is_potential_duplicate"`
	MatchReason          string              `json:"match_reason"`
	ExistingTransaction  *models.Transaction `json:"existing_transaction,omitempty"`
	MCC                  string              `json:"mcc"`
}

type PreviewResult struct {
	Transactions []PreviewTransaction `json:"transactions"`
}

func (s *ImportService) ProcessFile(file *multipart.FileHeader, accountID string, bankType string) (*PreviewResult, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	fileBytes := buf.Bytes()

	if strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
		start := bytes.Index(fileBytes, []byte("%PDF"))
		end := bytes.LastIndex(fileBytes, []byte("%%EOF"))
		if start != -1 && end != -1 && start < end+5 {
			fileBytes = fileBytes[start : end+5]
		}
	}

	var parser parsers.BankStatementParser
	switch bankType {
	case "privatbank":
		parser = parsers.NewPrivatBankParser()
	case "monobank":
		parser = parsers.NewMonobankParser()
	default:
		return nil, fmt.Errorf("unknown bank: %s", bankType)
	}

	parsedTxs, err := parser.Parse(bytes.NewReader(fileBytes), int64(len(fileBytes)), file.Filename)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	var account models.Account
	var categoryMap map[string]string

	if err := s.DB.Select("family_id").Where("id = ?", accountID).First(&account).Error; err == nil {
		categoryMap = s.loadCategoryMap(account.FamilyID)
	}

	result := &PreviewResult{Transactions: []PreviewTransaction{}}

	for _, pt := range parsedTxs {
		if pt.Amount == 0 {
			continue
		}

		existingTx, isExact, isPotential, reason := s.checkMatchStatus(accountID, pt)
		
		// 1. Отримуємо чисте ім'я через твій клінер
		normalizedName := utils.NormalizeCounterparty(pt.Description)

		var foundCounterpartyID string
		var finalName = normalizedName

		// 2. Якщо такий контрагент є в БД, беремо його офіційну назву
		if cp, found := s.findCounterpartyByName(account.FamilyID, normalizedName); found {
			foundCounterpartyID = cp.ID
			finalName = cp.Name
		}

		// 🔥 3. Викликаємо винесену функцію з пакету utils, передаючи чистий finalName
		predictedCatID := utils.PredictCategoryID(pt.Description, finalName, pt.MCC, pt.Type, categoryMap)

		result.Transactions = append(result.Transactions, PreviewTransaction{
			Date:                 pt.Date.UnixMilli(),
			Amount:               pt.Amount,
			Description:          pt.Description,
			CounterpartyName:     finalName,
			CounterpartyID:       foundCounterpartyID,
			Type:                 pt.Type,
			PredictedCategory:    predictedCatID,
			IsDuplicate:          isExact,
			IsPotentialDuplicate: isPotential,
			MatchReason:          reason,
			ExistingTransaction:  existingTx,
			MCC:                  pt.MCC,
		})
	}
	return result, nil
}

func (s *ImportService) findCounterpartyByName(familyID, name string) (*models.Counterparty, bool) {
	if name == "" {
		return nil, false
	}
	var cps []models.Counterparty
	err := s.DB.Where("family_id = ? AND LOWER(name) = ? AND deleted_at IS NULL", familyID, strings.ToLower(name)).Limit(1).Find(&cps).Error
	if err != nil {
		return nil, false
	}
	if len(cps) > 0 {
		return &cps[0], true
	}
	return nil, false
}

func (s *ImportService) checkMatchStatus(accountID string, pt parsers.ParsedTransaction) (*models.Transaction, bool, bool, string) {
	amount := pt.Amount
	if amount < 0 {
		amount = -amount
	}
	timeWindow := int64(10 * 60 * 1000)
	startTime := pt.Date.UnixMilli() - timeWindow
	endTime := pt.Date.UnixMilli() + timeWindow

	var exactTx models.Transaction
	err := s.DB.Model(&models.Transaction{}).
		Preload("Category").
		Preload("Counterparty").
		Where("account_id = ?", accountID).
		Where("amount = ?", amount).
		Where("type = ?", pt.Type).
		Where("date >= ? AND date <= ?", startTime, endTime).
		Order("date ASC").
		First(&exactTx).Error

	if err == nil {
		return &exactTx, true, false, "Вже в базі (збіг часу)"
	}

	txTime := pt.Date
	dayStart := time.Date(txTime.Year(), txTime.Month(), txTime.Day(), 0, 0, 0, 0, txTime.Location()).UnixMilli()
	dayEnd := time.Date(txTime.Year(), txTime.Month(), txTime.Day(), 23, 59, 59, 999, txTime.Location()).UnixMilli()

	var candidates []models.Transaction
	s.DB.Model(&models.Transaction{}).
		Preload("Category").
		Preload("Counterparty").
		Where("account_id = ?", accountID).
		Where("amount = ?", amount).
		Where("type = ?", pt.Type).
		Where("date >= ? AND date <= ?", dayStart, dayEnd).
		Find(&candidates)

	if len(candidates) > 0 {
		return &candidates[0], false, true, "Збіг суми в цей день"
	}
	return nil, false, false, ""
}

func (s *ImportService) loadCategoryMap(familyID string) map[string]string {
	var categories []models.Category
	mapping := make(map[string]string)
	s.DB.Where("family_id = ?", familyID).Find(&categories)

	for _, cat := range categories {
		// Створюємо унікальний ключ з типом, напр: "income_інше"
		exactKey := fmt.Sprintf("%s_%s", cat.Type, strings.ToLower(cat.Name))
		mapping[exactKey] = cat.ID

		// Залишаємо і звичайну назву як запасний варіант
		mapping[strings.ToLower(cat.Name)] = cat.ID
	}
	return mapping
}