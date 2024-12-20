package services_test

import (
	"context"
	"errors"
	"testing"

	"good_blast/models"
	"good_blast/services"
	"good_blast/services/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUser(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockDatabase)
	userService := services.NewUserService(mockDB)

	ctx := context.Background()
	username := "testuser"
	country := "US"

	mockDB.On("PutUser", mock.Anything, mock.AnythingOfType("models.User")).Return(nil)

	// Act
	user, err := userService.CreateUser(ctx, username, country)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, country, user.Country)
	assert.Equal(t, 1, user.Level)    // default level is 1
	assert.Equal(t, 1000, user.Coins) // default coins is 1000
	assert.NotEmpty(t, user.UserID)   // userId should be generated
	assert.Equal(t, "GLOBAL", user.GlobalPK)
	mockDB.AssertExpectations(t)
}

func TestGetUser_Found(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockDatabase)
	userService := services.NewUserService(mockDB)

	ctx := context.Background()
	userId := uuid.New().String()
	expectedUser := &models.User{
		UserID:   userId,
		Username: "existingUser",
		Level:    5,
		Coins:    2000,
		Country:  "FR",
		GlobalPK: "GLOBAL",
	}

	// Mock database response: user is found
	mockDB.On("GetUser", mock.Anything, userId).Return(expectedUser, nil)

	// Act
	user, err := userService.GetUser(ctx, userId)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser, user)
	mockDB.AssertExpectations(t)
}

func TestGetUser_NotFound(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockDatabase)
	userService := services.NewUserService(mockDB)

	ctx := context.Background()
	userId := "nonexistent"

	// Mock database response: user not found
	mockDB.On("GetUser", mock.Anything, userId).Return((*models.User)(nil), nil)

	// Act
	user, err := userService.GetUser(ctx, userId)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")
	mockDB.AssertExpectations(t)
}

func TestUpdateUserProgress_Success(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockDatabase)
	userService := services.NewUserService(mockDB)

	ctx := context.Background()
	userId := uuid.New().String()
	currentUser := &models.User{
		UserID:   userId,
		Username: "player1",
		Level:    5,
		Coins:    2000,
		Country:  "US",
		GlobalPK: "GLOBAL",
	}

	newLevel := 7
	levelIncrement := newLevel - currentUser.Level
	coinsGained := levelIncrement * 100
	expectedCoins := currentUser.Coins + coinsGained

	// Mock database calls:
	mockDB.On("GetUser", mock.Anything, userId).Return(currentUser, nil).Once()
	mockDB.On("UpdateUserCoinsAndLevel", mock.Anything, userId, newLevel, expectedCoins).Return(nil).Once()

	updatedUser := &models.User{
		UserID:   userId,
		Username: "player1",
		Level:    newLevel,
		Coins:    expectedCoins,
		Country:  "US",
		GlobalPK: "GLOBAL",
	}
	mockDB.On("GetUser", mock.Anything, userId).Return(updatedUser, nil).Once()

	// Act
	user, err := userService.UpdateUserProgress(ctx, userId, newLevel)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, newLevel, user.Level)
	assert.Equal(t, expectedCoins, user.Coins)
	mockDB.AssertExpectations(t)
}

func TestUpdateUserProgress_InvalidLevel(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockDatabase)
	userService := services.NewUserService(mockDB)

	ctx := context.Background()
	userId := uuid.New().String()
	currentUser := &models.User{
		UserID:   userId,
		Username: "player2",
		Level:    5,
		Coins:    1000,
		Country:  "US",
		GlobalPK: "GLOBAL",
	}

	// Trying to update to a lower or equal level than current
	newLevel := 5

	// Mock database calls:
	mockDB.On("GetUser", mock.Anything, userId).Return(currentUser, nil)

	// Act
	user, err := userService.UpdateUserProgress(ctx, userId, newLevel)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "newLevel must be greater than current level")
	mockDB.AssertExpectations(t)
}

func TestUpdateUserProgress_UserNotFound(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockDatabase)
	userService := services.NewUserService(mockDB)

	ctx := context.Background()
	userId := "nonexistent"

	// Mock database calls:
	mockDB.On("GetUser", mock.Anything, userId).Return((*models.User)(nil), nil)

	// Act
	user, err := userService.UpdateUserProgress(ctx, userId, 10)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")
	mockDB.AssertExpectations(t)
}

func TestCreateUser_DBError(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockDatabase)
	userService := services.NewUserService(mockDB)

	ctx := context.Background()
	username := "dbErrorUser"
	country := "US"

	mockDB.On("PutUser", mock.Anything, mock.AnythingOfType("models.User")).Return(errors.New("db error"))

	// Act
	user, err := userService.CreateUser(ctx, username, country)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "could not create user")
	mockDB.AssertExpectations(t)
}
