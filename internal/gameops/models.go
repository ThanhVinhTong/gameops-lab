package gameops

import "time"

// SessionEnded is the domain event produced when a gameplay session ends.
type SessionEnded struct {
	EventID         string    `json:"eventId"`
	UserID          string    `json:"userId"`
	GameID          string    `json:"gameId"`
	GameName        string    `json:"gameName"`
	Platform        string    `json:"platform"`
	StartedAt       time.Time `json:"startedAt"`
	EndedAt         time.Time `json:"endedAt"`
	DurationSeconds int64     `json:"durationSeconds"`
}
