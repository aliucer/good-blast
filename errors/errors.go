// errors/errors.go
package errors

import "errors"

// Define custom error variables
var (
	ErrAlreadyInTournament        = errors.New("you are already in the tournament")
	ErrGroupFull                  = errors.New("the tournament group is full")
	ErrRequirementsNotMet         = errors.New("you do not meet the requirements to claim the reward")
	ErrTournamentNotActive        = errors.New("the tournament is not active")
	ErrInsufficientCoins          = errors.New("insufficient coins to enter the tournament")
	ErrUserNotFound               = errors.New("user not found")
	ErrUserLevelTooLow            = errors.New("user level is too low")
	ErrTournamentNotFound         = errors.New("tournament not found")
	ErrTournamentAlreadyInactive  = errors.New("tournament is already inactive")
	ErrTournamentEntryNotFound    = errors.New("tournament entry not found")
	ErrRewardAlreadyClaimed       = errors.New("reward has already been claimed")
	ErrGroupIDMissing             = errors.New("user's groupId is missing")
	ErrNoRewardForRank            = errors.New("no reward available for your rank in the group")
	ErrTournamentStillActive      = errors.New("tournament is still active")
	ErrUserNotFoundInLeaderboard  = errors.New("user not found in the leaderboard")
	ErrInvalidLevelIncrease       = errors.New("newLevel must be greater than current level")
	ErrRequirementsNotMetForEntry = errors.New("you do not meet the requirements to enter the tournament")
)
