package gameops

import (
	"errors"
	"testing"
	"time"
)

func TestSessionServiceEndSessionNormalizesInput(t *testing.T) {
	input := validEndSessionInput()

	got, err := (SessionService{}).EndSession(input)
	if err != nil {
		t.Fatalf("EndSession() unexpected error: %v", err)
	}

	stringFields := []struct {
		name string
		got  string
		want string
	}{
		{name: "EventID", got: got.EventID, want: "event-1"},
		{name: "UserID", got: got.UserID, want: "user-1"},
		{name: "GameID", got: got.GameID, want: "game-1"},
		{name: "GameName", got: got.GameName, want: "Game Name"},
		{name: "Platform", got: got.Platform, want: "PC"},
	}

	for _, field := range stringFields {
		if field.got != field.want {
			t.Errorf("%s = %q, want %q", field.name, field.got, field.want)
		}
	}

	wantStartedAt := input.StartedAt.UTC().Round(0)
	wantEndedAt := input.EndedAt.UTC().Round(0)
	if got.StartedAt != wantStartedAt {
		t.Errorf("StartedAt = %v, want %v", got.StartedAt, wantStartedAt)
	}
	if got.EndedAt != wantEndedAt {
		t.Errorf("EndedAt = %v, want %v", got.EndedAt, wantEndedAt)
	}
	if got.StartedAt.Location() != time.UTC || got.EndedAt.Location() != time.UTC {
		t.Error("timestamps were not normalized to UTC")
	}
	if got.DurationSeconds != 2 {
		t.Errorf("DurationSeconds = %d, want 2", got.DurationSeconds)
	}
}

func TestSessionServiceEndSessionDurationSeconds(t *testing.T) {
	startedAt := time.Date(2026, time.August, 21, 1, 2, 3, 0, time.UTC)

	tests := []struct {
		name  string
		delta time.Duration
		want  int64
	}{
		{name: "equal timestamps", delta: 0, want: 0},
		{name: "subsecond", delta: 999 * time.Millisecond, want: 0},
		{name: "fractional seconds floor", delta: 1999 * time.Millisecond, want: 1},
		{name: "exact seconds", delta: 2 * time.Second, want: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validEndSessionInput()
			input.StartedAt = startedAt
			input.EndedAt = startedAt.Add(test.delta)

			got, err := (SessionService{}).EndSession(input)
			if err != nil {
				t.Fatalf("EndSession() unexpected error: %v", err)
			}
			if got.DurationSeconds != test.want {
				t.Errorf(
					"DurationSeconds = %d, want %d",
					got.DurationSeconds,
					test.want,
				)
			}
		})
	}
}

func TestSessionServiceEndSessionRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		field  string
		mutate func(*EndSessionInput)
	}{
		{
			name:  "event ID",
			field: "eventId",
			mutate: func(input *EndSessionInput) {
				input.EventID = " \t "
			},
		},
		{
			name:  "user ID",
			field: "userId",
			mutate: func(input *EndSessionInput) {
				input.UserID = ""
			},
		},
		{
			name:  "game ID",
			field: "gameId",
			mutate: func(input *EndSessionInput) {
				input.GameID = ""
			},
		},
		{
			name:  "game name",
			field: "gameName",
			mutate: func(input *EndSessionInput) {
				input.GameName = ""
			},
		},
		{
			name:  "platform",
			field: "platform",
			mutate: func(input *EndSessionInput) {
				input.Platform = ""
			},
		},
		{
			name:  "started at",
			field: "startedAt",
			mutate: func(input *EndSessionInput) {
				input.StartedAt = time.Time{}
			},
		},
		{
			name:  "ended at",
			field: "endedAt",
			mutate: func(input *EndSessionInput) {
				input.EndedAt = time.Time{}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validEndSessionInput()
			test.mutate(&input)

			_, err := (SessionService{}).EndSession(input)
			assertValidationError(t, err, test.field, ValidationRequired)
		})
	}
}

func TestSessionServiceEndSessionRejectsReverseTimestamps(t *testing.T) {
	input := validEndSessionInput()
	input.EndedAt = input.StartedAt.Add(-500 * time.Millisecond)

	_, err := (SessionService{}).EndSession(input)
	assertValidationError(
		t,
		err,
		"endedAt",
		ValidationInvalidTimestampOrder,
	)
}

func TestValidateSessionEndedRejectsInconsistentDuration(t *testing.T) {
	event := validSessionEndedEvent(t)
	event.DurationSeconds++

	err := validateSessionEnded(event)
	assertValidationError(
		t,
		err,
		"durationSeconds",
		ValidationInconsistentDuration,
	)
}

func validEndSessionInput() EndSessionInput {
	location := time.FixedZone("AWST", 8*60*60)
	startedAt := time.Date(
		2026,
		time.August,
		21,
		10,
		30,
		0,
		400_000_000,
		location,
	)

	return EndSessionInput{
		EventID:   " event-1 ",
		UserID:    " user-1 ",
		GameID:    " game-1 ",
		GameName:  " Game Name ",
		Platform:  " PC ",
		StartedAt: startedAt,
		EndedAt:   startedAt.Add(2900 * time.Millisecond),
	}
}

func validSessionEndedEvent(t *testing.T) SessionEnded {
	t.Helper()

	event, err := (SessionService{}).EndSession(validEndSessionInput())
	if err != nil {
		t.Fatalf("could not create valid SessionEnded fixture: %v", err)
	}

	return event
}

func assertValidationError(
	t *testing.T,
	err error,
	wantField string,
	wantCode ValidationCode,
) {
	t.Helper()

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("errors.Is(err, ErrInvalidSession) = false: %v", err)
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("errors.As(err, *ValidationError) = false: %T: %v", err, err)
	}
	if validationErr.Field != wantField {
		t.Errorf("ValidationError.Field = %q, want %q", validationErr.Field, wantField)
	}
	if validationErr.Code != wantCode {
		t.Errorf("ValidationError.Code = %q, want %q", validationErr.Code, wantCode)
	}
}
