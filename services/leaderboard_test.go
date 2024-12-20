package services_test

import (
	"context"
	"errors"
	"testing"

	"good_blast/models"
	"good_blast/services"
	"good_blast/services/mocks"
	redisclient "good_blast/services/redis_client"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	redisclient.RDB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func TestGetGlobalLeaderboard_DBError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewLeaderboardService(mockDB)
	ctx := context.Background()

	// Clear cache
	redisclient.RDB.Del(ctx, "leaderboard:global")

	mockDB.On("QueryGlobalLeaderboard", mock.Anything).Return(nil, errors.New("db error"))

	result, err := service.GetGlobalLeaderboard(ctx)
	assert.Error(t, err)
	assert.Nil(t, result)
	mockDB.AssertExpectations(t)
}

func TestGetCountryLeaderboard_DBError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewLeaderboardService(mockDB)
	ctx := context.Background()
	countryCode := "DE"

	redisclient.RDB.Del(ctx, "leaderboard:country:"+countryCode)

	mockDB.On("QueryUsersByCountryLevel", mock.Anything, countryCode).Return(nil, errors.New("db error"))

	result, err := service.GetCountryLeaderboard(ctx, countryCode)
	assert.Error(t, err)
	assert.Nil(t, result)
	mockDB.AssertExpectations(t)
}

func TestGetTournamentLeaderboard_DBError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewLeaderboardService(mockDB)
	ctx := context.Background()
	groupId := "g-err"

	redisclient.RDB.Del(ctx, "leaderboard:tournament:"+groupId)

	mockDB.On("QueryTournamentEntriesByGroupScore", mock.Anything, groupId).Return(nil, errors.New("db error"))

	result, err := service.GetTournamentLeaderboard(ctx, groupId)
	assert.Error(t, err)
	assert.Nil(t, result)
	mockDB.AssertExpectations(t)
}

func TestGetTournamentRank_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewLeaderboardService(mockDB)
	ctx := context.Background()
	tID := "t-2024"
	userID := "user123"
	groupId := "g-xyz"
	entry := &models.TournamentEntry{
		TournamentID: tID,
		UserID:       userID,
		Score:        1000,
		GroupID:      groupId,
	}

	topEntries := []models.TournamentEntry{
		{TournamentID: tID, UserID: "user999", Score: 1200, GroupID: groupId},
		{TournamentID: tID, UserID: "user123", Score: 1000, GroupID: groupId},
		{TournamentID: tID, UserID: "user555", Score: 900, GroupID: groupId},
	}

	mockDB.On("GetTournamentEntry", mock.Anything, tID, userID).Return(entry, nil)
	mockDB.On("QueryTournamentEntriesByGroupScore", mock.Anything, groupId).Return(topEntries, nil)

	rank, err := service.GetTournamentRank(ctx, tID, userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, rank)
	mockDB.AssertExpectations(t)
}

func TestGetTournamentRank_EntryNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewLeaderboardService(mockDB)
	ctx := context.Background()
	tID := "t-2024"
	userID := "unknown-user"

	mockDB.On("GetTournamentEntry", mock.Anything, tID, userID).Return((*models.TournamentEntry)(nil), nil)

	_, err := service.GetTournamentRank(ctx, tID, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tournament entry not found")
	mockDB.AssertExpectations(t)
}

func TestGetTournamentRank_DBErrorOnEntry(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewLeaderboardService(mockDB)
	ctx := context.Background()
	tID := "t-2024"
	userID := "user123"

	mockDB.On("GetTournamentEntry", mock.Anything, tID, userID).Return(nil, errors.New("db error"))

	_, err := service.GetTournamentRank(ctx, tID, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get tournament entry")
	mockDB.AssertExpectations(t)
}

func TestGetTournamentRank_DBErrorOnQuery(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewLeaderboardService(mockDB)
	ctx := context.Background()
	tID := "t-2024"
	userID := "user123"
	entry := &models.TournamentEntry{
		TournamentID: tID,
		UserID:       userID,
		Score:        500,
		GroupID:      "g-1",
	}

	mockDB.On("GetTournamentEntry", mock.Anything, tID, userID).Return(entry, nil)
	mockDB.On("QueryTournamentEntriesByGroupScore", mock.Anything, "g-1").Return(nil, errors.New("db error"))

	_, err := service.GetTournamentRank(ctx, tID, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to query group leaderboard")
	mockDB.AssertExpectations(t)
}

func TestGetTournamentRank_UserNotInList(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	service := services.NewLeaderboardService(mockDB)
	ctx := context.Background()
	tID := "t-2024"
	userID := "user123"
	entry := &models.TournamentEntry{
		TournamentID: tID,
		UserID:       userID,
		Score:        500,
		GroupID:      "g-1",
	}

	topEntries := []models.TournamentEntry{
		{TournamentID: tID, UserID: "someoneelse", Score: 700, GroupID: "g-1"},
		{TournamentID: tID, UserID: "anotherone", Score: 600, GroupID: "g-1"},
	}

	mockDB.On("GetTournamentEntry", mock.Anything, tID, userID).Return(entry, nil)
	mockDB.On("QueryTournamentEntriesByGroupScore", mock.Anything, "g-1").Return(topEntries, nil)

	_, err := service.GetTournamentRank(ctx, tID, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found in the leaderboard")
	mockDB.AssertExpectations(t)
}
