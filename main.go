// main.go
package main

import (
	"fmt"
	"log"
	"os"

	"good_blast/api"
	"good_blast/api/handlers"
	"good_blast/database"
	"good_blast/services"

	"github.com/gin-gonic/gin"
)

// APIResponse defines a standard API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// initializeApp initializes the database, services, handlers, and router.
// It returns the router for further use.
func initializeApp() (*handlers.UserHandler, *handlers.TournamentHandler, *handlers.LeaderboardHandler, *gin.Engine, error) {
	// Initialize DynamoDB
	if err := database.InitDynamoDB(); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize DynamoDB: %w", err)
	}

	db := &database.DynamoDB{}
	userService := services.NewUserService(db)
	tournamentService := services.NewTournamentService(db)
	leaderboardService := services.NewLeaderboardService(db)

	userHandler := handlers.NewUserHandler(userService)
	tournamentHandler := handlers.NewTournamentHandler(tournamentService)
	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardService)

	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Setup routes with handlers
	api.SetupRoutes(router, userHandler, tournamentHandler, leaderboardHandler)

	return userHandler, tournamentHandler, leaderboardHandler, router, nil
}

func main() {
	// Set Gin to Release mode for production
	gin.SetMode(gin.ReleaseMode)

	// Initialize the application
	_, _, _, router, err := initializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Determine the port to listen on
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	// Start the server
	log.Printf("Starting server on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
