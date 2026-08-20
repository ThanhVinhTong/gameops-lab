package gameops

import "context"

// AnalyticsProcessor is the future boundary for processing completed sessions.
// Analytics persistence will be added behind this boundary in a later phase.
type AnalyticsProcessor interface {
	ProcessSessionEnded(ctx context.Context, event SessionEnded) error
}
