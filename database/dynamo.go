// database/dynamo.go
package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"good_blast/errors"
	"good_blast/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// DynamoDB is a wrapper struct to implement DatabaseInterface
type DynamoDB struct{}

// Singleton pattern to ensure a single DynamoDB client
var (
	svc      *dynamodb.DynamoDB
	svcOnce  sync.Once
	svcError error

	// Table names from environment variables
	usersTable             string
	tournamentsTable       string
	tournamentEntriesTable string
)

// InitDynamoDB initializes the DynamoDB client and reads table names from environment variables
func InitDynamoDB() error {
	svcOnce.Do(func() {
		region := os.Getenv("DYNAMODB_REGION")
		if region == "" {
			svcError = fmt.Errorf("DYNAMODB_REGION environment variable not set")
			return
		}

		// Initialize DynamoDB session
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(region),
		})
		if err != nil {
			svcError = fmt.Errorf("failed to create AWS session: %v", err)
			return
		}

		svc = dynamodb.New(sess)
		log.Println("DynamoDB initialized in region:", region)

		// Read table names from environment variables
		usersTable = os.Getenv("USERS_TABLE")
		tournamentsTable = os.Getenv("TOURNAMENTS_TABLE")
		tournamentEntriesTable = os.Getenv("TOURNAMENT_ENTRIES_TABLE")

		if usersTable == "" || tournamentsTable == "" || tournamentEntriesTable == "" {
			svcError = fmt.Errorf("one or more DynamoDB table environment variables are not set")
			return
		}
	})
	return svcError
}

// Ensure DynamoDB implements DatabaseInterface
var _ DatabaseInterface = (*DynamoDB)(nil)

// PutUser inserts a new user into the Users table
func (db *DynamoDB) PutUser(ctx context.Context, user models.User) error {
	if svc == nil {
		return fmt.Errorf("DynamoDB client not initialized")
	}

	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %v", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(usersTable),
		Item:      av,
	}

	_, err = svc.PutItemWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put user: %v", err)
	}

	return nil
}

// GetUser retrieves a user by userId from the Users table
func (db *DynamoDB) GetUser(ctx context.Context, userId string) (*models.User, error) {
	if svc == nil {
		return nil, fmt.Errorf("DynamoDB client not initialized")
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]*dynamodb.AttributeValue{
			"userId": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	if result.Item == nil {
		// User not found
		return nil, nil
	}

	var user models.User
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %v", err)
	}

	return &user, nil
}

// UpdateUserCoinsAndLevel updates the user's level and coin balance
func (db *DynamoDB) UpdateUserCoinsAndLevel(ctx context.Context, userId string, newLevel, newCoins int) error {
	if svc == nil {
		return fmt.Errorf("DynamoDB client not initialized")
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]*dynamodb.AttributeValue{
			"userId": {S: aws.String(userId)},
		},
		UpdateExpression: aws.String("SET #lvl = :lvlVal, #cns = :coinsVal"),
		ExpressionAttributeNames: map[string]*string{
			"#lvl": aws.String("level"),
			"#cns": aws.String("coins"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":lvlVal":   {N: aws.String(fmt.Sprintf("%d", newLevel))},
			":coinsVal": {N: aws.String(fmt.Sprintf("%d", newCoins))},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	_, err := svc.UpdateItemWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	return nil
}

// PutTournament inserts a new tournament into the Tournaments table
func (db *DynamoDB) PutTournament(ctx context.Context, tournament models.Tournament) error {
	if svc == nil {
		return fmt.Errorf("DynamoDB client not initialized")
	}

	av, err := dynamodbattribute.MarshalMap(tournament)
	if err != nil {
		return fmt.Errorf("failed to marshal tournament: %v", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tournamentsTable),
		Item:      av,
	}

	_, err = svc.PutItemWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put tournament: %v", err)
	}

	return nil
}

// GetTournament retrieves a tournament by tournamentId from the Tournaments table
func (db *DynamoDB) GetTournament(ctx context.Context, tournamentId string) (*models.Tournament, error) {
	if svc == nil {
		return nil, fmt.Errorf("DynamoDB client not initialized")
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(tournamentsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"tournamentId": {S: aws.String(tournamentId)},
		},
	}

	result, err := svc.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament: %v", err)
	}
	if result.Item == nil {
		return nil, nil
	}

	var t models.Tournament
	err = dynamodbattribute.UnmarshalMap(result.Item, &t)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tournament: %v", err)
	}
	return &t, nil
}

