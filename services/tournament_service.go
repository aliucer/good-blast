// services/tournament_service.go
package services

import (
	"context"
	"log"
	"time"

	"good_blast/database"
	"good_blast/errors"
	"good_blast/models"
)

// TournamentService implements TournamentServiceInterface.
type TournamentService struct {
	DB database.DatabaseInterface
}

// NewTournamentService creates a new instance of TournamentService.
func NewTournamentService(db database.DatabaseInterface) *TournamentService {
	return &TournamentService{
		DB: db,
	}
}

// StartTournament initializes a new tournament.
func (s *TournamentService) StartTournament(ctx context.Context) (*models.Tournament, error) {
	nowUTC := time.Now().UTC()
	tournamentID := nowUTC.Format("2006-01-02") // e.g., "2024-01-15"

	// Check if tournament already exists for today
	existingTournament, err := s.DB.GetTournament(ctx, tournamentID)
	if err != nil {
		log.Println("Error checking existing tournament:", err)
		return nil, err
	}
	if existingTournament != nil && existingTournament.Active {
		return nil, errors.ErrAlreadyInTournament
	}

	startTime := nowUTC.Format(time.RFC3339)                   // e.g., "2024-01-15T00:00:00Z"
	endTime := nowUTC.Add(24 * time.Hour).Format(time.RFC3339) // e.g., "2024-01-16T00:00:00Z"
	tournament := models.Tournament{
		TournamentID:      tournamentID,
		StartTime:         startTime,
		EndTime:           endTime,
		Active:            true,
		CurrentGroupIndex: 1, // Initialize group index
		CurrentGroupCount: 0, // Initialize group count
	}

	// Insert into Tournaments table
	if err := s.DB.PutTournament(ctx, tournament); err != nil {
		log.Println("Error creating tournament:", err)
		return nil, err
	}

	return &tournament, nil
}

// EndTournament marks a tournament as inactive.
func (s *TournamentService) EndTournament(ctx context.Context, tournamentID string) error {
	t, err := s.DB.GetTournament(ctx, tournamentID)
	if err != nil {
		log.Println("Error fetching tournament:", err)
		return err
	}
	if t == nil {
		return errors.ErrTournamentNotFound
	}
	if !t.Active {
		return errors.ErrTournamentAlreadyInactive
	}

	// Mark the tournament as inactive
	if err := s.DB.UpdateTournamentStatus(ctx, tournamentID, false); err != nil {
		log.Println("Error ending tournament:", err)
		return err
	}

	return nil
}

// EnterTournament allows a user to enter an active tournament.
func (s *TournamentService) EnterTournament(ctx context.Context, userID string, tournamentID string) (int, error) {
	// Fetch the tournament
	t, err := s.DB.GetTournament(ctx, tournamentID)
	if err != nil {
		log.Println("Error fetching tournament:", err)
		return 0, err
	}
	if t == nil || !t.Active {
		return 0, errors.ErrTournamentNotActive
	}

	// Fetch the user
	user, err := s.DB.GetUser(ctx, userID)
	if err != nil {
		log.Println("Error fetching user:", err)
		return 0, err
	}
	if user == nil {
		return 0, errors.ErrUserNotFound
	}
	if user.Level < 10 {
		return 0, errors.ErrUserLevelTooLow
	}
	if user.Coins < 500 {
		return 0, errors.ErrInsufficientCoins
	}

	// Perform the tournament entry transaction
	err = s.DB.EnterTournamentTransaction(ctx, userID, user.Level, user.Coins, t)
	if err != nil {
		log.Println("EnterTournamentTransaction error:", err)
		return 0, err
	}

	remainingCoins := user.Coins - 500
	return remainingCoins, nil
}

// UpdateScore increments a user's score during the active tournament.
func (s *TournamentService) UpdateScore(ctx context.Context, tournamentID string, userID string, increment int) (int, error) {
	// Fetch the tournament entry
	entry, err := s.DB.GetTournamentEntry(ctx, tournamentID, userID)
	if err != nil {
		log.Println("Error fetching tournament entry:", err)
		return 0, err
	}
	if entry == nil {
		return 0, errors.ErrTournamentEntryNotFound
	}

	// Update the score
	if err := s.DB.UpdateTournamentScore(ctx, tournamentID, userID, increment); err != nil {
		log.Println("Error updating tournament score:", err)
		return 0, err
	}

	newScore := entry.Score + increment
	return newScore, nil
}

// ClaimReward allows a user to claim their reward after the tournament has ended.
func (s *TournamentService) ClaimReward(ctx context.Context, tournamentID string, userID string) (int, int, error) {
	// Fetch the tournament
	t, err := s.DB.GetTournament(ctx, tournamentID)
	if err != nil || t == nil {
		log.Println("Error or no tournament found:", err)
		return 0, 0, errors.ErrTournamentNotFound
	}

	// Ensure the tournament has ended
	if t.Active {
		return 0, 0, errors.ErrTournamentStillActive
	}

	// Fetch the user's tournament entry
	entry, err := s.DB.GetTournamentEntry(ctx, tournamentID, userID)
	if err != nil {
		log.Println("Error fetching tournament entry:", err)
		return 0, 0, err
	}
	if entry == nil {
		return 0, 0, errors.ErrTournamentEntryNotFound
	}

	// Check if reward has already been claimed
	if entry.ClaimedReward {
		return 0, 0, errors.ErrRewardAlreadyClaimed
	}

	// Retrieve the groupId from the user's tournament entry
	groupID := entry.GroupID
	if groupID == "" {
		return 0, 0, errors.ErrGroupIDMissing
	}

	// Query top users within the user's group using the GroupScoreIndex
	topEntries, err := s.DB.QueryTournamentEntriesByGroupScore(ctx, groupID) // Fetch top 10 for rewards
	if err != nil {
		log.Println("Error querying GroupScoreIndex:", err)
		return 0, 0, err
	}

	// Determine user's rank within the group
	var userRank int
	for i, e := range topEntries {
		if e.UserID == userID {
			userRank = i + 1 // 1-based index
			break
		}
	}

	// If user is not in the top 10 of their group, no reward is applicable
	if userRank == 0 {
		return 0, 0, errors.ErrNoRewardForRank
	}

	// Reward logic based on rank within the group
	var reward int
	switch {
	case userRank == 1:
		reward = 5000
	case userRank == 2:
		reward = 3000
	case userRank == 3:
		reward = 2000
	case userRank >= 4 && userRank <= 10:
		reward = 1000
	default:
		reward = 0
	}

	if reward == 0 {
		return userRank, 0, errors.ErrNoRewardForRank
	}

	// Perform a transaction to update user coins and mark reward as claimed
	err = s.DB.ClaimRewardTransaction(ctx, userID, reward, tournamentID)
	if err != nil {
		log.Println("Error during reward transaction:", err)
		return 0, 0, err
	}

	return userRank, reward, nil
}
