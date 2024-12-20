// services/interfaces.go
package services

import (
	"context"
	"good_blast/models"
)

// LeaderboardServiceInterface defines all the methods related to leaderboard operations.
type LeaderboardServiceInterface interface {
	GetGlobalLeaderboard(ctx context.Context) ([]models.User, error)
	GetCountryLeaderboard(ctx context.Context, countryCode string) ([]models.User, error)
	GetTournamentLeaderboard(ctx context.Context, groupId string) ([]models.TournamentEntry, error)
	GetTournamentRank(ctx context.Context, tournamentId string, userId string) (int, error)
}

// TournamentServiceInterface defines all the methods related to tournament operations.
type TournamentServiceInterface interface {
	StartTournament(ctx context.Context) (*models.Tournament, error)
	EndTournament(ctx context.Context, tournamentID string) error
	EnterTournament(ctx context.Context, userID string, tournamentID string) (int, error)
	UpdateScore(ctx context.Context, tournamentID string, userID string, increment int) (int, error)
	ClaimReward(ctx context.Context, tournamentID string, userID string) (int, int, error)
}

// UserServiceInterface defines all the methods related to user operations.
type UserServiceInterface interface {
	CreateUser(ctx context.Context, username, country string) (*models.User, error)
	GetUser(ctx context.Context, userID string) (*models.User, error)
	UpdateUserProgress(ctx context.Context, userID string, newLevel int) (*models.User, error)
}
