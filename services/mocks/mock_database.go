// services/mocks/mock_database.go
package mocks

import (
	"context"

	"good_blast/models"

	"github.com/stretchr/testify/mock"
)

// MockDatabase is a mock implementation of the DatabaseInterface.
type MockDatabase struct {
	mock.Mock
}

// PutUser mocks the PutUser method of DatabaseInterface.
func (m *MockDatabase) PutUser(ctx context.Context, user models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// GetUser mocks the GetUser method of DatabaseInterface.
func (m *MockDatabase) GetUser(ctx context.Context, userId string) (*models.User, error) {
	args := m.Called(ctx, userId)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateUserCoinsAndLevel mocks the UpdateUserCoinsAndLevel method of DatabaseInterface.
func (m *MockDatabase) UpdateUserCoinsAndLevel(ctx context.Context, userId string, newLevel, newCoins int) error {
	args := m.Called(ctx, userId, newLevel, newCoins)
	return args.Error(0)
}

// PutTournament mocks the PutTournament method of DatabaseInterface.
func (m *MockDatabase) PutTournament(ctx context.Context, tournament models.Tournament) error {
	args := m.Called(ctx, tournament)
	return args.Error(0)
}

// GetTournament mocks the GetTournament method of DatabaseInterface.
func (m *MockDatabase) GetTournament(ctx context.Context, tournamentId string) (*models.Tournament, error) {
	args := m.Called(ctx, tournamentId)
	if tournament, ok := args.Get(0).(*models.Tournament); ok {
		return tournament, args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateTournamentStatus mocks the UpdateTournamentStatus method of DatabaseInterface.
func (m *MockDatabase) UpdateTournamentStatus(ctx context.Context, tournamentId string, active bool) error {
	args := m.Called(ctx, tournamentId, active)
	return args.Error(0)
}

// PutTournamentEntry mocks the PutTournamentEntry method of DatabaseInterface.
func (m *MockDatabase) PutTournamentEntry(ctx context.Context, entry models.TournamentEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

// GetTournamentEntry mocks the GetTournamentEntry method of DatabaseInterface.
func (m *MockDatabase) GetTournamentEntry(ctx context.Context, tournamentId, userId string) (*models.TournamentEntry, error) {
	args := m.Called(ctx, tournamentId, userId)
	if entry, ok := args.Get(0).(*models.TournamentEntry); ok {
		return entry, args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateTournamentScore mocks the UpdateTournamentScore method of DatabaseInterface.
func (m *MockDatabase) UpdateTournamentScore(ctx context.Context, tournamentId, userId string, increment int) error {
	args := m.Called(ctx, tournamentId, userId, increment)
	return args.Error(0)
}

// QueryGlobalLeaderboard mocks the QueryGlobalLeaderboard method of DatabaseInterface.
func (m *MockDatabase) QueryGlobalLeaderboard(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	if users, ok := args.Get(0).([]models.User); ok {
		return users, args.Error(1)
	}
	return nil, args.Error(1)
}

// QueryUsersByCountryLevel mocks the QueryUsersByCountryLevel method of DatabaseInterface.
func (m *MockDatabase) QueryUsersByCountryLevel(ctx context.Context, country string) ([]models.User, error) {
	args := m.Called(ctx, country)
	if users, ok := args.Get(0).([]models.User); ok {
		return users, args.Error(1)
	}
	return nil, args.Error(1)
}

// QueryTournamentEntriesByGroupScore mocks the QueryTournamentEntriesByGroupScore method of DatabaseInterface.
func (m *MockDatabase) QueryTournamentEntriesByGroupScore(ctx context.Context, groupId string) ([]models.TournamentEntry, error) {
	args := m.Called(ctx, groupId)
	if entries, ok := args.Get(0).([]models.TournamentEntry); ok {
		return entries, args.Error(1)
	}
	return nil, args.Error(1)
}

// EnterTournamentTransaction mocks the EnterTournamentTransaction method of DatabaseInterface.
func (m *MockDatabase) EnterTournamentTransaction(ctx context.Context, userID string, level, coins int, t *models.Tournament) error {
	args := m.Called(ctx, userID, level, coins, t)
	return args.Error(0)
}

// ClaimRewardTransaction mocks the ClaimRewardTransaction method of DatabaseInterface.
func (m *MockDatabase) ClaimRewardTransaction(ctx context.Context, userID string, reward int, tournamentID string) error {
	args := m.Called(ctx, userID, reward, tournamentID)
	return args.Error(0)
}

// QueryTournamentEntries mocks the QueryTournamentEntries method of DatabaseInterface.
func (m *MockDatabase) QueryTournamentEntries(ctx context.Context, tournamentId string) ([]models.TournamentEntry, error) {
	args := m.Called(ctx, tournamentId)
	if entries, ok := args.Get(0).([]models.TournamentEntry); ok {
		return entries, args.Error(1)
	}
	return nil, args.Error(1)
}
