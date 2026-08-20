package gameops

import "time"

// SessionAnalyticsContribution is one validated session's contribution to
// future aggregated analytics.
type SessionAnalyticsContribution struct {
	EventID              string
	UserID               string
	GameID               string
	GameName             string
	Platform             string
	SessionCount         int64
	TotalDurationSeconds int64
	LastSessionEndedAt   time.Time
}

// AnalyticsService contains framework-independent analytics application logic.
// Its zero value is ready to use.
type AnalyticsService struct{}

// ProcessSessionEnded validates an event and derives its analytics contribution.
// It performs no persistence, idempotency handling, or other I/O.
func (AnalyticsService) ProcessSessionEnded(event SessionEnded) (SessionAnalyticsContribution, error) {
	if err := validateSessionEnded(event); err != nil {
		return SessionAnalyticsContribution{}, err
	}

	return SessionAnalyticsContribution{
		EventID:              event.EventID,
		UserID:               event.UserID,
		GameID:               event.GameID,
		GameName:             event.GameName,
		Platform:             event.Platform,
		SessionCount:         1,
		TotalDurationSeconds: event.DurationSeconds,
		LastSessionEndedAt:   event.EndedAt,
	}, nil
}
