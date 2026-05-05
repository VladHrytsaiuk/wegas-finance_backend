package controllers

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	service services.TransactionService
}
type CreateTxJSON struct {
  AccountID       string `json:"account_id" binding:"required"`
  TargetAccountID string `json:"target_account_id"`
  CategoryID      string `json:"category_id"`
  CounterpartyID  string `json:"counterparty_id"`

  Amount int64 `json:"amount" binding:"required"`

  // 🔥 Поле для трансферів
  TargetAmount *int64 `json:"target_amount"`

  Date int64  `json:"date" binding:"required"`
  Note string `json:"note"`
  Type string `json:"type" binding:"required"`

  // 🔥 НОВЕ ПОЛЕ: Прощення боргу
  IsForgiveness bool `json:"is_forgiveness"`

  Items  []TxItemJSON `json:"items"`
  TagIDs []string     `json:"tag_ids"`

  AssetID *string `json:"asset_id"`
  
  // 🔥🔥🔥 ДОДАНО: Без цього поля контролер ігнорував пробіг!
  Mileage *int    `json:"mileage"` 

  NewAsset *models.CreateAssetOnFlyInput `json:"new_asset"`
}

type TxItemJSON struct {
	Name         string  `json:"name" binding:"required"`
	Quantity     int64   `json:"quantity" binding:"required"`
	PricePerUnit int64   `json:"price_per_unit" binding:"required"`
	TotalAmount  int64   `json:"total_amount" binding:"required"`
	CategoryID   *string `json:"category_id"`
}

func NewTransactionController(service services.TransactionService) *TransactionController {
	return &TransactionController{service: service}
}

// === HANDLERS ===

func (h *TransactionController) Create(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var jsonInput CreateTxJSON
	var files []*multipart.FileHeader

	contentType := c.GetHeader("Content-Type")

	// 1. Обробка Multipart (файли + json)
	if strings.Contains(contentType, "multipart/form-data") {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Multipart form error: " + err.Error()})
			return
		}

		jsonData := c.PostForm("json")
		if jsonData == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'json' form field"})
			return
		}

		if err := json.Unmarshal([]byte(jsonData), &jsonInput); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}

		files = form.File["files"]
		if len(files) == 0 {
			if f, ok := form.File["file"]; ok && len(f) > 0 {
				files = f
			}
		}

	} else {
		// 2. Обробка звичайного JSON
		if err := c.ShouldBindJSON(&jsonInput); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Мапінг Items
	var itemsInput []services.TransactionItemInput
	for _, it := range jsonInput.Items {
		itemsInput = append(itemsInput, services.TransactionItemInput{
			Name:         it.Name,
			Quantity:     it.Quantity,
			PricePerUnit: it.PricePerUnit,
			TotalAmount:  it.TotalAmount,
			CategoryID:   it.CategoryID,
		})
	}

	input := services.CreateTransactionInput{
		AccountID:       jsonInput.AccountID,
		TargetAccountID: jsonInput.TargetAccountID,
		CategoryID:      jsonInput.CategoryID,
		CounterpartyID:  jsonInput.CounterpartyID,
		Amount:          jsonInput.Amount,
		TargetAmount:    jsonInput.TargetAmount,
		Date:            jsonInput.Date,
		Note:            jsonInput.Note,
		Type:            jsonInput.Type,
		Items:           itemsInput,
		TagIDs:          jsonInput.TagIDs,
		AssetID:         jsonInput.AssetID,
		NewAsset:        jsonInput.NewAsset,
		Mileage:         jsonInput.Mileage,

		// 🔥 ПЕРЕДАЄМО ПРАПОРЕЦЬ У СЕРВІС
		IsForgiveness: jsonInput.IsForgiveness,
	}

	txID, err := h.service.Create(input, files, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "id": txID})
}

func (h *TransactionController) UploadReceipt(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer file.Close()

	path, err := h.service.UploadReceipt(id, file, header, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"path": path})
}

func (h *TransactionController) GetAll(c *gin.Context) {
    currentUser, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    user := currentUser.(*models.User)

    // Вручну беремо параметри з сирого запиту, щоб уникнути проблем Gin
    queryParams := c.Request.URL.Query()

    filter := repositories.TransactionFilter{
        FamilyID:        user.FamilyID,
        // Спробуємо взяти і з [] і без них
        AccountIDs:      append(queryParams["account_id"], queryParams["account_id[]"]...),
        CategoryIDs:     append(queryParams["category_id"], queryParams["category_id[]"]...),
        CounterpartyIDs: append(queryParams["counterparty_id"], queryParams["counterparty_id[]"]...),
        Type:            c.Query("type"),
        Search:          c.Query("search"),
        Sort:            c.Query("sort"),
        AssetID:         c.Query("asset_id"),
    }

    // Парсинг пагінації
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    filter.Limit = limit
    filter.Offset = (page - 1) * limit

    // Парсинг дат
    if dateFrom, err := strconv.ParseInt(c.Query("date_from"), 10, 64); err == nil {
        filter.DateFrom = dateFrom
    }
    if dateTo, err := strconv.ParseInt(c.Query("date_to"), 10, 64); err == nil {
        filter.DateTo = dateTo
    }

    // Парсинг сум
    if minAmt, err := strconv.ParseInt(c.Query("min_amount"), 10, 64); err == nil {
        filter.MinAmount = &minAmt
    }
    if maxAmt, err := strconv.ParseInt(c.Query("max_amount"), 10, 64); err == nil {
        filter.MaxAmount = &maxAmt
    }

    txs, count, err := h.service.GetAll(filter, user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "data":  txs,
        "count": count,
        "page":  page,
        "limit": limit,
    })
}

