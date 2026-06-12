package services

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/google/uuid" // <--- ДОДАЛИ ІМПОРТ
)

type ShoppingService interface {
	CreateList(req models.CreateShoppingListRequest, userID, familyID string) (*models.ShoppingList, error)
	GetLists(familyID string, userID string) ([]models.ShoppingList, error)
	UpdateList(id string, req models.UpdateShoppingListRequest, familyID string) error
	DeleteList(id string, familyID string) error
	AddItemToList(listID string, req models.CreateShoppingItemRequest) (*models.ShoppingItem, error)
	UpdateItem(id string, req models.UpdateShoppingItemRequest) error
	DeleteItem(id string) error
	ClearCompletedInList(listID string) error
}

type shoppingService struct {
	repo *repositories.ShoppingRepo
}

func NewShoppingService(repo *repositories.ShoppingRepo) ShoppingService {
	return &shoppingService{repo: repo}
}

// --- LISTS ---
func (s *shoppingService) CreateList(req models.CreateShoppingListRequest, userID, familyID string) (*models.ShoppingList, error) {
	visibility := "public"
	if req.Visibility != "" {
		visibility = req.Visibility
	}

	list := &models.ShoppingList{
		UserID:     userID,
		FamilyID:   familyID,
		Title:      req.Title,
		Color:      req.Color,
		Visibility: visibility,
		HiddenFrom: req.HiddenFrom,
	}
	
	// 🔥 ЯВНО ГЕНЕРУЄМО ID
	list.ID = uuid.NewString()

	err := s.repo.CreateList(list)
	return list, err
}

func (s *shoppingService) GetLists(familyID string, userID string) ([]models.ShoppingList, error) {
	return s.repo.GetLists(familyID, userID)
}

func (s *shoppingService) UpdateList(id string, req models.UpdateShoppingListRequest, familyID string) error {
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.Visibility != nil {
		updates["visibility"] = *req.Visibility
	}
	if req.HiddenFrom != nil {
		updates["hidden_from"] = *req.HiddenFrom
	}

	if len(updates) > 0 {
		return s.repo.UpdateList(id, familyID, updates)
	}
	return nil
}

func (s *shoppingService) DeleteList(id string, familyID string) error {
	return s.repo.DeleteList(id, familyID)
}

// --- ITEMS ---
func (s *shoppingService) AddItemToList(listID string, req models.CreateShoppingItemRequest) (*models.ShoppingItem, error) {
	item := &models.ShoppingItem{
		ListID:   listID,
		Name:     req.Name,
		IsBought: false,
	}
	
	// 🔥 ЯВНО ГЕНЕРУЄМО ID
	item.ID = uuid.NewString()

	err := s.repo.CreateItem(item)
	return item, err
}

func (s *shoppingService) UpdateItem(id string, req models.UpdateShoppingItemRequest) error {
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.IsBought != nil {
		updates["is_bought"] = *req.IsBought
	}

	if len(updates) > 0 {
		return s.repo.UpdateItem(id, updates)
	}
	return nil
}

func (s *shoppingService) DeleteItem(id string) error {
	return s.repo.DeleteItem(id)
}

func (s *shoppingService) ClearCompletedInList(listID string) error {
	return s.repo.ClearCompletedInList(listID)
}