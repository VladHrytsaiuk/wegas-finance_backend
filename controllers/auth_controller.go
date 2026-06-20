package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service services.AuthService
}

func NewAuthController(service services.AuthService) *AuthController {
	return &AuthController{service: service}
}

type RegisterJSON struct {
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"invite_code" binding:"required"` // 🔥
}

type LoginJSON struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user and create a family for them. Requires a valid invite code.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body RegisterJSON true "Registration details"
// @Success 201 {object} services.LoginResponse
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 403 {object} map[string]string "Invalid invite code"
// @Router /users [post]
func (h *AuthController) Register(c *gin.Context) {
	var json RegisterJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.service.Register(services.RegisterInput{
		Name:       json.Name,
		Email:      json.Email,
		Password:   json.Password,
		InviteCode: json.InviteCode,
	})

	if err != nil {
		if err.Error() == "invalid invite code" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Невірний код запрошення"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set Refresh Token as HttpOnly cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("refresh_token", res.RefreshToken, 30*24*3600, "/", "", true, true)

	c.JSON(http.StatusCreated, res)
}

// Login godoc
// @Summary Login a user
// @Description Authenticate a user and return a JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginJSON true "Login credentials"
// @Success 200 {object} services.LoginResponse
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Router /login [post]
func (h *AuthController) Login(c *gin.Context) {
	var json LoginJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.service.Login(services.LoginInput{
		Email:    json.Email,
		Password: json.Password,
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Невірний логін або пароль"})
		return
	}

	// Set Refresh Token as HttpOnly cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("refresh_token", res.RefreshToken, 30*24*3600, "/", "", true, true)

	c.JSON(http.StatusOK, res)
}

func (h *AuthController) SetPIN(c *gin.Context) {
	userID := c.GetString("userID")
	var body struct {
		PIN string `json:"pin" binding:"required,len=4"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ПІН має бути 4 цифри"})
		return
	}

	if err := h.service.SetPIN(userID, body.PIN); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *AuthController) LoginWithPIN(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
		PIN   string `json:"pin" binding:"required,len=4"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Невірні дані"})
		return
	}

	res, err := h.service.LoginWithPIN(body.Email, body.PIN)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set Refresh Token as HttpOnly cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("refresh_token", res.RefreshToken, 30*24*3600, "/", "", true, true)

	c.JSON(http.StatusOK, res)
}

func (h *AuthController) GetSecurityStatus(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	status, err := h.service.GetSecurityStatus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *AuthController) RemovePIN(c *gin.Context) {
	userID := c.GetString("userID")
	if err := h.service.RemovePIN(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *AuthController) RemovePasskeys(c *gin.Context) {
	userID := c.GetString("userID")
	if err := h.service.RemovePasskeys(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}