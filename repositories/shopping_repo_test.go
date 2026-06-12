package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestShoppingRepo(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	repo := NewShoppingRepo(db)
	familyID := "fam-shop"
	userID := "u-shop"

	t.Run("Create and Get Lists", func(t *testing.T) {
		list := &models.ShoppingList{
			Base:       models.Base{ID: "l-1"},
			FamilyID:   familyID,
			UserID:     userID,
			Title:      "Groceries",
			Visibility: "public",
		}
		err := repo.CreateList(list)
		assert.NoError(t, err)

		lists, err := repo.GetLists(familyID, userID)
		assert.NoError(t, err)
		assert.Len(t, lists, 1)
		assert.Equal(t, "Groceries", lists[0].Title)
	})

	t.Run("Create and Manage Items", func(t *testing.T) {
		item := &models.ShoppingItem{
			Base:     models.Base{ID: "i-1"},
			ListID:   "l-1",
			Name:     "Milk",
			IsBought: false,
		}
		err := repo.CreateItem(item)
		assert.NoError(t, err)

		// Update
		err = repo.UpdateItem("i-1", map[string]interface{}{"is_bought": true})
		assert.NoError(t, err)

		// Check
		lists, _ := repo.GetLists(familyID, userID)
		assert.Len(t, lists[0].Items, 1)
		assert.True(t, lists[0].Items[0].IsBought)

		// Clear completed
		err = repo.ClearCompletedInList("l-1")
		assert.NoError(t, err)
		
		lists, _ = repo.GetLists(familyID, userID)
		assert.Len(t, lists[0].Items, 0)
	})

	t.Run("Delete List", func(t *testing.T) {
		err := repo.DeleteList("l-1", familyID)
		assert.NoError(t, err)

		lists, _ := repo.GetLists(familyID, userID)
		assert.Len(t, lists, 0)
	})
}
