package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type InboxController struct {
	service services.InboxService
}

type InboxCreateItemJSON struct {
	Name         string  `json:"name"`
	Quantity     int64   `json:"quantity"`
	PricePerUnit int64   `json:"price_per_unit"`
	TotalAmount  int64   `json:"total_amount"`
	CategoryID   *string `json:"category_id"`
}

type InboxCreateJSON struct {
	Status string `json:"status"`
	Reason string `json:"reason"`

	SelectedAccountID *string `json:"selected_account_id"`
	ReviewRequired    *bool   `json:"review_required"`

	SourceType string `json:"source_type" binding:"required"`
	Origin     string `json:"origin"`

	FilePath  string `json:"file_path"`
	SourceURL string `json:"source_url"`
	MimeType  string `json:"mime_type"`

	RawPayload    string `json:"raw_payload"`
	ParsedPayload string `json:"parsed_payload"`

	Merchant      string `json:"merchant"`
	ReceiptNumber string `json:"receipt_number"`
	ReceiptDate   *int64 `json:"receipt_date"`
	Subtotal      *int64 `json:"subtotal"`
	DiscountTotal *int64 `json:"discount_total"`
	Total         *int64 `json:"total"`
	Currency      string `json:"currency"`
	OccurredAt    *int64 `json:"occurred_at"`
	Note          string `json:"note"`

	CounterpartyID *string               `json:"counterparty_id"`
	CategoryID     *string               `json:"category_id"`
	Items          []InboxCreateItemJSON `json:"items"`
}

type InboxSelectAccountJSON struct {
	AccountID string `json:"account_id" binding:"required"`
}

type InboxLinkJSON struct {
	TransactionID        string `json:"transaction_id" binding:"required"`
	ApplyItems           *bool  `json:"apply_items"`
	LearnFromTransaction bool   `json:"learn_from_transaction"`
}

func NewInboxController(service services.InboxService) *InboxController {
	return &InboxController{service: service}
}

// Create godoc
// @Summary Create inbox entry
// @Description Creates a new inbox entry and receipt source payload.
// @Tags Inbox
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body InboxCreateJSON true "Inbox entry payload"
// @Success 201 {object} models.InboxEntry
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /inbox [post]
func (h *InboxController) Create(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var input InboxCreateJSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []services.InboxCreateItemInput
	for _, item := range input.Items {
		items = append(items, services.InboxCreateItemInput{
			Name:         item.Name,
			Quantity:     item.Quantity,
			PricePerUnit: item.PricePerUnit,
			TotalAmount:  item.TotalAmount,
			CategoryID:   item.CategoryID,
		})
	}

	entry, err := h.service.Create(services.InboxCreateInput{
		Status:            input.Status,
		Reason:            input.Reason,
		SelectedAccountID: input.SelectedAccountID,
		ReviewRequired:    input.ReviewRequired,
		SourceType:        input.SourceType,
		Origin:            input.Origin,
		FilePath:          input.FilePath,
		SourceURL:         input.SourceURL,
		MimeType:          input.MimeType,
		RawPayload:        input.RawPayload,
		ParsedPayload:     input.ParsedPayload,
		Merchant:          input.Merchant,
		ReceiptNumber:     input.ReceiptNumber,
		ReceiptDate:       input.ReceiptDate,
		Subtotal:          input.Subtotal,
		DiscountTotal:     input.DiscountTotal,
		Total:             input.Total,
		Currency:          input.Currency,
		OccurredAt:        input.OccurredAt,
		Note:              input.Note,
		CounterpartyID:    input.CounterpartyID,
		CategoryID:        input.CategoryID,
		Items:             items,
	}, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// GetAll godoc
// @Summary Get inbox entries
// @Description Returns receipt-driven inbox entries that still require linking or review.
// @Tags Inbox
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "Comma-separated inbox statuses"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /inbox [get]
func (h *InboxController) GetAll(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var statuses []string
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		for _, item := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				statuses = append(statuses, trimmed)
			}
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	entries, total, err := h.service.GetAll(services.InboxListFilter{
		Status: statuses,
		Limit:  limit,
		Offset: offset,
	}, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  entries,
		"total": total,
	})
}

// GetOne godoc
// @Summary Get inbox entry by ID
// @Description Returns one inbox entry with receipt details.
// @Tags Inbox
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Inbox entry ID"
// @Success 200 {object} models.InboxEntry
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Inbox entry not found"
// @Router /inbox/{id} [get]
func (h *InboxController) GetOne(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	entry, err := h.service.GetByID(c.Param("id"), user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inbox entry not found"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// GetAccountCandidates godoc
// @Summary Find account candidates from the receipt payment mask
// @Description Matches the trailing 1-4 digits of EПЗ against physical and device-token card masks.
// @Tags Inbox
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Inbox entry ID"
// @Success 200 {array} services.InboxAccountCandidate
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Inbox entry not found"
// @Router /inbox/{id}/account-candidates [get]
func (h *InboxController) GetAccountCandidates(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	candidates, err := h.service.FindAccountCandidates(c.Param("id"), currentUser.(*models.User))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inbox entry not found"})
		return
	}

	c.JSON(http.StatusOK, candidates)
}

// GetTransactionCandidates godoc
// @Summary Find bank transaction candidates for an inbox receipt
// @Description Returns unmatched synced transactions with the selected account and exact receipt amount.
// @Tags Inbox
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Inbox entry ID"
// @Success 200 {array} services.InboxTransactionCandidate
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Inbox entry not found"
// @Router /inbox/{id}/transaction-candidates [get]
func (h *InboxController) GetTransactionCandidates(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	candidates, err := h.service.FindTransactionCandidates(c.Param("id"), currentUser.(*models.User))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inbox entry not found"})
		return
	}

	c.JSON(http.StatusOK, candidates)
}

// SelectAccount godoc
// @Summary Select account for inbox entry
// @Description Attaches a selected account to an inbox entry and advances its basic state.
// @Tags Inbox
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Inbox entry ID"
// @Param body body InboxSelectAccountJSON true "Account selection payload"
// @Success 200 {object} models.InboxEntry
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Inbox entry not found"
// @Router /inbox/{id}/account [patch]
func (h *InboxController) SelectAccount(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var input InboxSelectAccountJSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.service.SelectAccount(c.Param("id"), input.AccountID, user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inbox entry not found"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// Link godoc
// @Summary Link inbox entry to transaction
// @Description Links a receipt inbox entry to an existing transaction and optionally applies parsed items.
// @Tags Inbox
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Inbox entry ID"
// @Param body body InboxLinkJSON true "Link payload"
// @Success 200 {object} models.InboxEntry
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Inbox entry not found"
// @Router /inbox/{id}/link [post]
func (h *InboxController) Link(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var input InboxLinkJSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	applyItems := true
	if input.ApplyItems != nil {
		applyItems = *input.ApplyItems
	}

	entry, err := h.service.Link(c.Param("id"), input.TransactionID, applyItems, input.LearnFromTransaction, user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// Unlink godoc
// @Summary Unlink receipt from transaction
// @Description Removes a receipt link and returns the inbox entry back to active inbox state.
// @Tags Inbox
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Inbox entry ID"
// @Success 200 {object} models.InboxEntry
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Inbox entry not found"
// @Router /inbox/{id}/unlink [post]
func (h *InboxController) Unlink(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	entry, err := h.service.Unlink(c.Param("id"), user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}
