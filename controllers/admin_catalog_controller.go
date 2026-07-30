package controllers

import (
	"net/http"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminCatalogController struct {
	db           *gorm.DB
	auditService services.AuditService
}

func NewAdminCatalogController(db *gorm.DB, auditService services.AuditService) *AdminCatalogController {
	return &AdminCatalogController{db: db, auditService: auditService}
}

type adminCategoryInput struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Icon      string `json:"icon"`
	Color     string `json:"color"`
	ParentID  string `json:"parent_id"`
	SystemKey string `json:"system_key"`
}

type adminCounterpartyCategoryInput struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Icon      string `json:"icon"`
	Color     string `json:"color"`
	SystemKey string `json:"system_key"`
}
type adminCounterpartyInput struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Icon       string  `json:"icon"`
	Logo       string  `json:"logo"`
	SystemKey  string  `json:"system_key"`
	CategoryID *string `json:"category_id"`
}

func (h *AdminCatalogController) CreateCategory(c *gin.Context) {
	var in adminCategoryInput
	if err := c.ShouldBindJSON(&in); err != nil || in.Name == "" || in.Type == "" || in.SystemKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, type and system_key are required"})
		return
	}
	now := time.Now().UnixMilli()
	category := models.Category{Base: models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true}, Name: in.Name, Type: in.Type, Icon: in.Icon, Color: in.Color, ParentID: in.ParentID, SystemKey: in.SystemKey}
	if err := h.db.Create(&category).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	adminID := c.GetString("userID")
	h.auditService.LogAction(adminID, "create_category", "global_category", category.ID, in, c.ClientIP())
	c.JSON(http.StatusCreated, category)
}

func (h *AdminCatalogController) UpdateCategory(c *gin.Context) {
	var in adminCategoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var category models.Category
	if err := h.db.Where("id = ? AND family_id = ?", c.Param("id"), "").First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}
	category.Name, category.Type, category.Icon, category.Color, category.ParentID, category.SystemKey, category.UpdatedAt = in.Name, in.Type, in.Icon, in.Color, in.ParentID, in.SystemKey, time.Now().UnixMilli()
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&category).Error; err != nil {
			return err
		}
		return tx.Model(&models.Category{}).Where("global_template_id = ?", category.ID).Updates(map[string]interface{}{"name": category.Name, "type": category.Type, "icon": category.Icon, "color": category.Color, "updated_at": category.UpdatedAt}).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	adminID := c.GetString("userID")
	h.auditService.LogAction(adminID, "update_category", "global_category", category.ID, in, c.ClientIP())
	c.JSON(http.StatusOK, category)
}

func (h *AdminCatalogController) GetCategories(c *gin.Context) {
	var categories []models.Category
	if err := h.db.Model(&models.Category{}).
		Select("categories.*, COUNT(transactions.id) AS usage_count").
		Joins("LEFT JOIN categories AS local_categories ON local_categories.global_template_id = categories.id AND local_categories.family_id <> '' AND local_categories.deleted_at IS NULL").
		Joins("LEFT JOIN transactions ON transactions.category_id = local_categories.id AND transactions.deleted_at IS NULL").
		Where("categories.family_id = ? AND categories.deleted_at IS NULL", "").
		Group("categories.id").
		Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h *AdminCatalogController) ArchiveCategory(c *gin.Context) {
	if err := h.db.Model(&models.Category{}).Where("id = ? AND family_id = ?", c.Param("id"), "").Update("is_archived", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	adminID := c.GetString("userID")
	h.auditService.LogAction(adminID, "archive_category", "global_category", c.Param("id"), nil, c.ClientIP())
	c.Status(http.StatusNoContent)
}

func (h *AdminCatalogController) GetCounterpartyCategories(c *gin.Context) {
	var categories []models.CounterpartyCategory
	if err := h.db.Model(&models.CounterpartyCategory{}).
		Select("counterparty_categories.*, COUNT(transactions.id) AS usage_count").
		Joins("LEFT JOIN counterparty_categories AS local_categories ON local_categories.global_template_id = counterparty_categories.id AND local_categories.family_id <> '' AND local_categories.deleted_at IS NULL").
		Joins("LEFT JOIN counterparties AS local_counterparties ON local_counterparties.category_id = local_categories.id AND local_counterparties.deleted_at IS NULL").
		Joins("LEFT JOIN transactions ON transactions.counterparty_id = local_counterparties.id AND transactions.deleted_at IS NULL").
		Where("counterparty_categories.family_id = ? AND counterparty_categories.deleted_at IS NULL", "").
		Group("counterparty_categories.id").
		Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h *AdminCatalogController) CreateCounterpartyCategory(c *gin.Context) {
	var in adminCounterpartyCategoryInput
	if err := c.ShouldBindJSON(&in); err != nil || in.Name == "" || in.Type == "" || in.SystemKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, type and system_key are required"})
		return
	}
	now := time.Now().UnixMilli()
	value := models.CounterpartyCategory{Base: models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true}, Name: in.Name, Type: in.Type, Icon: in.Icon, Color: in.Color, SystemKey: in.SystemKey}
	if err := h.db.Create(&value).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	adminID := c.GetString("userID")
	h.auditService.LogAction(adminID, "create_counterparty_category", "global_counterparty_category", value.ID, in, c.ClientIP())
	c.JSON(http.StatusCreated, value)
}

