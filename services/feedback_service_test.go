package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFeedbackService(t *testing.T) {
	mockTg := new(MockTelegramClient)
	service := &feedbackService{tgClient: mockTg}

	t.Run("Send Feedback - Text Only", func(t *testing.T) {
		mockTg.On("SendMessage", mock.MatchedBy(func(text string) bool {
			return assert.Contains(t, text, "Test User") && assert.Contains(t, text, "Bug")
		})).Return(nil).Once()

		err := service.SendFeedback("Test User", "test@test.com", "Bug report", "high", nil)
		assert.NoError(t, err)
		mockTg.AssertExpectations(t)
	})

	t.Run("Send Feedback - Single Photo", func(t *testing.T) {
		images := [][]byte{{0x01, 0x02}}
		mockTg.On("SendPhoto", mock.Anything, "feedback.jpg", images[0]).Return(nil).Once()

		err := service.SendFeedback("Test User", "test@test.com", "Bug report", "medium", images)
		assert.NoError(t, err)
		mockTg.AssertExpectations(t)
	})

	t.Run("Send Feedback - Multiple Photos", func(t *testing.T) {
		images := [][]byte{{0x01}, {0x02}}
		mockTg.On("SendMediaGroup", mock.Anything, images).Return(nil).Once()

		err := service.SendFeedback("Test User", "test@test.com", "Bug report", "low", images)
		assert.NoError(t, err)
		mockTg.AssertExpectations(t)
	})
}
