package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestWishlistRepo(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	repo := NewWishlistRepo(db)
	familyID := "fam-wish"
	userID := "u-wish"

	t.Run("Create and Get Groups", func(t *testing.T) {
		group := &models.WishlistGroup{
			Base:     models.Base{ID: "g-1"},
			FamilyID: familyID,
			UserID:   userID,
			Name:     "Birthday",
		}
		err := repo.CreateGroup(group)
		assert.NoError(t, err)

		groups, err := repo.GetGroups(familyID, userID)
		assert.NoError(t, err)
		assert.Len(t, groups, 1)
		assert.Equal(t, "Birthday", groups[0].Name)
	})

	t.Run("Create and Get Items", func(t *testing.T) {
		item := &models.WishlistItem{
			Base:     models.Base{ID: "i-1"},
			FamilyID: familyID,
			UserID:   userID,
			Name:     "PS5",
			Priority: 3,
		}
		err := repo.Create(item)
		assert.NoError(t, err)

		items, err := repo.GetAll(familyID, userID, "", "")
		assert.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, "PS5", items[0].Name)
	})

	t.Run("Delete Group and Unlink Items", func(t *testing.T) {
		// Link item to group
		db.Model(&models.WishlistItem{}).Where("id = ?", "i-1").Update("group_id", "g-1")

		err := repo.DeleteGroup("g-1", familyID)
		assert.NoError(t, err)

		// Item should still exist but group_id should be nil
		item, _ := repo.GetByID("i-1", familyID)
		assert.Nil(t, item.GroupID)

		groups, _ := repo.GetGroups(familyID, userID)
		assert.Len(t, groups, 0)
	})
}
