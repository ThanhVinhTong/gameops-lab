package aws

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	eventbridgetypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"

	"gameops-lab/internal/gameops"
)

func TestNewEventBridgePublisherValidatesConfiguration(t *testing.T) {
	_, err := NewEventBridgePublisher(nil, "gameops-events")
	if !errors.Is(err, ErrEventBridgeClientRequired) {
		t.Fatalf(
			"NewEventBridgePublisher(nil client) error = %v, want %v",
			err,
			ErrEventBridgeClientRequired,
		)
	}

	_, err = NewEventBridgePublisher(&fakePutEventsClient{}, " ")
	if !errors.Is(err, ErrEventBusNameRequired) {
		t.Fatalf(
			"NewEventBridgePublisher(blank bus) error = %v, want %v",
			err,
			ErrEventBusNameRequired,
		)
	}
}

func TestEventBridgePublisherPublishesSessionEnded(t *testing.T) {
	client := &fakePutEventsClient{
		output: &eventbridge.PutEventsOutput{
			Entries: []eventbridgetypes.PutEventsResultEntry{
				{
					EventId: awssdk.String("eventbridge-event-id"),
				},
			},
		},
	}
	publisher, err := NewEventBridgePublisher(client, " gameops-events ")
	if err != nil {
		t.Fatalf("NewEventBridgePublisher() unexpected error: %v", err)
	}

	event := awsTestSessionEnded()
	ctx := context.WithValue(
		context.Background(),
		eventBridgeContextKey{},
		"request-context",
	)
	if err := publisher.PublishSessionEnded(ctx, event); err != nil {
		t.Fatalf("PublishSessionEnded() unexpected error: %v", err)
	}

	if client.calls != 1 {
		t.Fatalf("PutEvents() calls = %d, want 1", client.calls)
	}
	if client.ctx != ctx {
		t.Error("PublishSessionEnded() did not forward the context")
	}
	if client.input == nil || len(client.input.Entries) != 1 {
		t.Fatalf("PutEvents input = %#v, want one entry", client.input)
	}

	entry := client.input.Entries[0]
	if got := awssdk.ToString(entry.EventBusName); got != "gameops-events" {
		t.Errorf("EventBusName = %q, want %q", got, "gameops-events")
	}
	if got := awssdk.ToString(entry.Source); got != sessionEndedSource {
		t.Errorf("Source = %q, want %q", got, sessionEndedSource)
	}
	if got := awssdk.ToString(entry.DetailType); got != sessionEndedDetailType {
		t.Errorf("DetailType = %q, want %q", got, sessionEndedDetailType)
	}

	var decoded gameops.SessionEnded
	if err := json.Unmarshal([]byte(awssdk.ToString(entry.Detail)), &decoded); err != nil {
		t.Fatalf("unmarshal EventBridge detail: %v", err)
	}
	if !reflect.DeepEqual(decoded, event) {
		t.Errorf("decoded detail = %#v, want %#v", decoded, event)
	}
}

func TestEventBridgePublisherRejectsFailedResponses(t *testing.T) {
	tests := []struct {
		name   string
		output *eventbridge.PutEventsOutput
	}{
		{
			name: "failed entry",
			output: &eventbridge.PutEventsOutput{
				FailedEntryCount: 1,
				Entries: []eventbridgetypes.PutEventsResultEntry{
					{
						ErrorCode:    awssdk.String("InternalFailure"),
						ErrorMessage: awssdk.String("failed"),
					},
				},
			},
		},
		{
			name: "entry error despite zero failed count",
			output: &eventbridge.PutEventsOutput{
				Entries: []eventbridgetypes.PutEventsResultEntry{
					{
						ErrorCode: awssdk.String("MalformedDetail"),
					},
				},
			},
		},
		{
			name:   "nil response",
			output: nil,
		},
		{
			name: "missing result entry",
			output: &eventbridge.PutEventsOutput{
				Entries: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &fakePutEventsClient{
				output: test.output,
			}
			publisher, err := NewEventBridgePublisher(client, "gameops-events")
			if err != nil {
				t.Fatalf("NewEventBridgePublisher() unexpected error: %v", err)
			}

			err = publisher.PublishSessionEnded(
				context.Background(),
				awsTestSessionEnded(),
			)
			if !errors.Is(err, ErrEventPublishRejected) {
				t.Fatalf(
					"PublishSessionEnded() error = %v, want %v",
					err,
					ErrEventPublishRejected,
				)
			}
		})
	}
}

func TestEventBridgePublisherPreservesClientError(t *testing.T) {
	clientError := errors.New("transport unavailable")
	client := &fakePutEventsClient{
		err: clientError,
	}
	publisher, err := NewEventBridgePublisher(client, "gameops-events")
	if err != nil {
		t.Fatalf("NewEventBridgePublisher() unexpected error: %v", err)
	}

	err = publisher.PublishSessionEnded(
		context.Background(),
		awsTestSessionEnded(),
	)
	if !errors.Is(err, clientError) {
		t.Fatalf("PublishSessionEnded() error = %v, want wrapped client error", err)
	}
}

type eventBridgeContextKey struct{}

type fakePutEventsClient struct {
	ctx    context.Context
	input  *eventbridge.PutEventsInput
	output *eventbridge.PutEventsOutput
	err    error
	calls  int
}

func (f *fakePutEventsClient) PutEvents(
	ctx context.Context,
	input *eventbridge.PutEventsInput,
	_ ...func(*eventbridge.Options),
) (*eventbridge.PutEventsOutput, error) {
	f.ctx = ctx
	f.input = input
	f.calls++
	return f.output, f.err
}

func awsTestSessionEnded() gameops.SessionEnded {
	startedAt := time.Date(
		2026,
		time.August,
		21,
		4,
		5,
		6,
		0,
		time.UTC,
	)

	return gameops.SessionEnded{
		EventID:         "event-1",
		UserID:          "user-1",
		GameID:          "game-1",
		GameName:        "Game Name",
		Platform:        "PC",
		StartedAt:       startedAt,
		EndedAt:         startedAt.Add(125 * time.Second),
		DurationSeconds: 125,
	}
}
