// database/interfaces.go
package database

import (
	"context"
	"good_blast/models"
)

// DatabaseInterface defines all the methods the database layer should implement.
type DatabaseInterface interface {
	PutUser(ctx context.Context, user models.User) error
	GetUser(ctx context.Context, userId string) (*models.User, error)
	UpdateUserCoinsAndLevel(ctx context.Context, userId string, newLevel, newCoins int) error

	PutTournament(ctx context.Context, tournament models.Tournament) error
	GetTournament(ctx context.Context, tournamentId string) (*models.Tournament, error)
	UpdateTournamentStatus(ctx context.Context, tournamentId string, active bool) error

	PutTournamentEntry(ctx context.Context, entry models.TournamentEntry) error
	GetTournamentEntry(ctx context.Context, tournamentId, userId string) (*models.TournamentEntry, error)
	UpdateTournamentScore(ctx context.Context, tournamentId, userId string, increment int) error

	QueryGlobalLeaderboard(ctx context.Context) ([]models.User, error)
	QueryUsersByCountryLevel(ctx context.Context, country string) ([]models.User, error)
	QueryTournamentEntriesByGroupScore(ctx context.Context, groupId string) ([]models.TournamentEntry, error)

	EnterTournamentTransaction(ctx context.Context, userID string, level, coins int, t *models.Tournament) error
	ClaimRewardTransaction(ctx context.Context, userID string, reward int, tournamentID string) error

	// Add the following if needed
	QueryTournamentEntries(ctx context.Context, tournamentId string) ([]models.TournamentEntry, error)
}
