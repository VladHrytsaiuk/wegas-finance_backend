package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestRoleRepository(t *testing.T) {
	db, _ := SetupTestDB()
	repo := NewRoleRepository(db)

	t.Run("Create and GetAll", func(t *testing.T) {
		role := &models.Role{
			Base: models.Base{ID: "role-1"},
			Name: "Admin",
		}
		err := repo.Create(role)
		assert.NoError(t, err)

		roles, err := repo.GetAll()
		assert.NoError(t, err)
		assert.Len(t, roles, 1)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete("role-1")
		assert.NoError(t, err)

		roles, _ := repo.GetAll()
		assert.Len(t, roles, 0)
	})
}
