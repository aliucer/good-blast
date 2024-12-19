package models

// TournamentEntry links a user to a specific tournament.
type TournamentEntry struct {
	TournamentID  string `json:"tournamentId" dynamodbav:"tournamentId"`               // Partition Key
	UserID        string `json:"userId" dynamodbav:"userId"`                           // Sort Key
	Score         int    `json:"score" dynamodbav:"score"`                             // Incremented as the user progresses
	GroupID       string `json:"groupId" dynamodbav:"groupId"`                         // Group identifier for partitioning users (max 35)
	ClaimedReward bool   `json:"claimedReward,omitempty" dynamodbav:"claimedReward"`   // Indicates if reward has been claimed
	ClaimedAt     string `json:"claimedAt,omitempty" dynamodbav:"claimedAt,omitempty"` // Timestamp of when reward was claimed
}
