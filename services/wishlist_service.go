package services

import (
	"errors"
	"fmt"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid"
)

type WishlistService struct {
  repo *repositories.WishlistRepo
}

func NewWishlistService(repo *repositories.WishlistRepo) *WishlistService {
  return &WishlistService{repo: repo}
}

// --- GROUPS LOGIC ---

func (s *WishlistService) CreateGroup(name, color, icon, visibility, hiddenFrom, userID, familyID string) (*models.WishlistGroup, error) {
  group := &models.WishlistGroup{
    Name:       name,
    Color:      color,
    Icon:       icon,
    Visibility: visibility,
    HiddenFrom: hiddenFrom,
    UserID:     userID,
    FamilyID:   familyID,
  }

  if group.Visibility == "" {
    group.Visibility = "public"
  }
  group.ID = uuid.New().String()
  err := s.repo.CreateGroup(group)
  return group, err
}

func (s *WishlistService) GetGroups(familyID, myUserID string) ([]models.WishlistGroup, error) {
  return s.repo.GetGroups(familyID, myUserID)
}

func (s *WishlistService) UpdateGroup(id string, req models.UpdateWishlistGroupRequest, familyID string) error {
  updates := make(map[string]interface{})
  if req.Name != "" {
    updates["name"] = req.Name
  }
  if req.Color != "" {
    updates["color"] = req.Color
  }
  if req.Icon != "" {
    updates["icon"] = req.Icon
  }
  if req.Visibility != "" {
    updates["visibility"] = req.Visibility
  }
  // Always allow updating hidden_from, even to clear it?
  // If it's a string, we cannot distinguish between "not provided" and "clear".
  // But it's an update, so if user provides "", maybe they want to clear it?
  // Or maybe they just didn't provide it?
  // Actually, if HiddenFrom is empty string, we can set it.
  updates["hidden_from"] = req.HiddenFrom

  if len(updates) > 0 {
    return s.repo.UpdateGroup(id, familyID, updates)
  }
  return nil
}

func (s *WishlistService) DeleteGroup(id, familyID string) error {
  return s.repo.DeleteGroup(id, familyID)
}

// --- ITEMS LOGIC ---

func (s *WishlistService) CreateItem(req models.CreateWishlistRequest, userID, familyID string) (*models.WishlistItem, error) {
  if req.Priority < 1 || req.Priority > 3 {
    req.Priority = 1
  }

  // Обробка GroupID (пуста стрічка = nil)
  var groupID *string
  if req.GroupID != "" {
    groupID = &req.GroupID
  }

  item := &models.WishlistItem{
    UserID:     userID,
    FamilyID:   familyID,
    GroupID:    groupID, // Прив'язка
    Name:       req.Name,
    URL:        req.URL,
    Price:      req.Price,
    Currency:   req.Currency,
    Priority:   req.Priority,
    Status:     "planning",
    Visibility: req.Visibility,
    HiddenFrom: req.HiddenFrom,
  }
  item.ID = uuid.New().String()

  if item.Visibility == "" {
    item.Visibility = "public"
  }

  err := s.repo.Create(item)
  return item, err
}

// ToggleReservation перемикає стан резерву
func (s *WishlistService) ToggleReservation(itemID string, userID string) error {
  var item models.WishlistItem

  // 1. Шукаємо товар напряму через DB (бо стандартний GetByID вимагає familyID)
  // Використовуємо s.repo.GetDB(), який ми щойно додали
  if err := s.repo.GetDB().Where("id = ?", itemID).First(&item).Error; err != nil {
    return errors.New("item not found")
  }

  // 2. Не можна резервувати своє бажання
  if item.UserID == userID {
    return errors.New("cannot reserve your own item")
  }

  // 3. Якщо вже кимось зарезервовано
  if item.ReservedByUserID != nil {
    // Якщо це я зарезервував -> Знімаємо резерв
    if *item.ReservedByUserID == userID {
      return s.repo.GetDB().Model(&models.WishlistItem{}).
        Where("id = ?", itemID).
        Update("reserved_by_user_id", nil).Error // nil знімає резерв
    }
    // Якщо хтось інший -> Помилка
    return errors.New("item is already reserved by someone else")
  }

  // 4. Якщо вільно -> Резервуємо
  return s.repo.GetDB().Model(&models.WishlistItem{}).
    Where("id = ?", itemID).
    Update("reserved_by_user_id", userID).Error
}

