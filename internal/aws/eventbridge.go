package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	eventbridgetypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"

	"gameops-lab/internal/gameops"
)

const (
	sessionEndedSource     = "gameops.session-api"
	sessionEndedDetailType = "SessionEnded"
)

var (
	ErrEventBridgeClientRequired = errors.New("EventBridge client is required")
	ErrEventBusNameRequired      = errors.New("EventBridge event bus name is required")
	ErrEventPublishRejected      = errors.New("EventBridge rejected the session event")
)

// EventBridgePutEventsAPI is the EventBridge operation used by the publisher.
type EventBridgePutEventsAPI interface {
	PutEvents(
		ctx context.Context,
		params *eventbridge.PutEventsInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.PutEventsOutput, error)
}

var _ EventBridgePutEventsAPI = (*eventbridge.Client)(nil)

// EventBridgePublisher publishes SessionEnded events to a custom event bus.
type EventBridgePublisher struct {
	client       EventBridgePutEventsAPI
	eventBusName string
}

// NewEventBridgePublisher validates publisher configuration.
func NewEventBridgePublisher(
	client EventBridgePutEventsAPI,
	eventBusName string,
) (*EventBridgePublisher, error) {
	if client == nil {
		return nil, ErrEventBridgeClientRequired
	}

	eventBusName = strings.TrimSpace(eventBusName)
	if eventBusName == "" {
		return nil, ErrEventBusNameRequired
	}

	return &EventBridgePublisher{
		client:       client,
		eventBusName: eventBusName,
	}, nil
}

// PublishSessionEnded sends one event and verifies EventBridge accepted its
// corresponding result entry.
func (p *EventBridgePublisher) PublishSessionEnded(
	ctx context.Context,
	event gameops.SessionEnded,
) error {
	detail, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal SessionEnded event detail: %w", err)
	}

	output, err := p.client.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: []eventbridgetypes.PutEventsRequestEntry{
			{
				EventBusName: awssdk.String(p.eventBusName),
				Source:       awssdk.String(sessionEndedSource),
				DetailType:   awssdk.String(sessionEndedDetailType),
				Detail:       awssdk.String(string(detail)),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("put SessionEnded event: %w", err)
	}
	if output == nil {
		return fmt.Errorf("%w: empty response", ErrEventPublishRejected)
	}
	if len(output.Entries) != 1 {
		return fmt.Errorf(
			"%w: expected one result entry, received %d",
			ErrEventPublishRejected,
			len(output.Entries),
		)
	}

	entry := output.Entries[0]
	if output.FailedEntryCount != 0 ||
		entry.ErrorCode != nil ||
		entry.ErrorMessage != nil {
		return fmt.Errorf(
			"%w: failed entries=%d code=%q message=%q",
			ErrEventPublishRejected,
			output.FailedEntryCount,
			awssdk.ToString(entry.ErrorCode),
			awssdk.ToString(entry.ErrorMessage),
		)
	}

	return nil
}
