package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayerLocationTypeCoordinates = iota
	PlayerLocationTypeHide
)

// PlayerLocation is sent by the server to the client to either update a player's position on the locator bar,
// or remove them completely. The client will determine how to render the player on the locator bar based on
// their own distance to Position.
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
