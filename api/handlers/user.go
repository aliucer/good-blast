// api/handlers/user.go
package handlers

import (
	"log"
	"net/http"

	"good_blast/errors"
	"good_blast/services"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	Service services.UserServiceInterface
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(service services.UserServiceInterface) *UserHandler {
	return &UserHandler{
		Service: service,
	}
}

// createUserRequest defines the expected payload for creating a user.
type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	Country  string `json:"country,omitempty"`
}

// CreateUser handles user creation requests.
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	ctx := c.Request.Context() // Extract context from the HTTP request

	// Create the user
	user, err := h.Service.CreateUser(ctx, req.Username, req.Country)
	if err != nil {
		log.Println("Error creating user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":   user.UserID,
		"username": user.Username,
		"level":    user.Level,
		"coins":    user.Coins,
		"country":  user.Country,
	})
}

// updateProgressRequest defines the expected payload for updating user progress.
type updateProgressRequest struct {
	NewLevel int `json:"newLevel" binding:"required"`
}

// UpdateProgress handles user progress updates.
func (h *UserHandler) UpdateProgress(c *gin.Context) {
	userID := c.Param("userId")

	var req updateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "newLevel is required and must be an integer"})
		return
	}

	ctx := c.Request.Context() // Extract context from the HTTP request

	// Update user progress
	updatedUser, err := h.Service.UpdateUserProgress(ctx, userID, req.NewLevel)
	if err != nil {
		log.Println("UpdateProgress error:", err)
		// Determine the type of error
		if err == errors.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		} else if err.Error() == "newLevel must be greater than current level" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "newLevel must be greater than current level"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update user progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":   updatedUser.UserID,
		"username": updatedUser.Username,
		"level":    updatedUser.Level,
		"coins":    updatedUser.Coins,
		"country":  updatedUser.Country,
	})
}