func (h *TransactionController) GetOne(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	tx, err := h.service.GetByID(id, user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}
	c.JSON(http.StatusOK, tx)
}

func (h *TransactionController) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	if err := h.service.Delete(id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted"})
}

func (h *TransactionController) Update(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var jsonInput CreateTxJSON
	if err := c.ShouldBindJSON(&jsonInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var itemsInput []services.TransactionItemInput
	for _, it := range jsonInput.Items {
		itemsInput = append(itemsInput, services.TransactionItemInput{
			Name:         it.Name,
			Quantity:     it.Quantity,
			PricePerUnit: it.PricePerUnit,
			TotalAmount:  it.TotalAmount,
			CategoryID:   it.CategoryID,
		})
	}

	input := services.CreateTransactionInput{
		AccountID:       jsonInput.AccountID,
		TargetAccountID: jsonInput.TargetAccountID,
		CategoryID:      jsonInput.CategoryID,
		CounterpartyID:  jsonInput.CounterpartyID,
		Amount:          jsonInput.Amount,
		// TargetAmount:    jsonInput.TargetAmount, // Update для переказу - це складно, поки не чіпаємо
		Date:     jsonInput.Date,
		Note:     jsonInput.Note,
		Type:     jsonInput.Type,
		Items:    itemsInput,
		TagIDs:   jsonInput.TagIDs,
		AssetID:  jsonInput.AssetID,
		Mileage:         jsonInput.Mileage,

		// 🔥 ПЕРЕДАЄМО ПРАПОРЕЦЬ ПРИ ОНОВЛЕННІ (Якщо раптом захочете змінити)
		IsForgiveness: jsonInput.IsForgiveness,
	}

	if err := h.service.Update(id, input, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *TransactionController) DeleteReceipt(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	if err := h.service.DeleteReceipt(id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Receipt deleted successfully"})
}

func (h *TransactionController) BatchCreate(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	var jsonInput []CreateTxJSON
	if err := c.ShouldBindJSON(&jsonInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format: " + err.Error()})
		return
	}

	if len(jsonInput) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No transactions provided"})
		return
	}

	var inputs []services.CreateTransactionInput
	for _, item := range jsonInput {
		var itemsInput []services.TransactionItemInput
		for _, it := range item.Items {
			itemsInput = append(itemsInput, services.TransactionItemInput{
				Name:         it.Name,
				Quantity:     it.Quantity,
				PricePerUnit: it.PricePerUnit,
				TotalAmount:  it.TotalAmount,
				CategoryID:   it.CategoryID,
			})
		}

		inputs = append(inputs, services.CreateTransactionInput{
			AccountID:       item.AccountID,
			TargetAccountID: item.TargetAccountID,
			CategoryID:      item.CategoryID,
			CounterpartyID:  item.CounterpartyID,
			Amount:          item.Amount,
			TargetAmount:    item.TargetAmount,
			Date:            item.Date,
			Note:            item.Note,
			Type:            item.Type,
			Items:           itemsInput,
			TagIDs:          item.TagIDs,
			AssetID:         item.AssetID,
			Mileage:         item.Mileage,
			// 🔥 ПЕРЕДАЄМО ПРАПОРЕЦЬ В BATCH
			IsForgiveness: item.IsForgiveness,
		})
	}

	count, err := h.service.BatchCreate(inputs, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "count": count})
}

func (h *TransactionController) DeletePhoto(c *gin.Context) {
	photoID := c.Param("id")
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	if err := h.service.DeletePhoto(photoID, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Photo deleted successfully"})
}

func (h *TransactionController) PredictCategory(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(*models.User)

	itemName := c.Query("name")
	if itemName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	categoryID, err := h.service.PredictCategory(itemName, user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"category_id": nil})
		return
	}

	if categoryID == "" {
		c.JSON(http.StatusOK, gin.H{"category_id": nil})
	} else {
		c.JSON(http.StatusOK, gin.H{"category_id": categoryID})
	}
}



