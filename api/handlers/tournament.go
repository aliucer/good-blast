// api/handlers/tournament.go
package handlers

import (
	"log"
	"net/http"
	"time"

	"good_blast/errors"
	"good_blast/services"

	"github.com/gin-gonic/gin"
)

// TournamentHandler handles tournament-related HTTP requests.
type TournamentHandler struct {
	Service services.TournamentServiceInterface
}

// NewTournamentHandler creates a new instance of TournamentHandler.
func NewTournamentHandler(service services.TournamentServiceInterface) *TournamentHandler {
	return &TournamentHandler{
		Service: service,
	}
}

// StartTournamentHandler creates a new daily tournament and marks it active.
func (h *TournamentHandler) StartTournamentHandler(c *gin.Context) {
	ctx := c.Request.Context() // Extract context from the HTTP request

	tournament, err := h.Service.StartTournament(ctx)
	if err != nil {
		log.Println("Error starting tournament:", err)
		if err == errors.ErrAlreadyInTournament {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tournament already active for today"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not start tournament"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Tournament started",
		"tournamentId": tournament.TournamentID,
		"startTime":    tournament.StartTime,
		"endTime":      tournament.EndTime,
		"active":       tournament.Active,
	})
}

// EndTournamentHandler sets active = false for the specified tournament.
func (h *TournamentHandler) EndTournamentHandler(c *gin.Context) {
	tournamentID := c.Param("tournamentId")
	ctx := c.Request.Context() // Extract context from the HTTP request

	err := h.Service.EndTournament(ctx, tournamentID)
	if err != nil {
		log.Println("Error ending tournament:", err)
		if err == errors.ErrTournamentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "tournament not found"})
			return
		} else if err == errors.ErrTournamentAlreadyInactive {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tournament is already inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not end tournament"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Tournament ended",
		"tournamentId": tournamentID,
		"active":       false,
	})
}

// EnterTournament allows a user to enter an active tournament before 12:00 UTC.
func (h *TournamentHandler) EnterTournament(c *gin.Context) {
	var req struct {
		UserID       string `json:"userId" binding:"required"`
		TournamentID string `json:"tournamentId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId and tournamentId are required"})
		return
	}

	ctx := c.Request.Context() // Extract context from the HTTP request

	nowUTC := time.Now().UTC()
	cutoff := time.Date(nowUTC.Year(), nowUTC.Month(), nowUTC.Day(), 12, 0, 0, 0, time.UTC)
	if nowUTC.After(cutoff) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot enter the tournament after 12:00 UTC"})
		return
	}

	remainingCoins, err := h.Service.EnterTournament(ctx, req.UserID, req.TournamentID)
	if err != nil {
		log.Println("EnterTournament error:", err)
		switch err {
		case errors.ErrTournamentNotActive:
			c.JSON(http.StatusBadRequest, gin.H{"error": "tournament is not active"})
		case errors.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.ErrUserLevelTooLow:
			c.JSON(http.StatusBadRequest, gin.H{"error": "user must be at least level 10"})
		case errors.ErrInsufficientCoins:
			c.JSON(http.StatusBadRequest, gin.H{"error": "not enough coins (need 500)"})
		case errors.ErrAlreadyInTournament:
			c.JSON(http.StatusBadRequest, gin.H{"error": "you are already in the tournament"})
		case errors.ErrRequirementsNotMet:
			c.JSON(http.StatusBadRequest, gin.H{"error": "you do not meet the requirements to enter the tournament"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "User entered tournament successfully",
		"userId":         req.UserID,
		"tournamentId":   req.TournamentID,
		"remainingCoins": remainingCoins,
	})
}

// UpdateScore increments a user's score during the active tournament.
func (h *TournamentHandler) UpdateScore(c *gin.Context) {
	tournamentID := c.Param("tournamentId")
	ctx := c.Request.Context() // Extract context from the HTTP request

	var req struct {
		UserID    string `json:"userId" binding:"required"`
		Increment int    `json:"increment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId and increment are required"})
		return
	}

	newScore, err := h.Service.UpdateScore(ctx, tournamentID, req.UserID, req.Increment)
	if err != nil {
		log.Println("UpdateScore error:", err)
		switch err {
		case errors.ErrTournamentEntryNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "tournament entry not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update tournament score"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Score updated successfully",
		"tournamentId": tournamentID,
		"userId":       req.UserID,
		"newScore":     newScore,
	})
}

// ClaimReward allows a user to claim their reward after the tournament has ended.
func (h *TournamentHandler) ClaimReward(c *gin.Context) {
	tournamentID := c.Param("tournamentId")
	var req struct {
		UserID string `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	ctx := c.Request.Context() // Extract context from the HTTP request

	rank, reward, err := h.Service.ClaimReward(ctx, tournamentID, req.UserID)
	if err != nil {
		log.Println("ClaimReward error:", err)
		switch err {
		case errors.ErrTournamentNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "tournament not found"})
		case errors.ErrTournamentEntryNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "no tournament entry found for this user"})
		case errors.ErrRewardAlreadyClaimed:
			c.JSON(http.StatusBadRequest, gin.H{"error": "reward has already been claimed"})
		case errors.ErrNoRewardForRank:
			c.JSON(http.StatusOK, gin.H{
				"message":      "No reward available for your rank in the group",
				"userId":       req.UserID,
				"tournamentId": tournamentID,
				"rank":         "beyond top 10 in group",
				"reward":       0,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not claim reward"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Reward claimed successfully",
		"userId":       req.UserID,
		"tournamentId": tournamentID,
		"rank":         rank,
		"reward":       reward,
	})
}
