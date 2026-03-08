package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// LocatorBar is sent by the server to synchronise waypoint/locator bar changes to the client.
type LocatorBar struct {
	// Waypoints is a list of waypoint payloads to add, remove or update.
	Waypoints []protocol.LocatorBarWaypointPayload
}

// ID ...
func (*LocatorBar) ID() uint32 {
	return IDLocatorBar
}

func (pk *LocatorBar) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Waypoints)
}
