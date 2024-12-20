package services_test

import (
	"context"
	"good_blast/errors"
	"testing"
	"time"

	"good_blast/models"
	"good_blast/services"
	"good_blast/services/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStartTournament_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	today := time.Now().UTC().Format("2006-01-02")

	// Mock that there's no existing active tournament
	mockDB.On("GetTournament", mock.Anything, today).Return((*models.Tournament)(nil), nil)
	// Expect a PutTournament call
	mockDB.On("PutTournament", mock.Anything, mock.AnythingOfType("models.Tournament")).
		Return(nil)

	tournament, err := service.StartTournament(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, tournament)
	assert.Equal(t, today, tournament.TournamentID)
	assert.True(t, tournament.Active)

	mockDB.AssertExpectations(t)
}

func TestStartTournament_AlreadyActive(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	today := time.Now().UTC().Format("2006-01-02")
	activeTournament := &models.Tournament{
		TournamentID:      today,
		StartTime:         "someStartTime",
		EndTime:           "someEndTime",
		Active:            true,
		CurrentGroupIndex: 1,
		CurrentGroupCount: 0,
	}

	mockDB.On("GetTournament", mock.Anything, today).Return(activeTournament, nil)

	tournament, err := service.StartTournament(ctx)
	assert.Nil(t, tournament)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrAlreadyInTournament, err)

	mockDB.AssertExpectations(t)
}

func TestEndTournament_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	tID := "2024-01-02"
	tournament := &models.Tournament{
		TournamentID:      tID,
		Active:            true,
		CurrentGroupIndex: 1,
		CurrentGroupCount: 0,
	}

	// Mock retrieval and update
	mockDB.On("GetTournament", mock.Anything, tID).Return(tournament, nil)
	mockDB.On("UpdateTournamentStatus", mock.Anything, tID, false).Return(nil)

	err := service.EndTournament(ctx, tID)
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestEndTournament_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	tID := "nonexistent"

	mockDB.On("GetTournament", mock.Anything, tID).Return((*models.Tournament)(nil), nil)

	err := service.EndTournament(ctx, tID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrTournamentNotFound, err)

	mockDB.AssertExpectations(t)
}

func TestEndTournament_AlreadyInactive(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	tID := "2024-01-02"
	inactiveTournament := &models.Tournament{
		TournamentID: tID,
		Active:       false,
	}

	mockDB.On("GetTournament", mock.Anything, tID).Return(inactiveTournament, nil)

	err := service.EndTournament(ctx, tID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrTournamentAlreadyInactive, err)

	mockDB.AssertExpectations(t)
}

func TestEnterTournament_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	userID := "user123"
	tID := "2024-01-02"
	tournament := &models.Tournament{
		TournamentID:      tID,
		Active:            true,
		CurrentGroupIndex: 1,
		CurrentGroupCount: 0,
	}
	user := &models.User{
		UserID: "user123",
		Level:  15,
		Coins:  1000,
	}

	// Mock DB calls
	mockDB.On("GetTournament", mock.Anything, tID).Return(tournament, nil).Once()
	mockDB.On("GetUser", mock.Anything, userID).Return(user, nil).Once()
	// EnterTournamentTransaction should succeed
	mockDB.On("EnterTournamentTransaction", mock.Anything, userID, user.Level, user.Coins, tournament).Return(nil).Once()

	remainingCoins, err := service.EnterTournament(ctx, userID, tID)
	assert.NoError(t, err)
	assert.Equal(t, user.Coins-500, remainingCoins)
	mockDB.AssertExpectations(t)
}

func TestEnterTournament_NotActive(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	userID := "user123"
	tID := "inactive-tournament"
	inactiveTournament := &models.Tournament{
		TournamentID: tID,
		Active:       false,
	}

	mockDB.On("GetTournament", mock.Anything, tID).Return(inactiveTournament, nil)

	_, err := service.EnterTournament(ctx, userID, tID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrTournamentNotActive, err)
	mockDB.AssertExpectations(t)
}

