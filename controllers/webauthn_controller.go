package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/gin-gonic/gin"
)

type WebAuthnController struct {
	waService  services.WebAuthnService
	jwtService services.JWTService
	userRepo   repositories.UserRepository
}

func NewWebAuthnController(waService services.WebAuthnService, jwtService services.JWTService, userRepo repositories.UserRepository) *WebAuthnController {
	return &WebAuthnController{
		waService:  waService,
		jwtService: jwtService,
		userRepo:   userRepo,
	}
}

// POST /api/webauthn/register/options
func (ctrl *WebAuthnController) RegisterOptions(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := ctrl.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	options, sessionID, err := ctrl.waService.BeginRegistration(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"options":    options,
		"session_id": sessionID,
	})
}

// POST /api/webauthn/register/verify
func (ctrl *WebAuthnController) RegisterVerify(c *gin.Context) {
	userID := c.GetString("userID")
	var body struct {
		SessionID string          `json:"session_id" binding:"required"`
		Response  json.RawMessage `json:"response" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ctrl.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Отримуємо Origin з запиту
	origin := c.Request.Header.Get("Origin")

	if err := ctrl.waService.FinishRegistration(user, body.SessionID, body.Response, origin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// POST /api/webauthn/login/options
func (ctrl *WebAuthnController) LoginOptions(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	options, sessionID, err := ctrl.waService.BeginLogin(body.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"options":    options,
		"session_id": sessionID,
	})
}

// POST /api/webauthn/login/verify
func (ctrl *WebAuthnController) LoginVerify(c *gin.Context) {
	var body struct {
		SessionID string          `json:"session_id" binding:"required"`
		Response  json.RawMessage `json:"response" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Отримуємо Origin з запиту
	origin := c.Request.Header.Get("Origin")

	user, err := ctrl.waService.FinishLogin(body.SessionID, body.Response, origin)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Issue Tokens
	accessToken, err := ctrl.jwtService.GenerateAccessToken(user.ID, user.FamilyID, user.RoleID, user.SessionVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := ctrl.jwtService.GenerateRefreshToken(user.ID, user.FamilyID, user.RoleID, user.SessionVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	setRefreshTokenCookie(c, refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"user":         user,
	})
}

// POST /api/refresh
func (ctrl *WebAuthnController) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token missing"})
		return
	}

	claims, err := ctrl.jwtService.ValidateToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	user, err := ctrl.userRepo.GetByID(claims.UserID)
	if err != nil || !user.IsActive || user.SessionVersion != claims.SessionVersion {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired or user blocked"})
		return
	}

	// Generate new Access Token
	accessToken, err := ctrl.jwtService.GenerateAccessToken(user.ID, user.FamilyID, user.RoleID, user.SessionVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}
