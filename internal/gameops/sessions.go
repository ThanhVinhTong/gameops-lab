package gameops

import (
	"errors"
	"strings"
	"time"
)

// ErrInvalidSession identifies input that cannot produce a valid SessionEnded
// event.
var ErrInvalidSession = errors.New("invalid session")

// ValidationCode identifies why a session field failed validation.
type ValidationCode string

const (
	ValidationRequired              ValidationCode = "required"
	ValidationInvalidTimestampOrder ValidationCode = "invalid_timestamp_order"
	ValidationInconsistentDuration  ValidationCode = "inconsistent_duration"
)

// ValidationError describes one invalid session field.
type ValidationError struct {
	Field string
	Code  ValidationCode
}

func (e *ValidationError) Error() string {
	return ErrInvalidSession.Error() + ": " + e.Field + " " + string(e.Code)
}

// Unwrap allows callers to identify all validation failures with
// errors.Is(err, ErrInvalidSession).
func (e *ValidationError) Unwrap() error {
	return ErrInvalidSession
}

// EndSessionInput contains the values needed to construct a SessionEnded event.
// Event IDs are supplied by the caller so ID generation remains outside the
// domain service.
type EndSessionInput struct {
	EventID   string
	UserID    string
	GameID    string
	GameName  string
	Platform  string
	StartedAt time.Time
	EndedAt   time.Time
}

// SessionService contains framework-independent session application logic.
// Its zero value is ready to use.
type SessionService struct{}

// EndSession normalizes and validates input, then constructs a SessionEnded
// event. It performs no publishing or other I/O.
func (SessionService) EndSession(input EndSessionInput) (SessionEnded, error) {
	event := SessionEnded{
		EventID:   strings.TrimSpace(input.EventID),
		UserID:    strings.TrimSpace(input.UserID),
		GameID:    strings.TrimSpace(input.GameID),
		GameName:  strings.TrimSpace(input.GameName),
		Platform:  strings.TrimSpace(input.Platform),
		StartedAt: input.StartedAt.UTC().Round(0),
		EndedAt:   input.EndedAt.UTC().Round(0),
	}

	if !event.StartedAt.IsZero() && !event.EndedAt.IsZero() {
		event.DurationSeconds = elapsedWholeSeconds(event.StartedAt, event.EndedAt)
	}

	if err := validateSessionEnded(event); err != nil {
		return SessionEnded{}, err
	}

	return event, nil
}

// validateSessionEnded verifies the invariants shared by application logic.
// SessionService.EndSession remains the public canonical construction boundary.
func validateSessionEnded(event SessionEnded) error {
	requiredFields := []struct {
		name  string
		value string
	}{
		{name: "eventId", value: event.EventID},
		{name: "userId", value: event.UserID},
		{name: "gameId", value: event.GameID},
		{name: "gameName", value: event.GameName},
		{name: "platform", value: event.Platform},
	}

	for _, field := range requiredFields {
		if strings.TrimSpace(field.value) == "" {
			return newValidationError(field.name, ValidationRequired)
		}
	}

	if event.StartedAt.IsZero() {
		return newValidationError("startedAt", ValidationRequired)
	}
	if event.EndedAt.IsZero() {
		return newValidationError("endedAt", ValidationRequired)
	}
	if event.EndedAt.Before(event.StartedAt) {
		return newValidationError("endedAt", ValidationInvalidTimestampOrder)
	}

	expectedDuration := elapsedWholeSeconds(event.StartedAt, event.EndedAt)
	if event.DurationSeconds != expectedDuration {
		return newValidationError("durationSeconds", ValidationInconsistentDuration)
	}

	return nil
}

func elapsedWholeSeconds(startedAt, endedAt time.Time) int64 {
	return int64(endedAt.Sub(startedAt) / time.Second)
}

func newValidationError(field string, code ValidationCode) error {
	return &ValidationError{
		Field: field,
		Code:  code,
	}
}
