// api/routes.go
package api

import (
	"good_blast/api/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up all the API routes with their respective handlers.
func SetupRoutes(router *gin.Engine, userHandler *handlers.UserHandler, tournamentHandler *handlers.TournamentHandler, leaderboardHandler *handlers.LeaderboardHandler) {
	// User routes
	router.POST("/users", userHandler.CreateUser)
	router.PUT("/users/:userId/progress", userHandler.UpdateProgress)

	// Tournament routes
	router.POST("/tournaments/start", tournamentHandler.StartTournamentHandler)
	router.PUT("/tournaments/end/:tournamentId", tournamentHandler.EndTournamentHandler)
	router.POST("/tournaments/enter", tournamentHandler.EnterTournament)
	router.PUT("/tournaments/:tournamentId/score", tournamentHandler.UpdateScore)
	router.POST("/tournaments/:tournamentId/claim", tournamentHandler.ClaimReward)

	// Leaderboard routes
	router.GET("/leaderboard/global", leaderboardHandler.GetGlobalLeaderboard)
	router.GET("/leaderboard/country", leaderboardHandler.GetCountryLeaderboard)
	router.GET("/leaderboard/tournament", leaderboardHandler.GetTournamentLeaderboard)
	router.GET("/tournaments/:tournamentId/rank", leaderboardHandler.GetTournamentRank)
}
