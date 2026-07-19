package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type ReceiptIngestionController struct {
	service services.ReceiptIngestionService
}

type ReceiptURLIngestionJSON struct {
	URL string `json:"url" binding:"required"`
}

func NewReceiptIngestionController(service services.ReceiptIngestionService) *ReceiptIngestionController {
	return &ReceiptIngestionController{service: service}
}

// IngestXML godoc
// @Summary Ingest XML receipt into inbox
// @Description Uploads an XML receipt file, parses it and creates an inbox entry.
// @Tags ReceiptIngestion
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param file formData file true "XML receipt file"
// @Success 201 {object} models.InboxEntry
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /receipt-ingestion/xml [post]
func (h *ReceiptIngestionController) IngestXML(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	entry, err := h.service.IngestXML(file, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// IngestURL godoc
// @Summary Ingest receipt URL into inbox
// @Description Fetches a supported receipt URL, parses it and creates an inbox entry.
// @Tags ReceiptIngestion
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body ReceiptURLIngestionJSON true "Receipt URL payload"
// @Success 201 {object} models.InboxEntry
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /receipt-ingestion/url [post]
func (h *ReceiptIngestionController) IngestURL(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var input ReceiptURLIngestionJSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.service.IngestURL(input.URL, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}
