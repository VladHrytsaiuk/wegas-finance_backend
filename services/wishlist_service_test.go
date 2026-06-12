package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
)

func TestWishlistService(t *testing.T) {
	db, err := repositories.SetupTestDB()
	assert.NoError(t, err)

	repo := repositories.NewWishlistRepo(db)
	service := NewWishlistService(repo)

	familyID := "fam-wish"
	userID := "u-wish"
	otherUser := "u-other"

	t.Run("Create Item and Toggle Reservation", func(t *testing.T) {
		item, err := service.CreateItem(models.CreateWishlistRequest{
			Name: "Lego",
		}, userID, familyID)
		assert.NoError(t, err)

		// Owner cannot reserve their own item
		err = service.ToggleReservation(item.ID, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot reserve your own item")

		// Other user can reserve
		err = service.ToggleReservation(item.ID, otherUser)
		assert.NoError(t, err)

		// Check surprise logic: owner shouldn't see who reserved
		itemsForOwner, _ := service.GetItems(familyID, userID, "", "")
		assert.Nil(t, itemsForOwner[0].ReservedByUserID)

		// Other user should see it
		itemsForOther, _ := service.GetItems(familyID, otherUser, "", "")
		assert.NotNil(t, itemsForOther[0].ReservedByUserID)
		assert.Equal(t, otherUser, *itemsForOther[0].ReservedByUserID)
	})
}
