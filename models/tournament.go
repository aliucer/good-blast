package models

// Tournament represents a daily tournament.
type Tournament struct {
	TournamentID      string `json:"tournamentId" dynamodbav:"tournamentId"`
	StartTime         string `json:"startTime" dynamodbav:"startTime"`
	EndTime           string `json:"endTime" dynamodbav:"endTime"`
	Active            bool   `json:"active" dynamodbav:"active"`
	CurrentGroupIndex int    `json:"currentGroupIndex" dynamodbav:"currentGroupIndex"` // New field for group indexing
	CurrentGroupCount int    `json:"currentGroupCount" dynamodbav:"currentGroupCount"` // How many users joined current group
}
