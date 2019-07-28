package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ActorFall is sent by the client when it falls from a distance onto a block that would damage the player.
// This packet should not be used at all by the server, as it can easily be spoofed using a proxy or custom
// client. Servers should implement fall damage using their own calculations.
type ActorFall struct {
	// EntityNetworkID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// FallDistance is the distance that the entity fell until it hit the ground. The damage would otherwise
	// be calculated using this field.
	FallDistance float32
	// InVoid specifies if the fall was in the void. The player can't fall below roughly Y=-40.
	InVoid bool
}

// ID ...
func (*ActorFall) ID() uint32 {
	return IDActorFall
}

// Marshal ...
func (pk *ActorFall) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	_ = protocol.WriteFloat32(buf, pk.FallDistance)
	_ = binary.Write(buf, binary.LittleEndian, pk.InVoid)
}

// Unmarshal ...
func (pk *ActorFall) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.Varuint64(buf, &pk.EntityRuntimeID),
		protocol.Float32(buf, &pk.FallDistance),
		binary.Read(buf, binary.LittleEndian, &pk.InVoid),
	)
}
