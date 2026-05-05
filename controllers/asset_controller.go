package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type AssetController struct {
  service *services.AssetService
}

func NewAssetController(s *services.AssetService) *AssetController {
  return &AssetController{service: s}
}

// GET /assets
func (h *AssetController) GetAll(c *gin.Context) {
  user := c.MustGet("user").(*models.User)
  assets, err := h.service.GetAll(user)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, assets)
}

// POST /assets
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

// GET /assets/:id
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

// PUT /assets/:id
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

func (h *AssetController) UpdateMileage(c *gin.Context) {
  id := c.Param("id")
  user := c.MustGet("user").(*models.User)

  var input struct {
    Mileage int `json:"mileage" binding:"required"`
  }
  
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

// DELETE /assets/:id
func (h *AssetController) Delete(c *gin.Context) {
  user := c.MustGet("user").(*models.User)
  id := c.Param("id")
  if err := h.service.Delete(id, user); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// POST /assets/:id/photo
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

// DELETE /assets/:id/photo
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

// 🔥 POST /assets/:id/documents
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

// 🔥 DELETE /assets/:id/documents/:doc_id
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