// UpdateTournamentStatus updates the 'active' status of a tournament
func (db *DynamoDB) UpdateTournamentStatus(ctx context.Context, tournamentId string, active bool) error {
	if svc == nil {
		return fmt.Errorf("DynamoDB client not initialized")
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tournamentsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"tournamentId": {S: aws.String(tournamentId)},
		},
		UpdateExpression: aws.String("SET #act = :actVal"),
		ExpressionAttributeNames: map[string]*string{
			"#act": aws.String("active"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":actVal": {BOOL: aws.Bool(active)},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	_, err := svc.UpdateItemWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update tournament status: %v", err)
	}
	return nil
}

// PutTournamentEntry inserts a new tournament entry into the TournamentEntries table
func (db *DynamoDB) PutTournamentEntry(ctx context.Context, entry models.TournamentEntry) error {
	if svc == nil {
		return fmt.Errorf("DynamoDB client not initialized")
	}

	av, err := dynamodbattribute.MarshalMap(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal tournament entry: %v", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tournamentEntriesTable),
		Item:      av,
	}

	_, err = svc.PutItemWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put tournament entry: %v", err)
	}
	return nil
}

// GetTournamentEntry retrieves a tournament entry by tournamentId and userId
func (db *DynamoDB) GetTournamentEntry(ctx context.Context, tournamentId, userId string) (*models.TournamentEntry, error) {
	if svc == nil {
		return nil, fmt.Errorf("DynamoDB client not initialized")
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(tournamentEntriesTable),
		Key: map[string]*dynamodb.AttributeValue{
			"tournamentId": {S: aws.String(tournamentId)},
			"userId":       {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament entry: %v", err)
	}
	if result.Item == nil {
		return nil, nil
	}

	var entry models.TournamentEntry
	err = dynamodbattribute.UnmarshalMap(result.Item, &entry)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tournament entry: %v", err)
	}

	return &entry, nil
}

// UpdateTournamentScore updates a user's score in a tournament entry
func (db *DynamoDB) UpdateTournamentScore(ctx context.Context, tournamentId, userId string, increment int) error {
	if svc == nil {
		return fmt.Errorf("DynamoDB client not initialized")
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tournamentEntriesTable),
		Key: map[string]*dynamodb.AttributeValue{
			"tournamentId": {S: aws.String(tournamentId)},
			"userId":       {S: aws.String(userId)},
		},
		UpdateExpression: aws.String("SET #scr = #scr + :inc"),
		ExpressionAttributeNames: map[string]*string{
			"#scr": aws.String("score"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":inc": {N: aws.String(fmt.Sprintf("%d", increment))},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	_, err := svc.UpdateItemWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update tournament score: %v", err)
	}
	return nil
}

// QueryTournamentEntries retrieves all entries for a specific tournament
func (db *DynamoDB) QueryTournamentEntries(ctx context.Context, tournamentId string) ([]models.TournamentEntry, error) {
	if svc == nil {
		return nil, fmt.Errorf("DynamoDB client not initialized")
	}

	input := &dynamodb.QueryInput{
		TableName:              aws.String(tournamentEntriesTable),
		KeyConditionExpression: aws.String("tournamentId = :tid"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":tid": {S: aws.String(tournamentId)},
		},
	}

	result, err := svc.QueryWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query tournament entries: %v", err)
	}

	entries := make([]models.TournamentEntry, 0, *result.Count)
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &entries)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tournament entries: %v", err)
	}
	return entries, nil
}

// QueryGlobalLeaderboard queries the GlobalLevelIndex to retrieve top 1000 users globally
func (db *DynamoDB) QueryGlobalLeaderboard(ctx context.Context) ([]models.User, error) {
	if svc == nil {
		return nil, fmt.Errorf("DynamoDB client not initialized")
	}

	input := &dynamodb.QueryInput{
		TableName:              aws.String(usersTable),
		IndexName:              aws.String("GlobalLevelIndex"),
		KeyConditionExpression: aws.String("globalPK = :g"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":g": {S: aws.String("GLOBAL")},
		},
		ScanIndexForward: aws.Bool(false), // descending by level
		Limit:            aws.Int64(1000), // Set limit to 1000
	}

	result, err := svc.QueryWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query global leaderboard: %v", err)
	}

	var users []models.User
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &users)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal users: %v", err)
	}
	return users, nil
}

