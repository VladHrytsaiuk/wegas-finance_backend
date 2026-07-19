package services

import (
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const receiptLearningMinConfirmations = 2

func receiptLearningKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.Join(strings.Fields(value), " ")
	return strings.Trim(value, " .,:;!?\"'")
}

func receiptPreferenceForMerchant(db *gorm.DB, user *models.User, merchant string) (*models.ReceiptMerchantPreference, error) {
	merchantKey := receiptLearningKey(merchant)
	if merchantKey == "" {
		return nil, nil
	}

	var preference models.ReceiptMerchantPreference
	err := db.Where("family_id = ? AND user_id = ? AND merchant_key = ? AND confirmations >= ?", user.FamilyID, user.ID, merchantKey, receiptLearningMinConfirmations).
		First(&preference).Error
	if errorsIsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &preference, nil
}

func receiptItemCategoryForName(db *gorm.DB, user *models.User, merchant, itemName string) *string {
	merchantKey := receiptLearningKey(merchant)
	itemKey := receiptLearningKey(itemName)
	if merchantKey == "" || itemKey == "" {
		return nil
	}

	var preference models.ReceiptItemCategoryPreference
	err := db.Where("family_id = ? AND user_id = ? AND merchant_key = ? AND item_key = ? AND confirmations >= ?", user.FamilyID, user.ID, merchantKey, itemKey, receiptLearningMinConfirmations).
		Order("confirmations DESC").
		First(&preference).Error
	if err != nil || preference.CategoryID == "" {
		return nil
	}
	return &preference.CategoryID
}

func learnReceiptPreferences(db *gorm.DB, user *models.User, source *models.ReceiptSource, transaction models.Transaction, transactionItems []models.TransactionItem, now int64) error {
	merchantKey := receiptLearningKey(source.Merchant)
	if merchantKey == "" {
		return nil
	}

	// Do not count a receipt which already arrived with both learned values.
	if source.CounterpartyID == nil || source.CategoryID == nil {
		preference := models.ReceiptMerchantPreference{
			Base:          models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
			FamilyID:      user.FamilyID,
			UserID:        user.ID,
			MerchantKey:   merchantKey,
			Confirmations: 1,
		}
		if transaction.CounterpartyID != "" {
			preference.CounterpartyID = &transaction.CounterpartyID
		}
		if transaction.CategoryID != "" {
			preference.CategoryID = &transaction.CategoryID
		}

		var existing models.ReceiptMerchantPreference
		err := db.Where("family_id = ? AND user_id = ? AND merchant_key = ?", user.FamilyID, user.ID, merchantKey).First(&existing).Error
		if errorsIsNotFound(err) {
			if err := db.Create(&preference).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if err := db.Model(&existing).Updates(map[string]interface{}{
			"counterparty_id": preference.CounterpartyID,
			"category_id":     preference.CategoryID,
			"confirmations":   gorm.Expr("confirmations + 1"),
			"updated_at":      now,
		}).Error; err != nil {
			return err
		}
	}

	for _, sourceItem := range source.Items {
		// Do not learn from a category which was already predicted for this receipt.
		if sourceItem.CategoryID != nil {
			continue
		}
		itemKey := receiptLearningKey(sourceItem.Name)
		if itemKey == "" {
			continue
		}
		for _, transactionItem := range transactionItems {
			if receiptLearningKey(transactionItem.Name) != itemKey || transactionItem.TotalAmount != sourceItem.TotalAmount || transactionItem.CategoryID == nil || *transactionItem.CategoryID == "" {
				continue
			}

			preference := models.ReceiptItemCategoryPreference{
				Base:          models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true},
				FamilyID:      user.FamilyID,
				UserID:        user.ID,
				MerchantKey:   merchantKey,
				ItemKey:       itemKey,
				CategoryID:    *transactionItem.CategoryID,
				Confirmations: 1,
			}

			var existing models.ReceiptItemCategoryPreference
			err := db.Where("family_id = ? AND user_id = ? AND merchant_key = ? AND item_key = ? AND category_id = ?", user.FamilyID, user.ID, merchantKey, itemKey, preference.CategoryID).First(&existing).Error
			if errorsIsNotFound(err) {
				if err := db.Create(&preference).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else if err := db.Model(&existing).Updates(map[string]interface{}{
				"confirmations": gorm.Expr("confirmations + 1"),
				"updated_at":    now,
			}).Error; err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func errorsIsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
