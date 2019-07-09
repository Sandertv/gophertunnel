package packet

import "bytes"

// AvailableEntityIdentifiers is sent by the server at the start of the game to let the client know all
// entities that are available on the server.
type AvailableEntityIdentifiers struct {
	// SerialisedEntityIdentifiers is a network NBT serialised compound of all entity identifiers that are
	// available in the server.
	SerialisedEntityIdentifiers []byte
}

// ID ...
func (*AvailableEntityIdentifiers) ID() uint32 {
	return IDAvailableEntityIdentifiers
}

// Marshal ...
func (pk *AvailableEntityIdentifiers) Marshal(buf *bytes.Buffer) {
	_, _ = buf.Write(pk.SerialisedEntityIdentifiers)
}

// Unmarshal ...
func (pk *AvailableEntityIdentifiers) Unmarshal(buf *bytes.Buffer) error {
	pk.SerialisedEntityIdentifiers = make([]byte, buf.Len())
	_, err := buf.Read(pk.SerialisedEntityIdentifiers)
	return err
}
