package packet

import (
	"bytes"
	"math"
)

// AvailableActorIdentifiers is sent by the server at the start of the game to let the client know all
// entities that are available on the server.
type AvailableActorIdentifiers struct {
	// SerialisedEntityIdentifiers is a network NBT serialised compound of all entity identifiers that are
	// available in the server.
	SerialisedEntityIdentifiers []byte
}

// ID ...
func (*AvailableActorIdentifiers) ID() uint32 {
	return IDAvailableActorIdentifiers
}

// Marshal ...
func (pk *AvailableActorIdentifiers) Marshal(buf *bytes.Buffer) {
	_, _ = buf.Write(pk.SerialisedEntityIdentifiers)
}

// Unmarshal ...
func (pk *AvailableActorIdentifiers) Unmarshal(buf *bytes.Buffer) error {
	pk.SerialisedEntityIdentifiers = buf.Next(math.MaxInt32)
	return nil
}
