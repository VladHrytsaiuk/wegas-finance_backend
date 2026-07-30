package services

import (
	"encoding/json"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type AuditService interface {
	LogAction(adminID, action, entityType, entityID string, changes interface{}, ipAddress string) error
	GetLogs(limit, offset int) ([]models.AuditLog, int64, error)
}

type auditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) AuditService {
	return &auditService{db: db}
}

func (s *auditService) LogAction(adminID, action, entityType, entityID string, changes interface{}, ipAddress string) error {
	changesJSON, _ := json.Marshal(changes)

	log := models.AuditLog{
		ID:         uuid.NewString(),
		AdminID:    adminID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Changes:    string(changesJSON),
		IPAddress:  ipAddress,
	}

	return s.db.Create(&log).Error
}

func (s *auditService) GetLogs(limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var count int64
	var eg errgroup.Group

	eg.Go(func() error {
		return s.db.Model(&models.AuditLog{}).Count(&count).Error
	})

	eg.Go(func() error {
		return s.db.Preload("Admin").Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error
	})

	if err := eg.Wait(); err != nil {
		return nil, 0, err
	}

	return logs, count, nil
}
