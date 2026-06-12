package controllers

import (
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type ImportController struct {
	Service services.ImportService
}

func NewImportController(service services.ImportService) *ImportController {
	return &ImportController{Service: service}
}

// UploadStatement godoc
// @Summary Upload and parse bank statement
// @Description Uploads a PDF file, parses it, and returns a preview of transactions
// @Tags Import
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param file formData file true "Bank statement PDF"
// @Param account_id formData string true "Account ID to link transactions"
// @Param bank_type formData string true "Type of bank (privatbank, monobank)"
// @Success 200 {object} services.PreviewResult
// @Router /import/upload [post]
func (c *ImportController) UploadStatement(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(400, gin.H{"error": "File is required"})
		return
	}

	accountID := ctx.PostForm("account_id")
	if accountID == "" {
		ctx.JSON(400, gin.H{"error": "account_id is required"})
		return
	}

	bankType := ctx.PostForm("bank_type")
	if bankType == "" {
		ctx.JSON(400, gin.H{"error": "bank_type is required"})
		return
	}

	result, err := c.Service.ProcessFile(file, accountID, bankType)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to process file: " + err.Error()})
		return
	}

	ctx.JSON(200, result)
}