// services/user_service.go
package services

import (
	"context"
	"fmt"
	"log"

	"good_blast/database"
	"good_blast/errors"
	"good_blast/models"

	"github.com/google/uuid"
)

// UserService implements UserServiceInterface.
type UserService struct {
	DB database.DatabaseInterface
}

// NewUserService creates a new instance of UserService.
func NewUserService(db database.DatabaseInterface) *UserService {
	return &UserService{
		DB: db,
	}
}

// CreateUser handles user creation logic.
func (s *UserService) CreateUser(ctx context.Context, username, country string) (*models.User, error) {
	// Generate a unique userId
	userId := uuid.New().String()

	// Initialize the user
	user := models.User{
		UserID:   userId,
		Username: username,
		Level:    1,
		Coins:    1000,
		Country:  country,
		GlobalPK: "GLOBAL",
	}

	// Save user to DynamoDB
	if err := s.DB.PutUser(ctx, user); err != nil {
		log.Println("Error creating user:", err)
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	return &user, nil
}

// GetUser retrieves user details by userID.
func (s *UserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.DB.GetUser(ctx, userID)
	if err != nil {
		log.Println("Error fetching user:", err)
		return nil, fmt.Errorf("could not fetch user: %w", err)
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

// UpdateUserProgress updates the user's level and coins based on progress.
func (s *UserService) UpdateUserProgress(ctx context.Context, userID string, newLevel int) (*models.User, error) {
	// Fetch current user data
	user, err := s.DB.GetUser(ctx, userID)
	if err != nil {
		log.Println("Error fetching user:", err)
		return nil, fmt.Errorf("could not fetch user data: %w", err)
	}

	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	// Validate that newLevel is greater than current level
	if newLevel <= user.Level {
		return nil, errors.ErrInvalidLevelIncrease
	}

	// Calculate coins gained
	levelIncrement := newLevel - user.Level
	coinsGained := levelIncrement * 100

	newCoins := user.Coins + coinsGained

	// Update user in DynamoDB
	if err := s.DB.UpdateUserCoinsAndLevel(ctx, userID, newLevel, newCoins); err != nil {
		log.Println("Error updating user progress:", err)
		return nil, fmt.Errorf("could not update user progress: %w", err)
	}

	// Fetch updated user data
	updatedUser, err := s.DB.GetUser(ctx, userID)
	if err != nil {
		log.Println("Error fetching updated user:", err)
		return nil, fmt.Errorf("could not fetch updated user data: %w", err)
	}

	return updatedUser, nil
}
