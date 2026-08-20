package aws

import (
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"testing"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"gameops-lab/internal/gameops"
)

func TestNewDynamoDBAnalyticsStoreValidatesConfiguration(t *testing.T) {
	_, err := NewDynamoDBAnalyticsStore(nil, "gameops")
	if !errors.Is(err, ErrDynamoDBClientRequired) {
		t.Fatalf(
			"NewDynamoDBAnalyticsStore(nil client) error = %v, want %v",
			err,
			ErrDynamoDBClientRequired,
		)
	}

	_, err = NewDynamoDBAnalyticsStore(&fakeUpdateItemClient{}, " ")
	if !errors.Is(err, ErrDynamoDBTableNameRequired) {
		t.Fatalf(
			"NewDynamoDBAnalyticsStore(blank table) error = %v, want %v",
			err,
			ErrDynamoDBTableNameRequired,
		)
	}
}

func TestDynamoDBAnalyticsStoreApplyContributionBuildsUpdate(t *testing.T) {
	client := &fakeUpdateItemClient{
		output: &dynamodb.UpdateItemOutput{},
	}
	store, err := NewDynamoDBAnalyticsStore(client, " gameops ")
	if err != nil {
		t.Fatalf("NewDynamoDBAnalyticsStore() unexpected error: %v", err)
	}

	contribution := awsTestAnalyticsContribution()
	ctx := context.WithValue(
		context.Background(),
		dynamoDBContextKey{},
		"message-context",
	)
	if err := store.ApplyContribution(ctx, contribution); err != nil {
		t.Fatalf("ApplyContribution() unexpected error: %v", err)
	}

	if client.calls != 1 {
		t.Fatalf("UpdateItem() calls = %d, want 1", client.calls)
	}
	if client.ctx != ctx {
		t.Error("ApplyContribution() did not forward the context")
	}
	if client.input == nil {
		t.Fatal("UpdateItem input is nil")
	}
	if got := awssdk.ToString(client.input.TableName); got != "gameops" {
		t.Errorf("TableName = %q, want %q", got, "gameops")
	}

	wantKey := map[string]dynamodbtypes.AttributeValue{
		"pk": &dynamodbtypes.AttributeValueMemberS{
			Value: "USER#" + base64.RawURLEncoding.EncodeToString(
				[]byte(contribution.UserID),
			),
		},
		"sk": &dynamodbtypes.AttributeValueMemberS{
			Value: "GAME#" + base64.RawURLEncoding.EncodeToString(
				[]byte(contribution.GameID),
			) + "#PLATFORM#" + base64.RawURLEncoding.EncodeToString(
				[]byte(contribution.Platform),
			),
		},
	}
	if !reflect.DeepEqual(client.input.Key, wantKey) {
		t.Errorf("Key = %#v, want %#v", client.input.Key, wantKey)
	}

	if got := awssdk.ToString(client.input.UpdateExpression); got != analyticsUpdateExpression {
		t.Errorf(
			"UpdateExpression = %q, want %q",
			got,
			analyticsUpdateExpression,
		)
	}

	wantNames := map[string]string{
		"#entity_type":            "entity_type",
		"#user_id":                "user_id",
		"#game_id":                "game_id",
		"#game_name":              "game_name",
		"#platform":               "platform",
		"#last_event_id":          "last_event_id",
		"#last_session_ended_at":  "last_session_ended_at",
		"#session_count":          "session_count",
		"#total_duration_seconds": "total_duration_seconds",
	}
	if !reflect.DeepEqual(client.input.ExpressionAttributeNames, wantNames) {
		t.Errorf(
			"ExpressionAttributeNames = %#v, want %#v",
			client.input.ExpressionAttributeNames,
			wantNames,
		)
	}

	wantValues := map[string]dynamodbtypes.AttributeValue{
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
			Value: "1",
		},
		":duration": &dynamodbtypes.AttributeValueMemberN{
			Value: "125",
		},
	}
	if !reflect.DeepEqual(client.input.ExpressionAttributeValues, wantValues) {
		t.Errorf(
			"ExpressionAttributeValues = %#v, want %#v",
			client.input.ExpressionAttributeValues,
			wantValues,
		)
	}
}

func TestDynamoDBAnalyticsStoreRejectsInvalidContribution(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*gameops.SessionAnalyticsContribution)
	}{
		{
			name: "missing event ID",
			mutate: func(contribution *gameops.SessionAnalyticsContribution) {
				contribution.EventID = " "
			},
		},
		{
			name: "zero session count",
			mutate: func(contribution *gameops.SessionAnalyticsContribution) {
				contribution.SessionCount = 0
			},
		},
		{
			name: "multi-session contribution",
			mutate: func(contribution *gameops.SessionAnalyticsContribution) {
				contribution.SessionCount = 2
			},
		},
		{
			name: "negative duration",
			mutate: func(contribution *gameops.SessionAnalyticsContribution) {
				contribution.TotalDurationSeconds = -1
			},
		},
		{
			name: "missing end time",
			mutate: func(contribution *gameops.SessionAnalyticsContribution) {
				contribution.LastSessionEndedAt = time.Time{}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &fakeUpdateItemClient{}
			store, err := NewDynamoDBAnalyticsStore(client, "gameops")
			if err != nil {
				t.Fatalf("NewDynamoDBAnalyticsStore() unexpected error: %v", err)
			}

			contribution := awsTestAnalyticsContribution()
			test.mutate(&contribution)

			err = store.ApplyContribution(context.Background(), contribution)
			if !errors.Is(err, ErrInvalidAnalyticsContribution) {
				t.Fatalf(
					"ApplyContribution() error = %v, want %v",
					err,
					ErrInvalidAnalyticsContribution,
				)
			}
			if client.calls != 0 {
				t.Errorf("UpdateItem() calls = %d, want 0", client.calls)
			}
		})
	}
}

func TestDynamoDBAnalyticsStorePreservesClientError(t *testing.T) {
	clientError := errors.New("DynamoDB unavailable")
	client := &fakeUpdateItemClient{
		err: clientError,
	}
	store, err := NewDynamoDBAnalyticsStore(client, "gameops")
	if err != nil {
		t.Fatalf("NewDynamoDBAnalyticsStore() unexpected error: %v", err)
	}

	err = store.ApplyContribution(
		context.Background(),
		awsTestAnalyticsContribution(),
	)
	if !errors.Is(err, clientError) {
		t.Fatalf("ApplyContribution() error = %v, want wrapped client error", err)
	}
}

type dynamoDBContextKey struct{}

type fakeUpdateItemClient struct {
	ctx    context.Context
	input  *dynamodb.UpdateItemInput
	output *dynamodb.UpdateItemOutput
	err    error
	calls  int
}

func (f *fakeUpdateItemClient) UpdateItem(
	ctx context.Context,
	input *dynamodb.UpdateItemInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.UpdateItemOutput, error) {
	f.ctx = ctx
	f.input = input
	f.calls++
	return f.output, f.err
}

func awsTestAnalyticsContribution() gameops.SessionAnalyticsContribution {
	location := time.FixedZone("AWST", 8*60*60)

	return gameops.SessionAnalyticsContribution{
		EventID:              "event#1",
		UserID:               "user#1",
		GameID:               "game/日本",
		GameName:             "日本 Game",
		Platform:             "PC + Cloud",
		SessionCount:         1,
		TotalDurationSeconds: 125,
		LastSessionEndedAt: time.Date(
			2026,
			time.August,
			21,
			12,
			30,
			0,
			123_000_000,
			location,
		),
	}
}
