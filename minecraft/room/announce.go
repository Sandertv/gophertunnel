package room

import "context"

// Announcer announces the Status of a Listener to an external service. Implementations of Announcer
// should define how to report the status using the provided Announce method.
//
// Example implementations might include XBLAnnouncer, which uses the Multiplayer Session Directory (MPSD)
// of Xbox Live for announcing the Status.
type Announcer interface {
	// Announce sends the given Status to an external service for reporting.
	// The [context.Context] may be used to control the deadline and cancellation
	// of announcement. An error may be returned, if the Status could not be announced.
	Announce(ctx context.Context, status Status) error

	Close() error
}
