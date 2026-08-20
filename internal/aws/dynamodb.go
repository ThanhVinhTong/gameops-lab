package aws

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"gameops-lab/internal/gameops"
)

const analyticsUpdateExpression = "SET " +
	"#entity_type = :entity_type, " +
	"#user_id = :user_id, " +
	"#game_id = :game_id, " +
	"#game_name = :game_name, " +
	"#platform = :platform, " +
	"#last_event_id = :event_id, " +
	"#last_session_ended_at = :ended_at " +
	"ADD #session_count :session_count, " +
	"#total_duration_seconds :duration"

var (
	ErrDynamoDBClientRequired       = errors.New("DynamoDB client is required")
	ErrDynamoDBTableNameRequired    = errors.New("DynamoDB table name is required")
	ErrInvalidAnalyticsContribution = errors.New("invalid analytics contribution")
)

// DynamoDBUpdateItemAPI is the DynamoDB operation used by the analytics store.
type DynamoDBUpdateItemAPI interface {
	UpdateItem(
		ctx context.Context,
		params *dynamodb.UpdateItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.UpdateItemOutput, error)
}

var _ DynamoDBUpdateItemAPI = (*dynamodb.Client)(nil)

// DynamoDBAnalyticsStore applies analytics contributions to aggregate items.
type DynamoDBAnalyticsStore struct {
	client    DynamoDBUpdateItemAPI
	tableName string
}

// NewDynamoDBAnalyticsStore validates store configuration.
func NewDynamoDBAnalyticsStore(
	client DynamoDBUpdateItemAPI,
	tableName string,
) (*DynamoDBAnalyticsStore, error) {
	if client == nil {
		return nil, ErrDynamoDBClientRequired
	}

	tableName = strings.TrimSpace(tableName)
	if tableName == "" {
		return nil, ErrDynamoDBTableNameRequired
	}

	return &DynamoDBAnalyticsStore{
		client:    client,
		tableName: tableName,
	}, nil
}

// ApplyContribution atomically increments one user/game/platform aggregate.
// It intentionally does not provide idempotency in this phase.
func (s *DynamoDBAnalyticsStore) ApplyContribution(
	ctx context.Context,
	contribution gameops.SessionAnalyticsContribution,
) error {
	if err := validateAnalyticsContribution(contribution); err != nil {
		return err
	}

	input := &dynamodb.UpdateItemInput{
		TableName: awssdk.String(s.tableName),
		Key: map[string]dynamodbtypes.AttributeValue{
			"pk": &dynamodbtypes.AttributeValueMemberS{
				Value: "USER#" + encodeKeyPart(contribution.UserID),
			},
			"sk": &dynamodbtypes.AttributeValueMemberS{
				Value: "GAME#" + encodeKeyPart(contribution.GameID) +
					"#PLATFORM#" + encodeKeyPart(contribution.Platform),
			},
		},
		UpdateExpression: awssdk.String(analyticsUpdateExpression),
		ExpressionAttributeNames: map[string]string{
			"#entity_type":            "entity_type",
			"#user_id":                "user_id",
			"#game_id":                "game_id",
			"#game_name":              "game_name",
			"#platform":               "platform",
			"#last_event_id":          "last_event_id",
			"#last_session_ended_at":  "last_session_ended_at",
			"#session_count":          "session_count",
			"#total_duration_seconds": "total_duration_seconds",
		},
		ExpressionAttributeValues: map[string]dynamodbtypes.AttributeValue{
			":entity_type": &dynamodbtypes.AttributeValueMemberS{
				Value: "session_analytics",
			},
			":user_id": &dynamodbtypes.AttributeValueMemberS{
				Value: contribution.UserID,
			},
			":game_id": &dynamodbtypes.AttributeValueMemberS{
				Value: contribution.GameID,
			},
			":game_name": &dynamodbtypes.AttributeValueMemberS{
				Value: contribution.GameName,
			},
			":platform": &dynamodbtypes.AttributeValueMemberS{
				Value: contribution.Platform,
			},
			":event_id": &dynamodbtypes.AttributeValueMemberS{
				Value: contribution.EventID,
			},
			":ended_at": &dynamodbtypes.AttributeValueMemberS{
				Value: contribution.LastSessionEndedAt.UTC().Format(time.RFC3339Nano),
			},
			":session_count": &dynamodbtypes.AttributeValueMemberN{
				Value: strconv.FormatInt(contribution.SessionCount, 10),
			},
			":duration": &dynamodbtypes.AttributeValueMemberN{
				Value: strconv.FormatInt(contribution.TotalDurationSeconds, 10),
			},
		},
	}

	if _, err := s.client.UpdateItem(ctx, input); err != nil {
		return fmt.Errorf("update session analytics: %w", err)
	}

	return nil
}

func validateAnalyticsContribution(
	contribution gameops.SessionAnalyticsContribution,
) error {
	requiredFields := []struct {
		name  string
		value string
	}{
		{name: "event ID", value: contribution.EventID},
		{name: "user ID", value: contribution.UserID},
		{name: "game ID", value: contribution.GameID},
		{name: "game name", value: contribution.GameName},
		{name: "platform", value: contribution.Platform},
	}

	for _, field := range requiredFields {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf(
				"%w: %s is required",
				ErrInvalidAnalyticsContribution,
				field.name,
			)
		}
	}
	if contribution.SessionCount != 1 {
		return fmt.Errorf(
			"%w: session count must equal one",
			ErrInvalidAnalyticsContribution,
		)
	}
	if contribution.TotalDurationSeconds < 0 {
		return fmt.Errorf(
			"%w: total duration cannot be negative",
			ErrInvalidAnalyticsContribution,
		)
	}
	if contribution.LastSessionEndedAt.IsZero() {
		return fmt.Errorf(
			"%w: last session end time is required",
			ErrInvalidAnalyticsContribution,
		)
	}

	return nil
}

func encodeKeyPart(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}
