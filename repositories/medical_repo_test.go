package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestMedicalRepository(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	repo := NewMedicalRepository(db)
	familyID := "fam-med"
	userID := "user-med"

	record := &models.MedicalRecord{
		Base:     models.Base{ID: "rec-1"},
		FamilyID: familyID,
		UserID:   userID,
		Title:    "General Checkup",
		Date:     1700000000000,
	}

	t.Run("Create and GetByID", func(t *testing.T) {
		err := repo.Create(record)
		assert.NoError(t, err)

		found, err := repo.GetByID(record.ID, familyID)
		assert.NoError(t, err)
		assert.Equal(t, "General Checkup", found.Title)
	})

	t.Run("GetAll", func(t *testing.T) {
		records, err := repo.GetAll(familyID, "")
		assert.NoError(t, err)
		assert.Len(t, records, 1)
	})

	t.Run("Update", func(t *testing.T) {
		record.Title = "Updated Checkup"
		err := repo.Update(record)
		assert.NoError(t, err)

		found, _ := repo.GetByID(record.ID, familyID)
		assert.Equal(t, "Updated Checkup", found.Title)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(record.ID, familyID)
		assert.NoError(t, err)

		_, err = repo.GetByID(record.ID, familyID)
		assert.Error(t, err)
	})
}
