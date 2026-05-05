package services

import (
	"errors"
	"image"
	"image/jpeg"
	_ "image/png" // Підтримка PNG
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
)

var (
	ErrGoalNotFound = errors.New("goal not found")
	ErrForbidden    = errors.New("access denied")
)

type GoalService struct {
	repo            *repositories.GoalRepository
	accountRepo     repositories.AccountRepository
	currencyService CurrencyService
}

func NewGoalService(
	repo *repositories.GoalRepository,
	accRepo repositories.AccountRepository,
	currService CurrencyService,
) *GoalService {
	return &GoalService{
		repo:            repo,
		accountRepo:     accRepo,
		currencyService: currService,
	}
}

// --- ДОПОМІЖНІ МЕТОДИ ПРИВАТНОСТІ ---

// verifyEditAccess: Редагувати може тільки власник (userID)
func (s *GoalService) verifyEditAccess(goal *models.Goal, userID string) error {
	if goal.UserID != userID {
		return ErrForbidden
	}
	return nil
}

// verifyReadAccess: Читати може власник АБО якщо ціль публічна і не прихована від цього юзера
func (s *GoalService) verifyReadAccess(goal *models.Goal, userID string) error {
	if goal.UserID == userID {
		return nil
	}
	if goal.Visibility == "public" && goal.HiddenFrom != userID {
		return nil
	}
	return ErrForbidden
}

// --- ОСНОВНА ЛОГІКА ---

// UploadGoalPhoto завантажує фото і оновлює БД
func (s *GoalService) UploadGoalPhoto(goalID string, userID string, fileReader io.Reader) (string, error) {
	// 1. Знаходимо ціль
	existingGoal, err := s.repo.FindOne(goalID)
	if err != nil {
		return "", ErrGoalNotFound
	}

	// 2. Перевіряємо права (тільки власник)
	if existingGoal.UserID != userID {
		return "", ErrForbidden
	}

	// 3. Видаляємо старе фото, якщо було
	if existingGoal.PhotoURL != "" {
		_ = os.Remove("." + existingGoal.PhotoURL) // Додаємо крапку, бо шлях в БД починається з /
	}

	// 4. Зберігаємо нове
	// Створюємо структуру папок: uploads/goals
	if err := os.MkdirAll("uploads/goals", 0755); err != nil {
		return "", err
	}

	// Генеруємо ім'я
	newFileName := uuid.New().String() + ".jpg"
	localPath := filepath.Join("uploads", "goals", newFileName) // uploads/goals/123.jpg

	// Створюємо файл
	out, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Декодуємо і зберігаємо картинку
	img, _, err := image.Decode(fileReader)
	if err != nil {
		return "", errors.New("invalid image format")
	}
	err = jpeg.Encode(out, img, &jpeg.Options{Quality: 80})
	if err != nil {
		return "", err
	}

	// 5. 🔥 ГОЛОВНЕ: Оновлюємо посилання в Базі Даних
	// URL для браузера: /uploads/goals/123.jpg
	webPath := "/uploads/goals/" + newFileName

	// Використовуємо прямий SQL UPDATE через GORM, щоб точно записати
	err = s.repo.GetDB().Model(&models.Goal{}).
		Where("id = ?", goalID).
		UpdateColumn("photo_url", webPath).Error

	if err != nil {
		return "", err
	}

	return webPath, nil
}

func (s *GoalService) Create(goal *models.Goal, userID string) error {
	goal.ID = uuid.New().String()
	goal.UserID = userID // Власник
	goal.CreatedAt = time.Now().UnixMilli()
	goal.UpdatedAt = time.Now().UnixMilli()

	if goal.Status == "" {
		goal.Status = "active"
	}
	// Дефолтні налаштування приватності
	if goal.Visibility == "" {
		goal.Visibility = "public"
	}

	return s.repo.Create(goal)
}

