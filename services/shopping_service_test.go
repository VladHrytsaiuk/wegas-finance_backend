package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
)

func TestShoppingService(t *testing.T) {
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)

	repo := repositories.NewShoppingRepo(db)
	service := NewShoppingService(repo)

	familyID := "fam-shop"
	userID := "u-shop"

	t.Run("Create List and Add Item", func(t *testing.T) {
		list, err := service.CreateList(models.CreateShoppingListRequest{
			Title: "Market",
		}, userID, familyID)
		assert.NoError(t, err)
		assert.Equal(t, "Market", list.Title)

		item, err := service.AddItemToList(list.ID, models.CreateShoppingItemRequest{
			Name: "Bread",
		})
		assert.NoError(t, err)
		assert.Equal(t, "Bread", item.Name)
		assert.Equal(t, list.ID, item.ListID)
	})

	t.Run("Update and Delete", func(t *testing.T) {
		lists, _ := service.GetLists(familyID, userID)
		listID := lists[0].ID

		newTitle := "New Market"
		err := service.UpdateList(listID, models.UpdateShoppingListRequest{
			Title: &newTitle,
		}, familyID)
		assert.NoError(t, err)

		err = service.DeleteList(listID, familyID)
		assert.NoError(t, err)

		listsAfter, _ := service.GetLists(familyID, userID)
		assert.Len(t, listsAfter, 0)
	})
}
