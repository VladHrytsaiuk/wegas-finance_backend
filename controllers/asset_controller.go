package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type AssetController struct {
  service services.AssetService
}

func NewAssetController(s services.AssetService) *AssetController {
  return &AssetController{service: s}
}

// GetAll godoc
// @Summary Get all assets
// @Description Returns a list of all assets for the current user.
// @Tags Assets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.Asset
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /assets [get]
func (h *AssetController) GetAll(c *gin.Context) {
  user := c.MustGet("user").(*models.User)
  assets, err := h.service.GetAll(user)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, assets)
}

// Create godoc
// @Summary Create a new asset
// @Description Creates a new asset (property, vehicle, etc.).
// @Tags Assets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.Asset true "Asset details"
// @Success 201 {object} map[string]string "Created asset ID"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /assets [post]
func (h *AssetController) Create(c *gin.Context) {
  user := c.MustGet("user").(*models.User)
  
  var input models.Asset 
  if err := c.ShouldBindJSON(&input); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  
  id, err := h.service.Create(input, user)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetOne godoc
// @Summary Get asset by ID
// @Description Returns a single asset by its ID.
// @Tags Assets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Asset ID"
// @Success 200 {object} models.Asset
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Asset not found"
// @Router /assets/{id} [get]
func (h *AssetController) GetOne(c *gin.Context) {
  user := c.MustGet("user").(*models.User)
  id := c.Param("id")
  asset, err := h.service.GetByID(id, user)
  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
    return
  }
  c.JSON(http.StatusOK, asset)
}

// Update godoc
// @Summary Update asset
// @Description Updates an existing asset by its ID.
// @Tags Assets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Asset ID"
// @Param body body models.Asset true "Updated asset details"
// @Success 200 {object} map[string]string "Update status"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /assets/{id} [put]
func (h *AssetController) Update(c *gin.Context) {
  id := c.Param("id")
  user := c.MustGet("user").(*models.User)

  var input models.Asset
  if err := c.ShouldBindJSON(&input); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  if err := h.service.Update(id, input, user); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

type UpdateMileageJSON struct {
	Mileage int `json:"mileage" binding:"required"`
}

// UpdateMileage godoc
// @Summary Update vehicle mileage
// @Description Updates the mileage for a vehicle asset.
// @Tags Assets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Asset ID"
// @Param body body UpdateMileageJSON true "Mileage details"
// @Success 200 {object} map[string]string "Update status"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /assets/{id}/mileage [patch]
func (h *AssetController) UpdateMileage(c *gin.Context) {
  id := c.Param("id")
  user := c.MustGet("user").(*models.User)

  var input UpdateMileageJSON
  
  if err := c.ShouldBindJSON(&input); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  if err := h.service.UpdateMileage(id, input.Mileage, user); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"status": "mileage updated"})
}

// Delete godoc
// @Summary Delete asset
// @Description Deletes an asset by its ID.
// @Tags Assets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Asset ID"
// @Success 200 {object} map[string]string "Delete status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /assets/{id} [delete]
func (h *AssetController) Delete(c *gin.Context) {
  user := c.MustGet("user").(*models.User)
  id := c.Param("id")
  if err := h.service.Delete(id, user); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// UploadPhoto godoc
// @Summary Upload asset photo
// @Description Uploads a photo for an asset.
// @Tags Assets
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Asset ID"
// @Param file formData file true "Photo file"
// @Success 200 {object} map[string]string "Path to the uploaded photo"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /assets/{id}/photo [post]
func (h *AssetController) UploadPhoto(c *gin.Context) {
  user := c.MustGet("user").(*models.User)
  id := c.Param("id")
  file, header, err := c.Request.FormFile("file")
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
    return
  }
  path, err := h.service.UploadPhoto(id, file, header, user)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"path": path})
}

// RemovePhoto godoc
// @Summary Remove asset photo
// @Description Removes a specific photo from an asset.
// @Tags Assets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Asset ID"
// @Param path query string true "Path of the photo to remove"
// @Success 200 {object} map[string]string "Delete status"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /assets/{id}/photo [delete]
func (h *AssetController) RemovePhoto(c *gin.Context) {
  id := c.Param("id")
  user := c.MustGet("user").(*models.User)

  path := c.Query("path")
  if path == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "path query parameter is required"})
    return
  }

  if err := h.service.RemovePhoto(id, path, user); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"status": "photo deleted"})
}

// UploadDocument godoc
// @Summary Upload asset document
// @Description Uploads a document file for an asset.
// @Tags Assets
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Asset ID"
// @Param file formData file true "Document file"
// @Success 200 {object} models.AssetDocument
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /assets/{id}/documents [post]
func (h *AssetController) UploadDocument(c *gin.Context) {
  user := c.MustGet("user").(*models.User)
  id := c.Param("id")
  file, header, err := c.Request.FormFile("file")
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
    return
  }
  
  doc, err := h.service.UploadDocument(id, file, header, user)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  
  c.JSON(http.StatusOK, doc) // Повертаємо об'єкт документа, щоб фронтенд відразу міг його відрендерити
}

// RemoveDocument godoc
// @Summary Remove asset document
// @Description Removes a specific document from an asset.
// @Tags Assets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Asset ID"
// @Param doc_id path string true "Document ID"
// @Success 200 {object} map[string]string "Delete status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /assets/{id}/documents/{doc_id} [delete]
func (h *AssetController) RemoveDocument(c *gin.Context) {
  id := c.Param("id")
  docID := c.Param("doc_id")
  user := c.MustGet("user").(*models.User)

  if err := h.service.RemoveDocument(id, docID, user); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"status": "document deleted"})
}