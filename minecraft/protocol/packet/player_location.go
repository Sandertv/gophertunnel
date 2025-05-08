package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayerLocationTypeCoordinates = iota
	PlayerLocationTypeHide
)

// PlayerLocation is sent by the client to the server when it updates its own skin using the in-game skin picker.
// It is relayed by the server, or sent if the server changes the skin of a player on its own accord. Note
// that the packet can only be sent for players that are in the player list at the time of sending.
type PlayerLocation struct {
	// Type is the action that is being performed. It is one of the constants above.
	Type int32
	// EntityUniqueID is the unique ID of the entity. The unique ID is a value that remains consistent across
	// different sessions of the same world.
	EntityUniqueID int64
	// Position is the position of the player to be used on the locator bar. This is only set when the Type is
	// PlayerLocationTypeCoordinates.
	Position mgl32.Vec3
}

// ID ...
func (*PlayerLocation) ID() uint32 {
	return IDPlayerLocation
}

func (pk *PlayerLocation) Marshal(io protocol.IO) {
	io.Int32(&pk.Type)
	io.Varint64(&pk.EntityUniqueID)
	if pk.Type == PlayerLocationTypeCoordinates {
		io.Vec3(&pk.Position)
	}
}
