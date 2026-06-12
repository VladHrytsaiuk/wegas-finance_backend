package repositories

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestGoalRepository(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	repo := NewGoalRepository(db)
	familyID := "fam-goals"
	userID := "u-goals"

	t.Run("Create and Find Goal", func(t *testing.T) {
		deadline := int64(2000000000000)
		goal := &models.Goal{
			Base:         models.Base{ID: "g-1"},
			FamilyID:     familyID,
			UserID:       userID,
			Name:         "Vacation",
			TargetAmount: 500000,
			Currency:     "USD",
			DateDeadline: &deadline,
			Status:       "active",
			Visibility:   "public",
		}
		err := repo.Create(goal)
		assert.NoError(t, err)

		saved, err := repo.FindOne("g-1")
		assert.NoError(t, err)
		assert.Equal(t, "Vacation", saved.Name)
	})

	t.Run("FindAllByFamily - Privacy Check", func(t *testing.T) {
		// Create another private goal by another user
		otherUser := "u-other"
		db.Create(&models.Goal{
			Base:       models.Base{ID: "g-private"},
			FamilyID:   familyID,
			UserID:     otherUser,
			Name:       "Secret",
			Visibility: "private",
		})

		// Current user should see their own goal + public goals, but not private ones from others
		goals, err := repo.FindAllByFamily(familyID, userID)
		assert.NoError(t, err)
		assert.Len(t, goals, 1)
		assert.Equal(t, "g-1", goals[0].ID)

		// Other user should see their private one too
		otherGoals, _ := repo.FindAllByFamily(familyID, otherUser)
		assert.Len(t, otherGoals, 2)
	})

	t.Run("FindAllActive", func(t *testing.T) {
		active, err := repo.FindAllActive()
		assert.NoError(t, err)
		// g-1 is active, g-private is active (by default if not specified)
		assert.GreaterOrEqual(t, len(active), 2)
	})

	t.Run("Update and Delete", func(t *testing.T) {
		goal, _ := repo.FindOne("g-1")
		goal.Status = "completed"
		err := repo.Update(goal)
		assert.NoError(t, err)

		err = repo.Delete("g-1")
		assert.NoError(t, err)

		_, err = repo.FindOne("g-1")
		assert.Error(t, err)
	})
}
