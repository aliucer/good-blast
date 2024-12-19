// api/handlers/leaderboard.go
package handlers

import (
	"log"
	"net/http"

	"good_blast/errors"
	"good_blast/services"

	"github.com/gin-gonic/gin"
)

// LeaderboardHandler handles leaderboard-related HTTP requests.
type LeaderboardHandler struct {
	Service services.LeaderboardServiceInterface
}

// NewLeaderboardHandler creates a new instance of LeaderboardHandler.
func NewLeaderboardHandler(service services.LeaderboardServiceInterface) *LeaderboardHandler {
	return &LeaderboardHandler{
		Service: service,
	}
}

// GetGlobalLeaderboard retrieves the top users globally based on level.
func (h *LeaderboardHandler) GetGlobalLeaderboard(c *gin.Context) {
	ctx := c.Request.Context()

	users, err := h.Service.GetGlobalLeaderboard(ctx)
	if err != nil {
		log.Println("Error retrieving global leaderboard:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve global leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": users,
		"count":       len(users),
	})
}

// GetCountryLeaderboard retrieves the top users in a specific country based on level.
func (h *LeaderboardHandler) GetCountryLeaderboard(c *gin.Context) {
	countryCode := c.Query("countryCode")
	if countryCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "countryCode is required"})
		return
	}

	ctx := c.Request.Context()

	users, err := h.Service.GetCountryLeaderboard(ctx, countryCode)
	if err != nil {
		log.Println("Error retrieving country leaderboard:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve country leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": users,
		"countryCode": countryCode,
		"count":       len(users),
	})
}

// GetTournamentLeaderboard retrieves the leaderboard for a specific tournament group.
func (h *LeaderboardHandler) GetTournamentLeaderboard(c *gin.Context) {
	groupId := c.Query("groupId")
	if groupId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "groupId query parameter is required"})
		return
	}

	ctx := c.Request.Context()

	entries, err := h.Service.GetTournamentLeaderboard(ctx, groupId)
	if err != nil {
		log.Println("Error retrieving tournament leaderboard:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve tournament leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groupId":     groupId,
		"leaderboard": entries,
		"count":       len(entries),
	})
}

// GetTournamentRank retrieves a user's rank in a specific tournament (by group).
func (h *LeaderboardHandler) GetTournamentRank(c *gin.Context) {
	tournamentId := c.Param("tournamentId")
	userId := c.Query("userId")

	if tournamentId == "" || userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tournamentId and userId are required"})
		return
	}

	ctx := c.Request.Context()

	rank, err := h.Service.GetTournamentRank(ctx, tournamentId, userId)
	if err != nil {
		log.Println("Error retrieving tournament rank:", err)
		switch err {
		case errors.ErrTournamentEntryNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found in the specified tournament"})
		case errors.ErrUserNotFoundInLeaderboard:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found in the leaderboard"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve tournament rank"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":       userId,
		"tournamentId": tournamentId,
		"rank":         rank,
	})
}
