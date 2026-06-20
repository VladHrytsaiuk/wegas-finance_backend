package controllers

import (
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/database"
	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	service services.UserService
}

func NewUserController(service services.UserService) *UserController {
	return &UserController{service: service}
}

// GetMe godoc
// @Summary Get current user profile
// @Description Returns the profile of the currently logged-in user.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /users/me [get]
func (h *UserController) GetMe(c *gin.Context) {
	// Беремо з контексту, це найшвидший спосіб
	user := c.MustGet("user").(*models.User)

	// Наповнюємо прапорці безпеки
	user.HasPassword = user.PasswordHash != ""
	user.HasPin = user.PinHash != ""

	var passkeyCount int64
	if database.DB != nil {
		database.DB.Model(&models.WebAuthnCredential{}).Where("user_id = ?", user.ID).Count(&passkeyCount)
	}
	user.HasPasskeys = passkeyCount > 0

	c.JSON(http.StatusOK, user)
}


// GetFamilyMembers godoc
// @Summary Get family members
// @Description Returns a list of all members in the current user's family.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.User
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users [get]
func (h *UserController) GetFamilyMembers(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	members, err := h.service.GetFamilyMembers(currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

type AddMemberJSON struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	RoleID   string `json:"role_id"`
}

// AddMember godoc
// @Summary Add a family member
// @Description Adds a new user to the current user's family. Only accessible by family parents/admins.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body AddMemberJSON true "Member details"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 403 {object} map[string]string "Permission denied"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /family/users [post]
func (h *UserController) AddMember(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	var json AddMemberJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Передаємо currentUser, щоб сервіс перевірив, чи це Parent
	newUser, err := h.service.AddMember(currentUser, services.CreateUserInput{
		Name:     json.Name,
		Email:    json.Email,
		Password: json.Password,
		RoleID:   json.RoleID,
	})

	if err != nil {
		handleUserError(c, err)
		return
	}
	c.JSON(http.StatusCreated, newUser)
}

type UpdateProfileJSON struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Description Updates the profile information for the currently authenticated user.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body UpdateProfileJSON true "Updated profile details"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me [put]
func (h *UserController) UpdateProfile(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	var json UpdateProfileJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Редагування власного профілю дозволено всім
	user, err := h.service.UpdateProfile(currentUser.ID, json.Name, json.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

type ChangePwdJSON struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword godoc
// @Summary Change current user password
// @Description Changes the password for the currently authenticated user.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body ChangePwdJSON true "Password details"
// @Success 200 {object} map[string]string "Password updated"
// @Failure 400 {object} map[string]string "Invalid input or old password"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /users/password [put]
func (h *UserController) ChangePassword(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	var json ChangePwdJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ChangePassword(currentUser.ID, json.OldPassword, json.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password updated"})
}

// DeleteMember godoc
// @Summary Delete a family member
// @Description Deletes a family member. Only accessible by family parents/admins.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string "User deleted"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 403 {object} map[string]string "Permission denied"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /users/{id} [delete]
func (h *UserController) DeleteMember(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	targetID := c.Param("id")

	if err := h.service.DeleteMember(currentUser, targetID); err != nil {
		handleUserError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User removed and moved to personal space"})
}

// LeaveFamily godoc
// @Summary Leave current family
// @Description Leaves the current family and moves to a personal space.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "Left family"
// @Failure 400 {object} map[string]string "Invalid request"
// @Router /family/leave [post]
func (h *UserController) LeaveFamily(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	if err := h.service.LeaveFamily(currentUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully left the family"})
}

type UpdateUserJSON struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	RoleID   string `json:"role_id"`
	Password string `json:"password"`
}

// UpdateUser godoc
// @Summary Update a family member
// @Description Updates information for a specific family member. Only accessible by family parents/admins.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Param body body UpdateUserJSON true "Updated member details"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 403 {object} map[string]string "Permission denied"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /users/{id} [put]
func (h *UserController) UpdateUser(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	targetID := c.Param("id")

	var json UpdateUserJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedUser, err := h.service.UpdateUser(currentUser, targetID, services.CreateUserInput{
		Name:     json.Name,
		Email:    json.Email,
		RoleID:   json.RoleID,
		Password: json.Password,
	})

	if err != nil {
		handleUserError(c, err)
		return
	}
	c.JSON(http.StatusOK, updatedUser)
}

func handleUserError(c *gin.Context, err error) {
	if err.Error() == "permission denied: only parents can manage members" {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}