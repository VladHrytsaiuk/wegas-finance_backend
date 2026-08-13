package services

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/pkg/telegram"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"gorm.io/gorm"
)

type TelegramBotService interface {
	HandleUpdate(update *telegram.Update) error
	StartPolling()
}

type telegramBotService struct {
	botClient        telegram.BotAPI
	linkRepo         repositories.TelegramLinkRepository
	userRepo         repositories.UserRepository
	accountService   AccountService
	inboxService     InboxService
	receiptIngestion ReceiptIngestionService
}

func NewTelegramBotService(
	botClient telegram.BotAPI,
	linkRepo repositories.TelegramLinkRepository,
	userRepo repositories.UserRepository,
	accountService AccountService,
	inboxService InboxService,
	receiptIngestion ReceiptIngestionService,
) TelegramBotService {
	return &telegramBotService{
		botClient:        botClient,
		linkRepo:         linkRepo,
		userRepo:         userRepo,
		accountService:   accountService,
		inboxService:     inboxService,
		receiptIngestion: receiptIngestion,
	}
}

func (s *telegramBotService) HandleUpdate(update *telegram.Update) error {
	switch {
	case update == nil:
		return nil
	case update.Message != nil:
		return s.handleMessage(update.Message)
	case update.EditedMessage != nil:
		return s.handleMessage(update.EditedMessage)
	case update.ChannelPost != nil:
		return s.handleMessage(update.ChannelPost)
	case update.EditedChannelPost != nil:
		return s.handleMessage(update.EditedChannelPost)
	default:
		return nil
	}
}

func (s *telegramBotService) handleMessage(message *telegram.Message) error {
	if message == nil || message.Chat.ID == 0 {
		return nil
	}

	// Channel posts and some forwarded updates may not contain a user context.
	// They should never break webhook delivery for the rest of the queue.
	if message.From == nil {
		log.Printf("telegram message ignored: missing from, chat_id=%d, file=%q, mime=%q", message.Chat.ID, documentFileName(message.Document), documentMimeType(message.Document))
		return nil
	}

	text := strings.TrimSpace(message.Text)
	if strings.HasPrefix(text, "/start") {
		return s.handleStart(message, text)
	}

	user, err := s.resolveLinkedUser(message.From)
	if err != nil {
		log.Printf("telegram linked user not found: chat_id=%d from_id=%d file=%q mime=%q err=%v", message.Chat.ID, message.From.ID, documentFileName(message.Document), documentMimeType(message.Document), err)
		return s.botClient.SendChatMessage(message.Chat.ID, "Спочатку підключіть Telegram у застосунку.", nil)
	}

	if message.Document != nil {
		return s.handleDocument(message.Chat.ID, user, message.Document)
	}

	if isLikelyURL(text) {
		entry, err := s.receiptIngestion.IngestURL(text, user)
		if err != nil {
			return s.botClient.SendChatMessage(message.Chat.ID, buildTelegramParseErrorMessage("url"), nil)
		}
		return s.respondToInboxEntry(message.Chat.ID, entry, user)
	}

	return s.botClient.SendChatMessage(message.Chat.ID, buildTelegramHelpMessage(), nil)
}

func (s *telegramBotService) handleStart(message *telegram.Message, text string) error {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return s.botClient.SendChatMessage(message.Chat.ID, "Відкрийте бота через кнопку підключення у застосунку.", nil)
	}
	if message.From == nil {
		return s.botClient.SendChatMessage(message.Chat.ID, "Не вдалося визначити ваш Telegram акаунт.", nil)
	}

	_, err := s.linkRepo.GetActiveLinkByTelegramUserID(message.From.ID)
	if err == nil {
		return s.botClient.SendChatMessage(message.Chat.ID, "Цей Telegram акаунт уже підключений.", nil)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	_, err = NewTelegramLinkService(s.linkRepo, s.userRepo.GetDB(), "").CompleteLink(TelegramLinkCompleteInput{
		Token:          parts[1],
		TelegramUserID: message.From.ID,
		TelegramChatID: message.Chat.ID,
		Username:       message.From.Username,
		FirstName:      message.From.FirstName,
	})
	if err != nil {
		return s.botClient.SendChatMessage(message.Chat.ID, "Не вдалося підключити Telegram. Перегенеруйте посилання в застосунку.", nil)
	}

	return s.botClient.SendChatMessage(message.Chat.ID, buildTelegramWelcomeMessage(), nil)
}

func (s *telegramBotService) handleDocument(chatID int64, user *models.User, document *telegram.Document) error {
	if document == nil {
		return nil
	}
	if !isXMLDocument(document) {
		return s.botClient.SendChatMessage(chatID, "Поки підтримуються лише XML-чеки.", nil)
	}

	file, err := s.botClient.GetFile(document.FileID)
	if err != nil {
		return s.botClient.SendChatMessage(chatID, "Не вдалося отримати файл з Telegram.", nil)
	}
	raw, err := s.botClient.DownloadFile(file.FilePath)
	if err != nil {
		return s.botClient.SendChatMessage(chatID, "Не вдалося завантажити XML-файл.", nil)
	}

	entry, err := s.receiptIngestion.IngestXMLBytes(raw, user)
	if err != nil {
		return s.botClient.SendChatMessage(chatID, buildTelegramParseErrorMessage("xml"), nil)
	}
	return s.respondToInboxEntry(chatID, entry, user)
}

func (s *telegramBotService) respondToInboxEntry(chatID int64, entry *models.InboxEntry, user *models.User) error {
	if entry == nil {
		return s.botClient.SendChatMessage(chatID, "<b>Чек оброблено</b>", nil)
	}

	_ = user
	return s.botClient.SendChatMessage(chatID, buildTelegramSavedMessage(entry), nil)
}

func (s *telegramBotService) resolveLinkedUser(from *telegram.User) (*models.User, error) {
	if from == nil {
		return nil, gorm.ErrRecordNotFound
	}
	link, err := s.linkRepo.GetActiveLinkByTelegramUserID(from.ID)
	if err != nil {
		return nil, err
	}
	return s.userRepo.GetByID(link.UserID)
}

func formatAmount(total *int64, currency string) string {
	if total == nil {
		return currency
	}
	return fmt.Sprintf("%.2f %s", float64(*total)/100, strings.TrimSpace(currency))
}

func isLikelyURL(value string) bool {
	return strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "http://")
}

func isXMLDocument(document *telegram.Document) bool {
	if document == nil {
		return false
	}
	if strings.EqualFold(document.MimeType, "application/xml") || strings.EqualFold(document.MimeType, "text/xml") {
		return true
	}
	return strings.EqualFold(filepath.Ext(document.FileName), ".xml")
}

func documentFileName(document *telegram.Document) string {
	if document == nil {
		return ""
	}
	return document.FileName
}

func documentMimeType(document *telegram.Document) string {
	if document == nil {
		return ""
	}
	return document.MimeType
}

func (s *telegramBotService) StartPolling() {
	offset := 0
	log.Println("🤖 Starting Telegram Bot in Long Polling mode...")

	for {
		updates, err := s.botClient.GetUpdates(offset, 30)
		if err != nil {
			log.Printf("❌ Error polling telegram updates: %v", err)
			continue
		}

		for _, update := range updates {
			if update.UpdateID >= int64(offset) {
				offset = int(update.UpdateID) + 1
			}

			go func(u telegram.Update) {
				if err := s.HandleUpdate(&u); err != nil {
					log.Printf("❌ Error handling update %d: %v", u.UpdateID, err)
				}
			}(update)
		}
	}
}