func (s *GoalService) GetAll(familyID string, userID string) ([]models.Goal, error) {
	// Репозиторій вже має фільтрувати за (familyID + Privacy Logic)
	goals, err := s.repo.FindAllByFamily(familyID, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	for i := range goals {
		// Розрахунок суми
		total := s.calculateTotalAmount(&goals[i])
		goals[i].CurrentAmount = total

		// Перевірка провалу цілі (failed)
		if goals[i].DateDeadline != nil && *goals[i].DateDeadline < now &&
			goals[i].Status == "active" && total < goals[i].TargetAmount {
			goals[i].Status = "failed"
			// Оновлюємо статус асинхронно, щоб не гальмувати відповідь
			go s.repo.Update(&goals[i])
		}
	}
	return goals, nil
}

func (s *GoalService) GetOne(id string, userID string) (*models.Goal, error) {
	goal, err := s.repo.FindOne(id)
	if err != nil {
		return nil, err
	}

	// Перевірка доступу (Private/HiddenFrom)
	if err := s.verifyReadAccess(goal, userID); err != nil {
		return nil, err
	}

	total := s.calculateTotalAmount(goal)
	goal.CurrentAmount = total

	now := time.Now().UnixMilli()
	if goal.DateDeadline != nil && *goal.DateDeadline < now &&
		goal.Status == "active" && total < goal.TargetAmount {
		goal.Status = "failed"
		_ = s.repo.Update(goal)
	}
	return goal, nil
}

func (s *GoalService) Update(incomingGoal *models.Goal, userID string) error {
	existingGoal, err := s.repo.FindOne(incomingGoal.ID)
	if err != nil {
		return err
	}

	// Перевірка прав (тільки власник)
	if err := s.verifyEditAccess(existingGoal, userID); err != nil {
		return err
	}

	updates := make(map[string]interface{})

	// Логіка видалення фото через звичайний Update (якщо прийшов прапорець RemovePhoto)
	if incomingGoal.RemovePhoto {
		if existingGoal.PhotoURL != "" {
			relativePath := strings.TrimPrefix(existingGoal.PhotoURL, "/")
			_ = os.Remove(relativePath)
		}
		updates["photo_url"] = ""
	}

	// Оновлення полів
	if incomingGoal.Name != "" { updates["name"] = incomingGoal.Name }
	if incomingGoal.Description != "" { updates["description"] = incomingGoal.Description }
	if incomingGoal.TargetAmount > 0 { updates["target_amount"] = incomingGoal.TargetAmount }
	if incomingGoal.Currency != "" { updates["currency"] = incomingGoal.Currency }
	if incomingGoal.Color != "" { updates["color"] = incomingGoal.Color }
	if incomingGoal.Icon != "" { updates["icon"] = incomingGoal.Icon }
	if incomingGoal.ExternalLink != "" { updates["external_link"] = incomingGoal.ExternalLink }
	if incomingGoal.DateStart > 0 { updates["date_start"] = incomingGoal.DateStart }
	if incomingGoal.DateDeadline != nil { updates["date_deadline"] = incomingGoal.DateDeadline }
	if incomingGoal.Status != "" { updates["status"] = incomingGoal.Status }

	// Приватність
	if incomingGoal.Visibility != "" { updates["visibility"] = incomingGoal.Visibility }
	// HiddenFrom може бути пустим рядком (щоб скинути налаштування), тому перевіряємо інакше або просто перезаписуємо
	updates["hidden_from"] = incomingGoal.HiddenFrom 

	updates["updated_at"] = time.Now().UnixMilli()

	err = s.repo.GetDB().Model(&models.Goal{}).Where("id = ?", existingGoal.ID).Updates(updates).Error
	if err != nil {
		return err
	}

	// Перерахунок статусу після оновлення
	updatedGoal, _ := s.repo.FindOne(existingGoal.ID)
	total := s.calculateTotalAmount(updatedGoal)
	now := time.Now().UnixMilli()
	if updatedGoal.DateDeadline != nil && *updatedGoal.DateDeadline < now &&
		updatedGoal.Status == "active" && total < updatedGoal.TargetAmount {
		return s.repo.GetDB().Model(&models.Goal{}).Where("id = ?", updatedGoal.ID).Update("status", "failed").Error
	}

	return nil
}

func (s *GoalService) Delete(id string, userID string) error {
	existingGoal, err := s.repo.FindOne(id)
	if err != nil {
		// Якщо вже видалено або не знайдено - вважаємо успіхом, щоб не блокувати фронт
		return nil 
	}

	// Перевірка прав (тільки власник)
	if err := s.verifyEditAccess(existingGoal, userID); err != nil {
		return err
	}

	if existingGoal.PhotoURL != "" {
		relativePath := strings.TrimPrefix(existingGoal.PhotoURL, "/")
		_ = os.Remove(relativePath)
	}
	return s.repo.Delete(id)
}

func (s *GoalService) calculateTotalAmount(goal *models.Goal) int64 {
	var total int64 = 0
	for _, acc := range goal.Accounts {
		if acc.Balance <= 0 {
			continue
		}
		if acc.Currency == goal.Currency {
			total += acc.Balance
			continue
		}
		convertedAmount, err := s.currencyService.Convert(acc.Balance, acc.Currency, goal.Currency)
		if err != nil {
			continue
		}
		total += convertedAmount
	}
	return total
}

func (s *GoalService) LinkAccount(goalID string, accountID string) error {
	// Тут також варто додати перевірку, чи належить accountID тій самій сім'ї/юзеру
	acc, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return err
	}
	acc.GoalID = &goalID
	return s.accountRepo.Update(acc)
}

func (s *GoalService) UnlinkAccount(accountID string) error {
	acc, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return err
	}
	acc.GoalID = nil
	return s.accountRepo.Update(acc)
}

// CheckOverdueGoals - системний метод (background job), тут userID не передаємо
func (s *GoalService) CheckOverdueGoals() error {
	goals, err := s.repo.FindAllActive()
	if err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	for _, goal := range goals {
		if goal.DateDeadline == nil {
			continue
		}
		if *goal.DateDeadline < now {
			total := s.calculateTotalAmount(&goal)
			if total < goal.TargetAmount {
				goal.Status = "failed"
				goal.UpdatedAt = now
				_ = s.repo.Update(&goal)
			}
		}
	}
	return nil
}