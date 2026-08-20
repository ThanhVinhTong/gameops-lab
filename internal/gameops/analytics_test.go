package gameops

import (
	"reflect"
	"testing"
)

func TestAnalyticsServiceProcessSessionEndedMapsContribution(t *testing.T) {
	event := validSessionEndedEvent(t)

	got, err := (AnalyticsService{}).ProcessSessionEnded(event)
	if err != nil {
		t.Fatalf("ProcessSessionEnded() unexpected error: %v", err)
	}

	want := SessionAnalyticsContribution{
		EventID:              event.EventID,
		UserID:               event.UserID,
		GameID:               event.GameID,
		GameName:             event.GameName,
		Platform:             event.Platform,
		SessionCount:         1,
		TotalDurationSeconds: event.DurationSeconds,
		LastSessionEndedAt:   event.EndedAt,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ProcessSessionEnded() = %#v, want %#v", got, want)
	}
}

func TestAnalyticsServiceProcessSessionEndedRejectsInvalidEvents(t *testing.T) {
	tests := []struct {
		name   string
		field  string
		code   ValidationCode
		mutate func(*SessionEnded)
	}{
		{
			name:  "missing event ID",
			field: "eventId",
			code:  ValidationRequired,
			mutate: func(event *SessionEnded) {
				event.EventID = " "
			},
		},
		{
			name:  "inconsistent duration",
			field: "durationSeconds",
			code:  ValidationInconsistentDuration,
			mutate: func(event *SessionEnded) {
				event.DurationSeconds++
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := validSessionEndedEvent(t)
			test.mutate(&event)

			got, err := (AnalyticsService{}).ProcessSessionEnded(event)
			assertValidationError(t, err, test.field, test.code)

			if !reflect.DeepEqual(got, SessionAnalyticsContribution{}) {
				t.Fatalf("ProcessSessionEnded() result = %#v, want zero value", got)
			}
		})
	}
}
