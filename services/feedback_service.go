package services

import (
	"fmt"
	"html" // 1. Додаємо імпорт стандартного пакета html

	"github.com/VladHrytsaiuk/wegas-finance/backend/pkg/telegram"
)

type FeedbackService interface {
	SendFeedback(name, contact, message, priority string, images [][]byte) error
}

type feedbackService struct {
	tgClient *telegram.Client
}

func NewFeedbackService(tgToken, tgChatID string) FeedbackService {
	return &feedbackService{
		tgClient: telegram.NewClient(tgToken, tgChatID),
	}
}

func (s *feedbackService) SendFeedback(name, contact, message, priority string, images [][]byte) error {
	var priorityIcon string
	var headerTitle string

	switch priority {
	case "high":
		priorityIcon = "🔴"
		headerTitle = "CRITICAL BUG / URGENT"
	case "medium":
		priorityIcon = "🟡"
		headerTitle = "Feedback / Question"
	case "low":
		priorityIcon = "🟢"
		headerTitle = "Idea / Low Priority"
	default:
		priorityIcon = "⚪"
		headerTitle = "New Feedback"
		priority = "normal"
	}

	// 2. Екрануємо ВСІ дані, які ввів користувач
	safeName := html.EscapeString(name)
	safeContact := html.EscapeString(contact)
	safeMessage := html.EscapeString(message)

	// 3. Підставляємо вже безпечні змінні (safeName, safeContact, safeMessage)
	text := fmt.Sprintf(
		"%s <b>%s</b>\n\n"+
			"📊 <b>Priority:</b> %s %s\n"+
			"👤 <b>Name:</b> %s\n"+
			"📧 <b>Contact:</b> %s\n\n"+
			"📝 <b>Message:</b>\n%s",
		priorityIcon, headerTitle,
		priorityIcon, priority,
		safeName, safeContact, safeMessage, // Використовуємо екрановані дані
	)

	count := len(images)

	if count > 1 {
		return s.tgClient.SendMediaGroup(text, images)
	} else if count == 1 {
		return s.tgClient.SendPhoto(text, "feedback.jpg", images[0])
	}

	return s.tgClient.SendMessage(text)
}