func (h *AdminCatalogController) UpdateCounterpartyCategory(c *gin.Context) {
	var in adminCounterpartyCategoryInput
	if err := c.ShouldBindJSON(&in); err != nil || in.Name == "" || in.Type == "" || in.SystemKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, type and system_key are required"})
		return
	}
	var category models.CounterpartyCategory
	if err := h.db.Where("id = ? AND family_id = ?", c.Param("id"), "").First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "counterparty category not found"})
		return
	}
	category.Name, category.Type, category.Icon, category.Color, category.SystemKey, category.UpdatedAt = in.Name, in.Type, in.Icon, in.Color, in.SystemKey, time.Now().UnixMilli()
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&category).Error; err != nil {
			return err
		}
		return tx.Model(&models.CounterpartyCategory{}).Where("global_template_id = ?", category.ID).Updates(map[string]interface{}{
			"name": category.Name, "type": category.Type, "icon": category.Icon, "color": category.Color, "updated_at": category.UpdatedAt,
		}).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	adminID := c.GetString("userID")
	h.auditService.LogAction(adminID, "update_counterparty_category", "global_counterparty_category", category.ID, in, c.ClientIP())
	c.JSON(http.StatusOK, category)
}

func (h *AdminCatalogController) ArchiveCounterpartyCategory(c *gin.Context) {
	if err := h.db.Model(&models.CounterpartyCategory{}).Where("id = ? AND family_id = ?", c.Param("id"), "").Update("is_archived", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	adminID := c.GetString("userID")
	h.auditService.LogAction(adminID, "archive_counterparty_category", "global_counterparty_category", c.Param("id"), nil, c.ClientIP())
	c.Status(http.StatusNoContent)
}

func (h *AdminCatalogController) GetCounterparties(c *gin.Context) {
	var counterparties []models.Counterparty
	if err := h.db.Model(&models.Counterparty{}).
		Select("counterparties.*, COUNT(transactions.id) AS usage_count").
		Joins("LEFT JOIN counterparties AS local_counterparties ON local_counterparties.global_template_id = counterparties.id AND local_counterparties.family_id <> '' AND local_counterparties.deleted_at IS NULL").
		Joins("LEFT JOIN transactions ON transactions.counterparty_id = local_counterparties.id AND transactions.deleted_at IS NULL").
		Where("counterparties.family_id = ? AND counterparties.deleted_at IS NULL", "").
		Group("counterparties.id").
		Preload("Category").
		Find(&counterparties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, counterparties)
}

func (h *AdminCatalogController) CreateCounterparty(c *gin.Context) {
	var in adminCounterpartyInput
	if err := c.ShouldBindJSON(&in); err != nil || in.Name == "" || in.Type == "" || in.SystemKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, type and system_key are required"})
		return
	}
	now := time.Now().UnixMilli()
	value := models.Counterparty{Base: models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now, IsSynced: true}, Name: in.Name, Type: in.Type, Icon: in.Icon, Logo: in.Logo, CategoryID: in.CategoryID, SystemKey: in.SystemKey}
	if err := h.db.Create(&value).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	adminID := c.GetString("userID")
	h.auditService.LogAction(adminID, "create_counterparty", "global_counterparty", value.ID, in, c.ClientIP())
	c.JSON(http.StatusCreated, value)
}

func (h *AdminCatalogController) UpdateCounterparty(c *gin.Context) {
	var in adminCounterpartyInput
	if err := c.ShouldBindJSON(&in); err != nil || in.Name == "" || in.Type == "" || in.SystemKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, type and system_key are required"})
		return
	}
	var counterparty models.Counterparty
	if err := h.db.Where("id = ? AND family_id = ?", c.Param("id"), "").First(&counterparty).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "counterparty not found"})
		return
	}
	counterparty.Name, counterparty.Type, counterparty.Icon, counterparty.Logo, counterparty.CategoryID, counterparty.SystemKey, counterparty.UpdatedAt = in.Name, in.Type, in.Icon, in.Logo, in.CategoryID, in.SystemKey, time.Now().UnixMilli()
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&counterparty).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Counterparty{}).Where("global_template_id = ?", counterparty.ID).Updates(map[string]interface{}{
			"name": counterparty.Name, "type": counterparty.Type, "icon": counterparty.Icon, "logo": counterparty.Logo, "updated_at": counterparty.UpdatedAt,
		}).Error; err != nil {
			return err
		}
		if in.CategoryID == nil {
			return tx.Model(&models.Counterparty{}).Where("global_template_id = ?", counterparty.ID).Update("category_id", nil).Error
		}
		return tx.Exec(`UPDATE counterparties
			SET category_id = (SELECT id FROM counterparty_categories WHERE global_template_id = ? AND family_id = counterparties.family_id LIMIT 1)
			WHERE global_template_id = ?`, *in.CategoryID, counterparty.ID).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	adminID := c.GetString("userID")
	h.auditService.LogAction(adminID, "update_counterparty", "global_counterparty", counterparty.ID, in, c.ClientIP())
	c.JSON(http.StatusOK, counterparty)
}

func (h *AdminCatalogController) ArchiveCounterparty(c *gin.Context) {
	if err := h.db.Model(&models.Counterparty{}).Where("id = ? AND family_id = ?", c.Param("id"), "").Update("is_archived", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	adminID := c.GetString("userID")
	h.auditService.LogAction(adminID, "archive_counterparty", "global_counterparty", c.Param("id"), nil, c.ClientIP())
	c.Status(http.StatusNoContent)
}
