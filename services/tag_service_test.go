package services

import (
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTagService_Mock(t *testing.T) {
	mockRepo := new(MockTagRepository)
	service := NewTagService(mockRepo)

	user := &models.User{
		Base:     models.Base{ID: "u-1"},
		FamilyID: "f-1",
	}

	t.Run("Create Tag", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything).Return(nil).Once()
		
		tag, err := service.Create("Work", "#ff0000", user)
		
		assert.NoError(t, err)
		assert.Equal(t, "Work", tag.Name)
		assert.Equal(t, "f-1", tag.FamilyID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetAll Tags", func(t *testing.T) {
		mockRepo.On("GetAll", "f-1").Return([]models.Tag{{Name: "Tag1"}}, nil).Once()
		
		tags, err := service.GetAll("f-1")
		
		assert.NoError(t, err)
		assert.Len(t, tags, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete Tag - Success", func(t *testing.T) {
		mockRepo.On("GetByID", "t-1", "f-1").Return(&models.Tag{}, nil).Once()
		mockRepo.On("Delete", "t-1", "f-1").Return(nil).Once()
		
		err := service.Delete("t-1", user)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete Tag - Not Found", func(t *testing.T) {
		mockRepo.On("GetByID", "t-2", "f-1").Return(nil, assert.AnError).Once()
		
		err := service.Delete("t-2", user)
		
		assert.Error(t, err)
		assert.Equal(t, "tag not found", err.Error())
		mockRepo.AssertExpectations(t)
	})
}
