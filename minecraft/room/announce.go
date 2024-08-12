package room

type Announcer interface {
	Announce(status Status) error
}