// QueryUsersByCountryLevel queries the CountryLevelIndex to retrieve top 1000 users in a country
func (db *DynamoDB) QueryUsersByCountryLevel(ctx context.Context, country string) ([]models.User, error) {
	if svc == nil {
		return nil, fmt.Errorf("DynamoDB client not initialized")
	}

	input := &dynamodb.QueryInput{
		TableName:              aws.String(usersTable),
		IndexName:              aws.String("CountryLevelIndex"), // Your GSI name
		KeyConditionExpression: aws.String("country = :c"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":c": {S: aws.String(country)},
		},
		// false => descending order by the sort key (level)
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int64(1000), // Set limit to 1000
	}

	result, err := svc.QueryWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error querying CountryLevelIndex: %w", err)
	}

	var users []models.User
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &users)
	if err != nil {
		log.Println("Error unmarshaling country leaderboard:", err)
		return nil, fmt.Errorf("failed to unmarshal users: %v", err)
	}

	return users, nil
}

// QueryTournamentEntriesByGroupScore queries the GroupScoreIndex to retrieve top 35 users in a group
func (db *DynamoDB) QueryTournamentEntriesByGroupScore(ctx context.Context, groupId string) ([]models.TournamentEntry, error) {
	if svc == nil {
		return nil, fmt.Errorf("DynamoDB client not initialized")
	}

	input := &dynamodb.QueryInput{
		TableName:              aws.String(tournamentEntriesTable),
		IndexName:              aws.String("GroupScoreIndex"),
		KeyConditionExpression: aws.String("groupId = :gid"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gid": {S: aws.String(groupId)},
		},
		ScanIndexForward: aws.Bool(false), // descending by score
		Limit:            aws.Int64(35),   // Set limit to 35
	}

	result, err := svc.QueryWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query tournament entries by group score: %v", err)
	}

	var entries []models.TournamentEntry
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &entries)
	if err != nil {
		log.Println("Error unmarshaling group leaderboard:", err)
		return nil, fmt.Errorf("failed to unmarshal tournament entries: %v", err)
	}

	return entries, nil
}

// EnterTournamentTransaction handles the transaction logic to enter a tournament
func (db *DynamoDB) EnterTournamentTransaction(ctx context.Context, userID string, level, coins int, t *models.Tournament) error {
	if svc == nil {
		return fmt.Errorf("DynamoDB client not initialized")
	}

	// 1. Update User Row: Deduct 500 coins, ensure coins >= 500 and level >= 10.
	updateUser := &dynamodb.Update{
		TableName:                aws.String(usersTable),
		Key:                      map[string]*dynamodb.AttributeValue{"userId": {S: aws.String(userID)}},
		UpdateExpression:         aws.String("SET #c = #c - :cost"),
		ConditionExpression:      aws.String("#c >= :cost AND #lvl >= :minLvl"),
		ExpressionAttributeNames: map[string]*string{"#c": aws.String("coins"), "#lvl": aws.String("level")},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":cost":   {N: aws.String("500")},
			":minLvl": {N: aws.String("10")},
		},
	}

	// 2. Update Tournaments Row: Increment currentGroupCount. If it hits 35, increment currentGroupIndex & reset currentGroupCount=1.
	groupIndex := t.CurrentGroupIndex
	groupCount := t.CurrentGroupCount

	newGroupIndex := groupIndex
	newGroupCount := groupCount + 1
	if newGroupCount > 35 {
		newGroupIndex = groupIndex + 1
		newGroupCount = 1
	}

	updateTournament := &dynamodb.Update{
		TableName:                aws.String(tournamentsTable),
		Key:                      map[string]*dynamodb.AttributeValue{"tournamentId": {S: aws.String(t.TournamentID)}},
		UpdateExpression:         aws.String("SET #gi = :newIndex, #gc = :newCount"),
		ConditionExpression:      aws.String("#gi = :oldIndex AND #gc = :oldCount"),
		ExpressionAttributeNames: map[string]*string{"#gi": aws.String("currentGroupIndex"), "#gc": aws.String("currentGroupCount")},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":oldIndex": {N: aws.String(fmt.Sprintf("%d", groupIndex))},
			":oldCount": {N: aws.String(fmt.Sprintf("%d", groupCount))},
			":newIndex": {N: aws.String(fmt.Sprintf("%d", newGroupIndex))},
			":newCount": {N: aws.String(fmt.Sprintf("%d", newGroupCount))},
		},
		ReturnValuesOnConditionCheckFailure: aws.String("NONE"),
	}

	// 3. Put the new entry in TournamentEntries with a unique groupID.
	groupID := fmt.Sprintf("%s-group-%d", t.TournamentID, newGroupIndex)

	entry := models.TournamentEntry{
		TournamentID:  t.TournamentID,
		UserID:        userID,
		Score:         0,
		GroupID:       groupID,
		ClaimedReward: false,
	}

	entryMap, err := dynamodbattribute.MarshalMap(entry)
	if err != nil {
		if tcErr, ok := err.(*dynamodb.TransactionCanceledException); ok {
			for _, r := range tcErr.CancellationReasons {
				if aws.StringValue(r.Code) == "ConditionalCheckFailed" {
					return errors.ErrAlreadyInTournament
				}
				// Handle other cancellation reasons if necessary
			}
			return fmt.Errorf("transaction canceled for unknown reasons")
		} else if aerr, ok := err.(awserr.Error); ok {
			log.Println("DynamoDB error:", aerr.Error())
			return fmt.Errorf("database error")
		} else {
			log.Println("Unknown error:", err.Error())
			return fmt.Errorf("unknown error")
		}
	}

	putEntry := &dynamodb.Put{
		TableName: aws.String(tournamentEntriesTable),
		Item:      entryMap,
	}

	// Build the transaction input.
	inputTxn := &dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			{Update: updateUser},
			{Update: updateTournament},
			{Put: putEntry},
		},
	}

	// Execute the transaction.
	_, err = svc.TransactWriteItemsWithContext(ctx, inputTxn)
	if err != nil {
		// Handle specific DynamoDB errors.
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeTransactionCanceledException:
				// Log the full error message for debugging
				log.Printf("Transaction canceled: %v", aerr.Message())
				// TODO: Implement more specific error handling if possible
				return errors.ErrAlreadyInTournament
			case dynamodb.ErrCodeConditionalCheckFailedException:
				log.Println("Conditional check failed:", aerr.Message())
				return errors.ErrRequirementsNotMet
			default:
				log.Println("DynamoDB error:", aerr.Error())
				return fmt.Errorf("database error")
			}
		} else {
			log.Println("Unknown error:", err.Error())
			return fmt.Errorf("unknown error")
		}
	}

	log.Printf("User %s successfully entered tournament %s in group %s", userID, t.TournamentID, groupID)
	return nil
}

