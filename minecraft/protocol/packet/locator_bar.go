package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// LocatorBar is sent by the server to add, remove or update waypoints on the client's locator bar.
type LocatorBar struct {
	// Waypoints is a slice of waypoints to add, remove or update.
	Waypoints []protocol.LocatorBarWaypoint
}

// ID ...
func (*LocatorBar) ID() uint32 {
	return IDLocatorBar
}

func (pk *LocatorBar) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Waypoints)
}
