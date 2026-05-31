package controllers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type FeedbackController struct {
	service services.FeedbackService
}

func NewFeedbackController(service services.FeedbackService) *FeedbackController {
	return &FeedbackController{service: service}
}

// FeedbackInput godoc
type FeedbackInput struct {
	Name     string `json:"name" form:"name"`
	Contact  string `json:"contact" form:"contact"`
	Message  string `json:"message" form:"message" binding:"required"`
	Priority string `json:"priority" form:"priority"`
}

// Submit godoc
// @Summary Submit feedback
// @Description Submits user feedback with optional images. Sent to a Telegram bot.
// @Tags Feedback
// @Accept mpfd
// @Produce json
// @Param name formData string false "User name"
// @Param contact formData string false "User contact info"
// @Param message formData string true "Feedback message"
// @Param priority formData string false "Priority level (low, medium, high)" default(medium)
// @Param images formData file false "Optional images for feedback"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Message is required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /feedback [post]
func (h *FeedbackController) Submit(c *gin.Context) {
	// 1. Зчитуємо Multipart форму, щоб отримати доступ і до полів, і до файлів
	// 32 MB - ліміт пам'яті для парсингу форми
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		// Якщо це не multipart запит, пробуємо просто JSON (але тоді без файлів)
		// У вашому випадку, скоріш за все, фронт завжди шле FormData
	}

	name := c.PostForm("name")
	contact := c.PostForm("contact")
	message := c.PostForm("message")

	// 1. Зчитуємо пріоритет
	priority := c.PostForm("priority")

	// Валідація / Дефолтне значення
	if priority == "" {
		priority = "medium"
	}

	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}

	// 2. Отримуємо файли
	var imagesBytes [][]byte
	
	form, err := c.MultipartForm()
	if err == nil {
		// Шукаємо файли під ключем "images" (множина)
		files := form.File["images"]
		
		// Якщо не знайшли під "images", спробуємо "image" (для сумісності)
		if len(files) == 0 {
			files = form.File["image"]
		}

		for _, file := range files {
			f, err := file.Open()
			if err != nil {
				continue
			}
			
			content, err := io.ReadAll(f)
			f.Close() // Закриваємо файл одразу після читання
			
			if err == nil && len(content) > 0 {
				imagesBytes = append(imagesBytes, content)
			}
		}
	}

	// Дефолтні значення
	if name == "" {
		name = "Anonymous"
	}
	if contact == "" {
		contact = "N/A"
	}

	// 3. Відправляємо в сервіс
	if err := h.service.SendFeedback(name, contact, message, priority, imagesBytes); err != nil {
		fmt.Printf("Error sending feedback: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send feedback"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feedback sent successfully"})
}