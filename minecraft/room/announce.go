package room

import "context"

type Reference interface {
	String() string
}

type Announcer interface {
	Announce(ctx context.Context, status Status) error
	Close() error
}