// GetItems з логікою "Сюрпризу"
func (s *WishlistService) GetItems(familyID, userID, groupID, targetUserID string) ([]models.WishlistItem, error) {
  // 1. Отримуємо всі записи з бази
  items, err := s.repo.GetAll(familyID, userID, groupID, targetUserID)
  if err != nil {
    return nil, err
  }

  // 2. 🔥 ФІЛЬТР СЮРПРИЗУ: Проходимось по списку і ховаємо резерв від власника
  for i := range items {
    // Якщо я дивлюсь СВІЙ список (item.UserID == userID)
    // То я не маю бачити, що поле ReservedByUserID заповнене.
    if items[i].UserID == userID {
      items[i].ReservedByUserID = nil
    }
  }

  return items, nil
}

func (s *WishlistService) GetItem(id string, familyID string) (*models.WishlistItem, error) {
  return s.repo.GetByID(id, familyID)
}
func (s *WishlistService) UploadPhoto(id string, familyID string, fileReader io.Reader) (string, error) {
  item, err := s.repo.GetByID(id, familyID)
  if err != nil {
    return "", errors.New("wishlist item not found")
  }

  // Правильне видалення старого фото
  if item.PhotoURL != "" {
    relativePath := strings.TrimPrefix(item.PhotoURL, "/")
    _ = os.Remove(relativePath)
  }

  // Створюємо папку, якщо її немає
  uploadDir := "uploads/wishlist"
  if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
    return "", fmt.Errorf("could not create upload directory: %w", err)
  }
  
  newFileName := uuid.New().String() + ".jpg"
  fullPath := filepath.Join(uploadDir, newFileName)
  webPath := "/uploads/wishlist/" + newFileName

  out, err := os.Create(fullPath)
  if err != nil {
    return "", err
  }
  defer out.Close()

  // 🔥 ПРОСТО КОПІЮЄМО ФАЙЛ БЕЗ ПОДВІЙНОГО СТИСНЕННЯ 🔥
  if _, err := io.Copy(out, fileReader); err != nil {
    return "", err
  }

  err = s.repo.Update(id, familyID, map[string]interface{}{
    "photo_url": webPath,
  })

  return webPath, err
}

func (s *WishlistService) UpdateItem(id string, req models.UpdateWishlistRequest, familyID string) (*models.WishlistItem, error) {
  updates := make(map[string]interface{})

  if req.Name != nil { updates["name"] = *req.Name }
  if req.URL != nil { updates["url"] = *req.URL }
  if req.Price != nil { updates["price"] = *req.Price }
  if req.Currency != nil { updates["currency"] = *req.Currency }
  if req.Status != nil { updates["status"] = *req.Status }
  if req.Visibility != nil { updates["visibility"] = *req.Visibility }
  if req.HiddenFrom != nil { updates["hidden_from"] = *req.HiddenFrom }
  
  if req.Priority != nil {
    if *req.Priority >= 1 && *req.Priority <= 3 {
      updates["priority"] = *req.Priority
    }
  }

  // Update GroupID
  if req.GroupID != nil {
    if *req.GroupID == "" {
      updates["group_id"] = nil // Видалити з групи
    } else {
      updates["group_id"] = *req.GroupID // Перемістити в групу
    }
  }

  // Update GoalID
  if req.GoalID != nil {
    if *req.GoalID == "" {
      updates["goal_id"] = nil
    } else {
      updates["goal_id"] = *req.GoalID
    }
  }

  if len(updates) > 0 {
    if err := s.repo.Update(id, familyID, updates); err != nil {
      return nil, err
    }
  }
  
  return s.repo.GetByID(id, familyID)
}

// DeleteItem видаляє запис і фото (ТІЛЬКИ ВЛАСНИК)
func (s *WishlistService) DeleteItem(id string, familyID, userID string) error {
  // 1. Спочатку отримуємо запис
  item, err := s.repo.GetByID(id, familyID)
  if err != nil {
    return err
  }

  // 2. Перевіряємо, чи це власник
  if item.UserID != userID {
    return errors.New("you do not have permission to delete this item")
  }

  // 3. ВИПРАВЛЕНО: Правильне видалення фото, якщо є
  if item.PhotoURL != "" {
    relativePath := strings.TrimPrefix(item.PhotoURL, "/")
    _ = os.Remove(relativePath)
  }

  // 4. Видаляємо запис з БД
  return s.repo.Delete(id, familyID)
}

func (s *WishlistService) RemovePhoto(id string, familyID string) error {
  item, err := s.repo.GetByID(id, familyID)
  if err != nil {
    return err
  }
  
  // ВИПРАВЛЕНО: Правильне видалення фото
  if item.PhotoURL != "" {
    relativePath := strings.TrimPrefix(item.PhotoURL, "/")
    _ = os.Remove(relativePath)
  }
  
  return s.repo.Update(id, familyID, map[string]interface{}{"photo_url": ""})
}