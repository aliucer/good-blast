// services/leaderboard_service.go
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"good_blast/database"
	"good_blast/errors"
	"good_blast/models" // to access services.RDB
	redisclient "good_blast/services/redis_client"
	"time"
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

func (s *LeaderboardService) GetGlobalLeaderboard(ctx context.Context) ([]models.User, error) {
	// 1. Attempt to get from Redis
	cachedData, err := redisclient.RDB.Get(ctx, "leaderboard:global").Result()
	if err == nil && cachedData != "" {
		// We got cached data
		var users []models.User
		if err := json.Unmarshal([]byte(cachedData), &users); err == nil {
			return users, nil
		}
		// If unmarshal fails, we fall through and fetch fresh data
	}

	// 2. If not in Redis or error, fetch from DynamoDB
	users, err := s.DB.QueryGlobalLeaderboard(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global leaderboard: %w", err)
	}

	// 3. Cache the result in Redis for, say, 60 seconds
	userBytes, err := json.Marshal(users)
	if err == nil {
		redisclient.RDB.Set(ctx, "leaderboard:global", string(userBytes), 60*time.Second)
	}

	return users, nil
}

func (s *LeaderboardService) GetCountryLeaderboard(ctx context.Context, countryCode string) ([]models.User, error) {
	cacheKey := "leaderboard:country:" + countryCode

	// Try Redis first
	cachedData, err := redisclient.RDB.Get(ctx, "leaderboard:global").Result()
	if err == nil && cachedData != "" {
		var users []models.User
		if err := json.Unmarshal([]byte(cachedData), &users); err == nil {
			return users, nil
		}
	}

	// Fetch from DB if Redis miss
	users, err := s.DB.QueryUsersByCountryLevel(ctx, countryCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get country leaderboard: %w", err)
	}

	// Cache result
	userBytes, err := json.Marshal(users)
	if err == nil {
		redisclient.RDB.Set(ctx, cacheKey, string(userBytes), 60*time.Second)
	}

	return users, nil
}

// GetTournamentLeaderboard retrieves the top 35 users in a specific tournament group based on score.
func (s *LeaderboardService) GetTournamentLeaderboard(ctx context.Context, groupId string) ([]models.TournamentEntry, error) {
	cacheKey := "leaderboard:tournament:" + groupId

	cachedData, err := redisclient.RDB.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		var entries []models.TournamentEntry
		if err := json.Unmarshal([]byte(cachedData), &entries); err == nil {
			return entries, nil
		}
	}

	entries, err := s.DB.QueryTournamentEntriesByGroupScore(ctx, groupId)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament leaderboard: %w", err)
	}

	entryBytes, err := json.Marshal(entries)
	if err == nil {
		redisclient.RDB.Set(ctx, cacheKey, string(entryBytes), 60*time.Second)
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
