// services/leaderboard_service.go
package services

import (
	"context"
	"fmt"
	"good_blast/database"
	"good_blast/errors"
	"good_blast/models"
)

// LeaderboardService implements LeaderboardServiceInterface
type LeaderboardService struct {
	DB database.DatabaseInterface
}

// NewLeaderboardService creates a new LeaderboardService
func NewLeaderboardService(db database.DatabaseInterface) *LeaderboardService {
	return &LeaderboardService{
		DB: db,
	}
}

// GetGlobalLeaderboard retrieves the top 1000 users globally based on level.
func (s *LeaderboardService) GetGlobalLeaderboard(ctx context.Context) ([]models.User, error) {
	users, err := s.DB.QueryGlobalLeaderboard(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global leaderboard: %w", err)
	}
	return users, nil
}

// GetCountryLeaderboard retrieves the top 1000 users in a specific country based on level.
func (s *LeaderboardService) GetCountryLeaderboard(ctx context.Context, countryCode string) ([]models.User, error) {
	users, err := s.DB.QueryUsersByCountryLevel(ctx, countryCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get country leaderboard: %w", err)
	}
	return users, nil
}

// GetTournamentLeaderboard retrieves the top 35 users in a specific tournament group based on score.
func (s *LeaderboardService) GetTournamentLeaderboard(ctx context.Context, groupId string) ([]models.TournamentEntry, error) {
	entries, err := s.DB.QueryTournamentEntriesByGroupScore(ctx, groupId)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament leaderboard: %w", err)
	}
	return entries, nil
}

// GetTournamentRank retrieves a user's rank in a specific tournament group.
func (s *LeaderboardService) GetTournamentRank(ctx context.Context, tournamentId string, userId string) (int, error) {
	// Fetch the user's tournament entry
	entry, err := s.DB.GetTournamentEntry(ctx, tournamentId, userId)
	if err != nil {
		return 0, fmt.Errorf("failed to get tournament entry: %w", err)
	}
	if entry == nil {
		return 0, errors.ErrTournamentEntryNotFound
	}

	// Fetch the group leaderboard
	groupEntries, err := s.DB.QueryTournamentEntriesByGroupScore(ctx, entry.GroupID)
	if err != nil {
		return 0, fmt.Errorf("failed to query group leaderboard: %w", err)
	}

	// Determine the rank
	for i, e := range groupEntries {
		if e.UserID == userId {
			return i + 1, nil // Rank is 1-based
		}
	}

	return 0, errors.ErrUserNotFoundInLeaderboard
}
