package models

// User represents a player in the Good Blast game.
type User struct {
	UserID   string `json:"userId" dynamodbav:"userId"`                       // Partition Key in DynamoDB
	Username string `json:"username" dynamodbav:"username"`                   // Unique username
	Level    int    `json:"level" dynamodbav:"level"`                         // User's current level
	Coins    int    `json:"coins" dynamodbav:"coins"`                         // User's coin balance
	Country  string `json:"country,omitempty" dynamodbav:"country,omitempty"` // Optional ISO country code
	GlobalPK string `json:"globalPK" dynamodbav:"globalPK"`                   // Global Leaderboard Partition Key
}
