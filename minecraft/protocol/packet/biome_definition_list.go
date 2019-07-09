package packet

import "bytes"

// BiomeDefinitionList is sent by the server to let the client know all biomes that are available and
// implemented on the server side. It is much like the AvailableEntityIdentifiers packet, but instead
// functions for biomes.
type BiomeDefinitionList struct {
	// SerialisedBiomeDefinitions is a network NBT serialised compound of all definitions of biomes that are
	// available on the server.
	SerialisedBiomeDefinitions []byte
}

// ID ...
func (*BiomeDefinitionList) ID() uint32 {
	return IDBiomeDefinitionList
}

// Marshal ...
func (pk *BiomeDefinitionList) Marshal(buf *bytes.Buffer) {
	_, _ = buf.Write(pk.SerialisedBiomeDefinitions)
}

// Unmarshal ...
func (pk *BiomeDefinitionList) Unmarshal(buf *bytes.Buffer) error {
	pk.SerialisedBiomeDefinitions = make([]byte, buf.Len())
	_, err := buf.Read(pk.SerialisedBiomeDefinitions)
	return err
}
