package services

import (
	"errors"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"gorm.io/gorm"
)

type AdminService interface {
	GetUsers(limit, offset int, search string) ([]models.User, int64, error)
	ToggleUserBlock(userID string) error
	ForceLogoutUser(userID string) error
	SetUserPlatformAdmin(userID string, isPlatformAdmin bool) error

	GetSettings() (map[string]string, error)
}

type adminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) AdminService {
	return &adminService{db: db}
}

func (s *adminService) GetUsers(limit, offset int, search string) ([]models.User, int64, error) {
	var users []models.User
	var count int64
	query := s.db.Model(&models.User{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("email LIKE ? OR name LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Family").Order("created_at desc").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, count, nil
}

func (s *adminService) ToggleUserBlock(userID string) error {
	var user models.User
	if err := s.db.Select("id", "is_active", "session_version", "is_platform_admin").First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	// Prevent blocking the last active platform admin
	if user.IsActive && user.IsPlatformAdmin {
		var activeAdminsCount int64
		if err := s.db.Model(&models.User{}).Where("is_platform_admin = ? AND is_active = ? AND deleted_at IS NULL", true, true).Count(&activeAdminsCount).Error; err != nil {
			return err
		}
		if activeAdminsCount <= 1 {
			return errors.New("Cannot block the last active platform admin")
		}
	}

	updates := map[string]interface{}{
		"is_active": !user.IsActive,
	}

	if !user.IsActive {
		updates["session_version"] = gorm.Expr("session_version + ?", 1)
	}

	return s.db.Model(&user).Updates(updates).Error
}

func (s *adminService) ForceLogoutUser(userID string) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).UpdateColumn("session_version", gorm.Expr("session_version + ?", 1)).Error
}

func (s *adminService) SetUserPlatformAdmin(userID string, isPlatformAdmin bool) error {
	if !isPlatformAdmin {
		var activeAdminsCount int64
		if err := s.db.Model(&models.User{}).Where("is_platform_admin = ? AND is_active = ? AND deleted_at IS NULL", true, true).Count(&activeAdminsCount).Error; err != nil {
			return err
		}
		
		var user models.User
		if err := s.db.Select("id", "is_platform_admin").First(&user, "id = ?", userID).Error; err != nil {
			return err
		}
		
		if user.IsPlatformAdmin && activeAdminsCount <= 1 {
			return errors.New("Cannot remove the last active platform admin")
		}
	}
	return s.db.Model(&models.User{}).Where("id = ?", userID).UpdateColumn("is_platform_admin", isPlatformAdmin).Error
}

func (s *adminService) GetSettings() (map[string]string, error) {
	var settings []models.SystemSetting
	if err := s.db.Find(&settings).Error; err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, setting := range settings {
		m[setting.Key] = setting.Value
	}
	return m, nil
}
