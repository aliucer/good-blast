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
	redisclient "good_blast/services/redis_client" // give it a distinct alias

	"github.com/gin-gonic/gin"
)

// APIResponse defines a standard API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func initializeApp() (*handlers.UserHandler, *handlers.TournamentHandler, *handlers.LeaderboardHandler, *gin.Engine, error) {
	log.Println("initializeApp: Starting application initialization...")

	if err := database.InitDynamoDB(); err != nil {
		log.Printf("initializeApp: failed to initialize DynamoDB: %v", err)
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize DynamoDB: %w", err)
	}
	log.Println("initializeApp: DynamoDB initialized successfully")

	db := &database.DynamoDB{}
	log.Println("initializeApp: DynamoDB struct created")

	// Initialize Redis
	if err := redisclient.InitRedis(); err != nil {
		log.Fatalf("initializeApp: failed to initialize Redis: %v", err)
	}

	log.Println("initializeApp: Redis initialized successfully")

	userService := services.NewUserService(db)
	log.Println("initializeApp: UserService initialized")

	tournamentService := services.NewTournamentService(db)
	log.Println("initializeApp: TournamentService initialized")

	leaderboardService := services.NewLeaderboardService(db)
	log.Println("initializeApp: LeaderboardService initialized")

	userHandler := handlers.NewUserHandler(userService)
	log.Println("initializeApp: UserHandler initialized")

	tournamentHandler := handlers.NewTournamentHandler(tournamentService)
	log.Println("initializeApp: TournamentHandler initialized")

	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardService)
	log.Println("initializeApp: LeaderboardHandler initialized")

	router := gin.Default()
	log.Println("initializeApp: Gin router created")

	// CORS middleware
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
	log.Println("initializeApp: CORS middleware set")

	// Setup routes
	api.SetupRoutes(router, userHandler, tournamentHandler, leaderboardHandler)
	log.Println("initializeApp: Routes set up successfully")

	return userHandler, tournamentHandler, leaderboardHandler, router, nil
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	log.Println("main: Initializing application...")
	_, _, _, router, err := initializeApp()
	if err != nil {
		log.Fatalf("main: Failed to initialize application: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("main: Starting server on port %s...", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("main: Failed to run server: %v", err)
	}
}
