package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateUserInput struct {
	Name     string
	Email    string
	Password string
	RoleID   string
}

type UserService interface {
	GetMe(id string) (*models.User, error)
	GetFamilyMembers(user *models.User) ([]models.User, error)
	// AddMember приймає ініціатора (actor)
	AddMember(actor *models.User, input CreateUserInput) (*models.User, error)
	UpdateProfile(id string, name string, email string) (*models.User, error)
	ChangePassword(id string, oldPwd, newPwd string) error
	// DeleteMember приймає ініціатора
	DeleteMember(actor *models.User, targetID string) error
	// UpdateUser приймає ініціатора
	UpdateUser(actor *models.User, targetID string, input CreateUserInput) (*models.User, error)
	LeaveFamily(user *models.User) error
}

type userService struct {
	repo  repositories.UserRepository
	wsHub *utils.WSHub
	db    *gorm.DB
}

func NewUserService(repo repositories.UserRepository, wsHub *utils.WSHub, db *gorm.DB) UserService {
	return &userService{
		repo:  repo,
		wsHub: wsHub,
		db:    db,
	}
}

var ErrUserPermission = errors.New("permission denied: only parents can manage members")

func (s *userService) GetMe(id string) (*models.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) GetFamilyMembers(user *models.User) ([]models.User, error) {
	// Дитина може бачити список, це ок
	return s.repo.GetFamilyMembers(user.FamilyID)
}

func (s *userService) AddMember(actor *models.User, input CreateUserInput) (*models.User, error) {
	// 🛑 ЗАХИСТ: Дитина не може додавати людей
	if actor.RoleID == "child" {
		return nil, ErrUserPermission
	}

	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Якщо роль не передана, за замовчуванням child (безпечніше)
	roleID := input.RoleID
	if roleID == "" {
		roleID = "child"
	}

	newUser := &models.User{
		Base: models.Base{
			ID:        uuid.NewString(),
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
			IsSynced:  true,
		},
		FamilyID:     actor.FamilyID,
		RoleID:       roleID,
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: hashed,
	}

	if err := s.repo.Create(newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *userService) UpdateUser(actor *models.User, targetID string, input CreateUserInput) (*models.User, error) {
	// 🛑 ЗАХИСТ: Дитина не може редагувати інших
	if actor.RoleID == "child" {
		return nil, ErrUserPermission
	}

	targetUser, err := s.repo.GetByID(targetID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Перевірка, чи ми редагуємо члена своєї сім'ї
	if targetUser.FamilyID != actor.FamilyID {
		return nil, errors.New("access denied")
	}

	targetUser.Name = input.Name
	targetUser.Email = input.Email

	if input.RoleID != "" {
		targetUser.RoleID = input.RoleID
	}

	if input.Password != "" {
		newHash, err := utils.HashPassword(input.Password)
		if err != nil {
			return nil, err
		}
		targetUser.PasswordHash = newHash
	}

	targetUser.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.Update(targetUser); err != nil {
		return nil, err
	}
	return targetUser, nil
}

func (s *userService) UpdateProfile(id string, name string, email string) (*models.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Кожен може змінювати своє ім'я
	user.Name = name
	user.Email = email
	user.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) ChangePassword(id string, oldPwd, newPwd string) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if !utils.CheckPassword(oldPwd, user.PasswordHash) {
		return errors.New("invalid old password")
	}

	newHash, err := utils.HashPassword(newPwd)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	user.UpdatedAt = time.Now().UnixMilli()

	return s.repo.Update(user)
}

func (s *userService) DeleteMember(actor *models.User, targetID string) error {
	// 🛑 ЗАХИСТ
	if actor.RoleID == "child" {
		return ErrUserPermission
	}

	if actor.ID == targetID {
		return errors.New("cannot delete yourself here. Use Leave instead")
	}

	targetUser, err := s.repo.GetByID(targetID)
	if err != nil {
		return errors.New("user not found")
	}

	if targetUser.FamilyID != actor.FamilyID {
		return errors.New("access denied")
	}

	oldFamilyID := targetUser.FamilyID

	// Використовуємо логіку "розлучення" (перенесення в нову персональну сім'ю)
	if err := s.separateUserFromFamily(targetUser, oldFamilyID); err != nil {
		return err
	}

	// Сповіщення про видалення через WebSocket в СТАРУ сім'ю
	s.wsHub.BroadcastToFamily(oldFamilyID, map[string]interface{}{
		"type":    "member_removed",
		"user_id": targetID,
		"message": "A member has been removed from the family",
	})

	return nil
}

func (s *userService) LeaveFamily(user *models.User) error {
	// Перевірка, чи це не єдиний член сім'ї
	count, err := s.repo.CountFamilyMembers(user.FamilyID)
	if err != nil {
		return err
	}

	if count <= 1 {
		return errors.New("you are the only member of this family. No need to leave")
	}

	oldFamilyID := user.FamilyID

	if err := s.separateUserFromFamily(user, oldFamilyID); err != nil {
		return err
	}

	// Сповіщення в СТАРУ сім'ю
	s.wsHub.BroadcastToFamily(oldFamilyID, map[string]interface{}{
		"type":    "member_removed",
		"user_id": user.ID,
		"message": "A member has left the family",
	})

	return nil
}

// separateUserFromFamily створює нову сім'ю для юзера і переносить його дані туди
func (s *userService) separateUserFromFamily(targetUser *models.User, oldFamilyID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Створюємо нову персональну сім'ю
		newFamily := &models.Family{
			Base: models.Base{
				ID:        uuid.NewString(),
				CreatedAt: time.Now().UnixMilli(),
				UpdatedAt: time.Now().UnixMilli(),
			},
			Name: fmt.Sprintf("%s's Personal Family", targetUser.Name),
		}

		if err := tx.Create(newFamily).Error; err != nil {
			return err
		}

		// 2. Оновлюємо користувача (новий FamilyID + роль Admin)
		if err := tx.Model(&models.User{}).Where("id = ?", targetUser.ID).Updates(map[string]interface{}{
			"family_id":  newFamily.ID,
			"role_id":    "admin", // Тепер він сам собі адмін
			"updated_at": time.Now().UnixMilli(),
		}).Error; err != nil {
			return err
		}

		// 3. Мігруємо дані (тільки ті, що мають user_id)
		tablesWithUserID := []string{
			"accounts",
			"transactions",
			"goals",
			"assets",
			"shopping_lists",
			"wishlist_groups",
			"wishlist_items",
			"medical_records",
			"bank_connections",
		}

		for _, table := range tablesWithUserID {
			if err := tx.Table(table).Where("user_id = ?", targetUser.ID).Update("family_id", newFamily.ID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
