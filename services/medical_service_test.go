package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
)

func TestMedicalService(t *testing.T) {
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)

	repo := repositories.NewMedicalRepository(db)
	service := NewMedicalService(repo)

	familyID := "fam-med"
	adminUser := &models.User{Base: models.Base{ID: "u-admin"}, FamilyID: familyID, RoleID: "admin"}
	childUser := &models.User{Base: models.Base{ID: "u-child"}, FamilyID: familyID, RoleID: "child"}

	record := &models.MedicalRecord{
		Base: models.Base{ID: "rec-1"},
		Title: "Flu",
	}

	t.Run("Create Record", func(t *testing.T) {
		err := service.CreateRecord(adminUser, record)
		assert.NoError(t, err)
		assert.Equal(t, familyID, record.FamilyID)
		assert.Equal(t, adminUser.ID, record.UserID)
	})

	t.Run("Get All Records", func(t *testing.T) {
		records, err := service.GetAllRecords(adminUser)
		assert.NoError(t, err)
		assert.Len(t, records, 1)

		childRecords, err := service.GetAllRecords(childUser)
		assert.NoError(t, err)
		assert.Len(t, childRecords, 0) // Child can't see admin's records
	})

	t.Run("Update Record - Forbidden for child", func(t *testing.T) {
		err := service.UpdateRecord(childUser, record)
		assert.Equal(t, ErrForbidden, err)
	})

	t.Run("Delete Record - Forbidden for child", func(t *testing.T) {
		err := service.DeleteRecord(childUser, record.ID)
		assert.Equal(t, ErrForbidden, err)
	})

	t.Run("Admin can delete", func(t *testing.T) {
		err := service.DeleteRecord(adminUser, record.ID)
		assert.NoError(t, err)
	})
}
