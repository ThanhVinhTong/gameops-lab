package gameops

import "context"

// SessionIngestor is the future boundary for accepting completed sessions.
// Event publishing will be added behind this boundary in a later phase.
type SessionIngestor interface {
	IngestSessionEnded(ctx context.Context, event SessionEnded) error
}