func TestEnterTournament_UserNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	userID := "unknown-user"
	tID := "2024-01-02"
	tournament := &models.Tournament{
		TournamentID:      tID,
		Active:            true,
		CurrentGroupIndex: 1,
		CurrentGroupCount: 0,
	}

	mockDB.On("GetTournament", mock.Anything, tID).Return(tournament, nil)
	mockDB.On("GetUser", mock.Anything, userID).Return((*models.User)(nil), nil)

	_, err := service.EnterTournament(ctx, userID, tID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrUserNotFound, err)
	mockDB.AssertExpectations(t)
}

func TestEnterTournament_LevelTooLow(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	userID := "lowlevel-user"
	tID := "2024-01-02"
	tournament := &models.Tournament{
		TournamentID:      tID,
		Active:            true,
		CurrentGroupIndex: 1,
		CurrentGroupCount: 0,
	}
	user := &models.User{
		UserID: "lowlevel-user",
		Level:  5,
		Coins:  1000,
	}

	mockDB.On("GetTournament", mock.Anything, tID).Return(tournament, nil)
	mockDB.On("GetUser", mock.Anything, userID).Return(user, nil)

	_, err := service.EnterTournament(ctx, userID, tID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrUserLevelTooLow, err)
	mockDB.AssertExpectations(t)
}

func TestEnterTournament_InsufficientCoins(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	userID := "poor-user"
	tID := "2024-01-02"
	tournament := &models.Tournament{
		TournamentID:      tID,
		Active:            true,
		CurrentGroupIndex: 1,
		CurrentGroupCount: 0,
	}
	user := &models.User{
		UserID: "poor-user",
		Level:  20,
		Coins:  300, // Not enough for 500 entry fee
	}

	mockDB.On("GetTournament", mock.Anything, tID).Return(tournament, nil)
	mockDB.On("GetUser", mock.Anything, userID).Return(user, nil)

	_, err := service.EnterTournament(ctx, userID, tID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInsufficientCoins, err)
	mockDB.AssertExpectations(t)
}

func TestUpdateScore_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	tID := "2024-01-02"
	userID := "user123"
	entry := &models.TournamentEntry{
		TournamentID: tID,
		UserID:       userID,
		Score:        100,
		GroupID:      "group-1",
	}

	mockDB.On("GetTournamentEntry", mock.Anything, tID, userID).Return(entry, nil).Once()
	mockDB.On("UpdateTournamentScore", mock.Anything, tID, userID, 50).Return(nil).Once()

	newScore, err := service.UpdateScore(ctx, tID, userID, 50)
	assert.NoError(t, err)
	assert.Equal(t, entry.Score+50, newScore)
	mockDB.AssertExpectations(t)
}

func TestUpdateScore_EntryNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	tID := "2024-01-02"
	userID := "unknown-user"

	mockDB.On("GetTournamentEntry", mock.Anything, tID, userID).Return((*models.TournamentEntry)(nil), nil)

	_, err := service.UpdateScore(ctx, tID, userID, 10)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrTournamentEntryNotFound, err)
	mockDB.AssertExpectations(t)
}

func TestClaimReward_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	tID := "2024-01-02"
	userID := "user123"
	tournament := &models.Tournament{
		TournamentID:      tID,
		Active:            false,
		CurrentGroupIndex: 1,
		CurrentGroupCount: 0,
	}
	entry := &models.TournamentEntry{
		TournamentID:  tID,
		UserID:        userID,
		Score:         2000,
		GroupID:       "g-1",
		ClaimedReward: false,
	}

	topEntries := []models.TournamentEntry{
		{
			TournamentID: tID, UserID: userID,
			Score: 2000, GroupID: "g-1", ClaimedReward: false,
		},
	}

	mockDB.On("GetTournament", mock.Anything, tID).Return(tournament, nil)
	mockDB.On("GetTournamentEntry", mock.Anything, tID, userID).Return(entry, nil)
	mockDB.On("QueryTournamentEntriesByGroupScore", mock.Anything, "g-1").Return(topEntries, nil)
	mockDB.On("ClaimRewardTransaction", mock.Anything, userID, 5000, tID).Return(nil)

	rank, reward, err := service.ClaimReward(ctx, tID, userID)
	assert.NoError(t, err)
	assert.Equal(t, 1, rank)
	assert.Equal(t, 5000, reward)
	mockDB.AssertExpectations(t)
}

