package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestDBResilience(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)
	
	repo := NewAccountRepository(db)
	
	t.Run("DB connection closed", func(t *testing.T) {
		// Get underlying SQL DB and close it
		sqlDB, _ := db.DB()
		sqlDB.Close()
		
		err := repo.Create(&models.Account{Name: "Failed"})
		assert.Error(t, err, "Should fail when DB is closed")
	})
}