// ClaimRewardTransaction handles the transaction logic to claim rewards
func (db *DynamoDB) ClaimRewardTransaction(ctx context.Context, userID string, reward int, tournamentID string) error {
	if svc == nil {
		return fmt.Errorf("DynamoDB client not initialized")
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			{
				Update: &dynamodb.Update{
					TableName: aws.String(usersTable),
					Key: map[string]*dynamodb.AttributeValue{
						"userId": {S: aws.String(userID)},
					},
					UpdateExpression: aws.String("SET #c = #c + :r"),
					ExpressionAttributeNames: map[string]*string{
						"#c": aws.String("coins"),
					},
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":r": {N: aws.String(fmt.Sprintf("%d", reward))},
					},
				},
			},
			{
				Update: &dynamodb.Update{
					TableName: aws.String(tournamentEntriesTable),
					Key: map[string]*dynamodb.AttributeValue{
						"tournamentId": {S: aws.String(tournamentID)},
						"userId":       {S: aws.String(userID)},
					},
					UpdateExpression: aws.String("SET #cr = :trueVal, #ca = :claimedAt"),
					ExpressionAttributeNames: map[string]*string{
						"#cr": aws.String("claimedReward"),
						"#ca": aws.String("claimedAt"),
					},
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":trueVal":   {BOOL: aws.Bool(true)},
						":claimedAt": {S: aws.String(time.Now().UTC().Format(time.RFC3339))},
						":falseVal":  {BOOL: aws.Bool(false)}, // For condition
					},
					ConditionExpression: aws.String("attribute_not_exists(#cr) OR #cr = :falseVal"),
				},
			},
		},
	}

	_, err := svc.TransactWriteItemsWithContext(ctx, input)
	if err != nil {
		// Handle specific DynamoDB errors.
		if tcErr, ok := err.(*dynamodb.TransactionCanceledException); ok {
			log.Printf("Transaction canceled: %v", tcErr.Message())
			for i, r := range tcErr.CancellationReasons {
				log.Printf("Cancellation reason %d: Code=%s, Message=%s", i, aws.StringValue(r.Code), aws.StringValue(r.Message))
			}
			for _, r := range tcErr.CancellationReasons {
				if aws.StringValue(r.Code) == "ConditionalCheckFailed" {
					return errors.ErrRewardAlreadyClaimed
				}
			}
			return fmt.Errorf("transaction canceled")
		} else if aerr, ok := err.(awserr.Error); ok {
			log.Println("DynamoDB error:", aerr.Error())
			return fmt.Errorf("database error")
		} else {
			log.Println("Unknown error:", err.Error())
			return fmt.Errorf("unknown error")
		}
	}

	return nil
}
