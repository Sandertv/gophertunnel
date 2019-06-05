package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	SpawnTypePlayer = iota
	SpawnTypeWorld
)

// SetSpawnPosition is sent by the server to update the spawn position of a player, for example when sleeping
// in a bed.
type SetSpawnPosition struct {
	// SpawnType is the type of spawn to set. It is either SpawnTypePlayer or SpawnTypeWorld, and specifies
	// the behaviour of the spawn set. If SpawnTypeWorld is set, the position to which compasses will point is
	// also changed.
	SpawnType int32
	// Position is the new position of the spawn that was set. If SpawnType is SpawnTypeWorld, compasses will
	// point to this position.
	Position protocol.BlockPos
	// SpawnForced specifies if the spawn is forced.
	SpawnForced bool
}

// ID ...
func (*SetSpawnPosition) ID() uint32 {
	return IDSetSpawnPosition
}

// Marshal ...
func (pk *SetSpawnPosition) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint32(buf, pk.SpawnType)
	_ = protocol.WriteUBlockPosition(buf, pk.Position)
	_ = binary.Write(buf, binary.LittleEndian, pk.SpawnForced)
}

// Unmarshal ...
func (pk *SetSpawnPosition) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.Varint32(buf, &pk.SpawnType),
		protocol.UBlockPosition(buf, &pk.Position),
		binary.Read(buf, binary.LittleEndian, &pk.SpawnForced),
	)
}
