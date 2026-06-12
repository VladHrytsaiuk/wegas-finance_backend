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

type WishlistService interface {
	CreateGroup(name, color, icon, visibility, hiddenFrom, userID, familyID string) (*models.WishlistGroup, error)
	GetGroups(familyID, myUserID string) ([]models.WishlistGroup, error)
	UpdateGroup(id string, req models.UpdateWishlistGroupRequest, familyID string) error
	DeleteGroup(id, familyID string) error
	CreateItem(req models.CreateWishlistRequest, userID, familyID string) (*models.WishlistItem, error)
	ToggleReservation(itemID string, userID string) error
	GetItems(familyID, userID, groupID, targetUserID string) ([]models.WishlistItem, error)
	GetItem(id string, familyID string) (*models.WishlistItem, error)
	UploadPhoto(id string, familyID string, fileReader io.Reader) (string, error)
	UpdateItem(id string, req models.UpdateWishlistRequest, familyID string) (*models.WishlistItem, error)
	DeleteItem(id string, familyID, userID string) error
	RemovePhoto(id string, familyID string) error
}

type wishlistService struct {
	repo *repositories.WishlistRepo
}

func NewWishlistService(repo *repositories.WishlistRepo) WishlistService {
	return &wishlistService{repo: repo}
}

// --- GROUPS LOGIC ---

func (s *wishlistService) CreateGroup(name, color, icon, visibility, hiddenFrom, userID, familyID string) (*models.WishlistGroup, error) {
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

func (s *wishlistService) GetGroups(familyID, myUserID string) ([]models.WishlistGroup, error) {
	return s.repo.GetGroups(familyID, myUserID)
}

func (s *wishlistService) UpdateGroup(id string, req models.UpdateWishlistGroupRequest, familyID string) error {
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
	updates["hidden_from"] = req.HiddenFrom

	if len(updates) > 0 {
		return s.repo.UpdateGroup(id, familyID, updates)
	}
	return nil
}

func (s *wishlistService) DeleteGroup(id, familyID string) error {
	return s.repo.DeleteGroup(id, familyID)
}

// --- ITEMS LOGIC ---

func (s *wishlistService) CreateItem(req models.CreateWishlistRequest, userID, familyID string) (*models.WishlistItem, error) {
	if req.Priority < 1 || req.Priority > 3 {
		req.Priority = 1
	}

	var groupID *string
	if req.GroupID != "" {
		groupID = &req.GroupID
	}

	item := &models.WishlistItem{
		UserID:     userID,
		FamilyID:   familyID,
		GroupID:    groupID,
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

func (s *wishlistService) ToggleReservation(itemID string, userID string) error {
	var item models.WishlistItem

	if err := s.repo.GetDB().Where("id = ?", itemID).First(&item).Error; err != nil {
		return errors.New("item not found")
	}

	if item.UserID == userID {
		return errors.New("cannot reserve your own item")
	}

	if item.ReservedByUserID != nil {
		if *item.ReservedByUserID == userID {
			return s.repo.GetDB().Model(&models.WishlistItem{}).
				Where("id = ?", itemID).
				Update("reserved_by_user_id", nil).Error
		}
		return errors.New("item is already reserved by someone else")
	}

	return s.repo.GetDB().Model(&models.WishlistItem{}).
		Where("id = ?", itemID).
		Update("reserved_by_user_id", userID).Error
}

func (s *wishlistService) GetItems(familyID, userID, groupID, targetUserID string) ([]models.WishlistItem, error) {
	items, err := s.repo.GetAll(familyID, userID, groupID, targetUserID)
	if err != nil {
		return nil, err
	}

	for i := range items {
		if items[i].UserID == userID {
			items[i].ReservedByUserID = nil
		}
	}

	return items, nil
}

func (s *wishlistService) GetItem(id string, familyID string) (*models.WishlistItem, error) {
	return s.repo.GetByID(id, familyID)
}

func (s *wishlistService) UploadPhoto(id string, familyID string, fileReader io.Reader) (string, error) {
	item, err := s.repo.GetByID(id, familyID)
	if err != nil {
		return "", errors.New("wishlist item not found")
	}

	if item.PhotoURL != "" {
		relativePath := strings.TrimPrefix(item.PhotoURL, "/")
		_ = os.Remove(relativePath)
	}

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

	if _, err := io.Copy(out, fileReader); err != nil {
		return "", err
	}

	err = s.repo.Update(id, familyID, map[string]interface{}{
		"photo_url": webPath,
	})

	return webPath, err
}

func (s *wishlistService) UpdateItem(id string, req models.UpdateWishlistRequest, familyID string) (*models.WishlistItem, error) {
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Visibility != nil {
		updates["visibility"] = *req.Visibility
	}
	if req.HiddenFrom != nil {
		updates["hidden_from"] = *req.HiddenFrom
	}

	if req.Priority != nil {
		if *req.Priority >= 1 && *req.Priority <= 3 {
			updates["priority"] = *req.Priority
		}
	}

	if req.GroupID != nil {
		if *req.GroupID == "" {
			updates["group_id"] = nil
		} else {
			updates["group_id"] = *req.GroupID
		}
	}

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

func (s *wishlistService) DeleteItem(id string, familyID, userID string) error {
	item, err := s.repo.GetByID(id, familyID)
	if err != nil {
		return err
	}

	if item.UserID != userID {
		return errors.New("you do not have permission to delete this item")
	}

	if item.PhotoURL != "" {
		relativePath := strings.TrimPrefix(item.PhotoURL, "/")
		_ = os.Remove(relativePath)
	}

	return s.repo.Delete(id, familyID)
}

func (s *wishlistService) RemovePhoto(id string, familyID string) error {
	item, err := s.repo.GetByID(id, familyID)
	if err != nil {
		return err
	}

	if item.PhotoURL != "" {
		relativePath := strings.TrimPrefix(item.PhotoURL, "/")
		_ = os.Remove(relativePath)
	}

	return s.repo.Update(id, familyID, map[string]interface{}{"photo_url": ""})
}
