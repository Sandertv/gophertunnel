package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
)

const (
	MoveActorDeltaFlagHasX = 1 << iota
	MoveActorDeltaFlagHasY
	MoveActorDeltaFlagHasZ
	MoveActorDeltaFlagHasRotX
	MoveActorDeltaFlagHasRotY
	MoveActorDeltaFlagHasRotZ
	MoveActorDeltaFlagOnGround
	MoveActorDeltaFlagTeleport
	MoveActorDeltaFlagForceMove
)

// MoveActorDelta is sent by the server to move an entity by a given delta. The packet is specifically
// optimised to save as much space as possible, by only writing non-zero fields.
type MoveActorDelta struct {
	// Flags is a list of flags that specify what data is in the packet.
	Flags uint16
	// EntityRuntimeID is the runtime ID of the entity that is being moved. The packet works provided a
	// non-player entity with this runtime ID is present.
	EntityRuntimeID uint64
	// DeltaPosition is the difference from the previous position to the new position. It is the distance on
	// each axis that the entity should be moved.
	DeltaPosition mgl32.Vec3
	// Rotation is the new absolute rotation. Unlike the position, it is not actually a delta. If any of the
	// values of this rotation are not sent, these values are 0 and no flag for them is present.
	Rotation mgl32.Vec3
}

// ID ...
func (*MoveActorDelta) ID() uint32 {
	return IDMoveActorDelta
}

// Marshal ...
func (pk *MoveActorDelta) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	_ = binary.Write(buf, binary.LittleEndian, pk.Flags)
	if pk.Flags&MoveActorDeltaFlagHasX != 0 {
		_ = protocol.WriteVarint32(buf, int32(math.Float32bits(pk.DeltaPosition[0])))
	}
	if pk.Flags&MoveActorDeltaFlagHasY != 0 {
		_ = protocol.WriteVarint32(buf, int32(math.Float32bits(pk.DeltaPosition[1])))
	}
	if pk.Flags&MoveActorDeltaFlagHasZ != 0 {
		_ = protocol.WriteVarint32(buf, int32(math.Float32bits(pk.DeltaPosition[2])))
	}
	if pk.Flags&MoveActorDeltaFlagHasRotX != 0 {
		_ = binary.Write(buf, binary.LittleEndian, byte(float32(math.Mod(float64(pk.Rotation[0]/(360.0/256.0)), 360))))
	}
	if pk.Flags&MoveActorDeltaFlagHasRotY != 0 {
		_ = binary.Write(buf, binary.LittleEndian, byte(float32(math.Mod(float64(pk.Rotation[1]/(360.0/256.0)), 360))))
	}
	if pk.Flags&MoveActorDeltaFlagHasRotZ != 0 {
		_ = binary.Write(buf, binary.LittleEndian, byte(float32(math.Mod(float64(pk.Rotation[2]/(360.0/256.0)), 360))))
	}
}

// Unmarshal ...
func (pk *MoveActorDelta) Unmarshal(r *protocol.Reader) {
	pk.DeltaPosition = mgl32.Vec3{}
	pk.Rotation = mgl32.Vec3{}
	r.Varuint64(&pk.EntityRuntimeID)
	r.Uint16(&pk.Flags)

	var v int32
	if pk.Flags&MoveActorDeltaFlagHasX != 0 {
		r.Varint32(&v)
		pk.DeltaPosition[0] = math.Float32frombits(uint32(v))
	}
	if pk.Flags&MoveActorDeltaFlagHasY != 0 {
		r.Varint32(&v)
		pk.DeltaPosition[1] = math.Float32frombits(uint32(v))
	}
	if pk.Flags&MoveActorDeltaFlagHasZ != 0 {
		r.Varint32(&v)
		pk.DeltaPosition[2] = math.Float32frombits(uint32(v))
	}
	if pk.Flags&MoveActorDeltaFlagHasRotX != 0 {
		r.ByteFloat(&pk.Rotation[0])
	}
	if pk.Flags&MoveActorDeltaFlagHasRotY != 0 {
		r.ByteFloat(&pk.Rotation[1])
	}
	if pk.Flags&MoveActorDeltaFlagHasRotZ != 0 {
		r.ByteFloat(&pk.Rotation[2])
	}
}