func TestClaimReward_TournamentNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	tID := "nonexistent"
	userID := "user123"

	mockDB.On("GetTournament", mock.Anything, tID).Return((*models.Tournament)(nil), nil)

	_, _, err := service.ClaimReward(ctx, tID, userID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrTournamentNotFound, err)
	mockDB.AssertExpectations(t)
}

func TestClaimReward_StillActive(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	tID := "active-tid"
	userID := "user123"
	activeTournament := &models.Tournament{
		TournamentID: tID,
		Active:       true,
	}

	mockDB.On("GetTournament", mock.Anything, tID).Return(activeTournament, nil)

	_, _, err := service.ClaimReward(ctx, tID, userID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrTournamentStillActive, err)
	mockDB.AssertExpectations(t)
}

func TestClaimReward_EntryNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	tID := "2024-01-02"
	userID := "unknown-user"
	inactiveTournament := &models.Tournament{
		TournamentID: tID,
		Active:       false,
	}

	mockDB.On("GetTournament", mock.Anything, tID).Return(inactiveTournament, nil)
	mockDB.On("GetTournamentEntry", mock.Anything, tID, userID).Return((*models.TournamentEntry)(nil), nil)

	_, _, err := service.ClaimReward(ctx, tID, userID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrTournamentEntryNotFound, err)
	mockDB.AssertExpectations(t)
}

func TestClaimReward_AlreadyClaimed(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	tID := "2024-01-02"
	userID := "user123"
	tournament := &models.Tournament{
		TournamentID: tID,
		Active:       false,
	}
	entry := &models.TournamentEntry{
		TournamentID:  tID,
		UserID:        userID,
		Score:         1000,
		GroupID:       "g-1",
		ClaimedReward: true,
	}

	mockDB.On("GetTournament", mock.Anything, tID).Return(tournament, nil)
	mockDB.On("GetTournamentEntry", mock.Anything, tID, userID).Return(entry, nil)

	_, _, err := service.ClaimReward(ctx, tID, userID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrRewardAlreadyClaimed, err)
	mockDB.AssertExpectations(t)
}

func TestClaimReward_NoRewardForRank(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewTournamentService(mockDB)

	ctx := context.Background()
	tID := "2024-01-02"
	userID := "user123"
	tournament := &models.Tournament{
		TournamentID: tID,
		Active:       false,
	}
	// Entry exists but user won't be in top 10
	entry := &models.TournamentEntry{
		TournamentID:  tID,
		UserID:        userID,
		Score:         500,
		GroupID:       "g-1",
		ClaimedReward: false,
	}
	// Simulate topEntries that do not include user at all
	topEntries := []models.TournamentEntry{
		{TournamentID: tID, UserID: "other1", Score: 1000, GroupID: "g-1"},
		{TournamentID: tID, UserID: "other2", Score: 900, GroupID: "g-1"},
	}

	mockDB.On("GetTournament", mock.Anything, tID).Return(tournament, nil)
	mockDB.On("GetTournamentEntry", mock.Anything, tID, userID).Return(entry, nil)
	mockDB.On("QueryTournamentEntriesByGroupScore", mock.Anything, "g-1").Return(topEntries, nil)

	_, _, err := service.ClaimReward(ctx, tID, userID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNoRewardForRank, err)
	mockDB.AssertExpectations(t)
}
