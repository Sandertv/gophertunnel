package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
)

const (
	moveFlagHasX = 1 << iota
	moveFlagHasY
	moveFlagHasZ
	moveFlagHasRotX
	moveFlagHasRotY
	moveFlagHasRotZ
)

// MoveEntityDelta is sent by the server to move an entity by a given delta. The packet is specifically
// optimised to save as much space as possible, by only writing non-zero fields.
type MoveEntityDelta struct {
	// EntityRuntimeID is the runtime ID of the entity that is being moved. The packet works provided a
	// non-player entity with this runtime ID is present.
	EntityRuntimeID uint64
	// DeltaPosition is the difference from the previous position to the new position. It is the distance on
	// each axis that the entity should be moved.
	DeltaPosition mgl32.Vec3
	// DeltaRotation is the difference from the previous rotation to the new rotation. It is the rotation on
	// each axis that the entity should be turned.
	DeltaRotation mgl32.Vec3
}

// ID ...
func (*MoveEntityDelta) ID() uint32 {
	return IDMoveEntityDelta
}

// Marshal ...
func (pk *MoveEntityDelta) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	var flags byte
	func(checks ...float32) {
		for i, check := range checks {
			if check != 0 {
				flags |= 1 << byte(i)
			}
		}
	}(pk.DeltaPosition[0], pk.DeltaPosition[1], pk.DeltaPosition[2], pk.DeltaRotation[0], pk.DeltaRotation[1], pk.DeltaRotation[2])
	_ = binary.Write(buf, binary.LittleEndian, flags)
	if pk.DeltaPosition[0] != 0 {
		_ = protocol.WriteVarint32(buf, int32(math.Float32bits(pk.DeltaPosition[0])))
	}
	if pk.DeltaPosition[1] != 0 {
		_ = protocol.WriteVarint32(buf, int32(math.Float32bits(pk.DeltaPosition[1])))
	}
	if pk.DeltaPosition[2] != 0 {
		_ = protocol.WriteVarint32(buf, int32(math.Float32bits(pk.DeltaPosition[2])))
	}
	if pk.DeltaRotation[0] != 0 {
		_ = binary.Write(buf, binary.LittleEndian, byte(pk.DeltaRotation[0]/(360.0/256.0)))
	}
	if pk.DeltaRotation[1] != 0 {
		_ = binary.Write(buf, binary.LittleEndian, byte(pk.DeltaRotation[1]/(360.0/256.0)))
	}
	if pk.DeltaRotation[2] != 0 {
		_ = binary.Write(buf, binary.LittleEndian, byte(pk.DeltaRotation[2]/(360.0/256.0)))
	}
}

// Unmarshal ...
func (pk *MoveEntityDelta) Unmarshal(buf *bytes.Buffer) error {
	pk.DeltaPosition = mgl32.Vec3{}
	pk.DeltaRotation = mgl32.Vec3{}
	var flags byte
	if err := chainErr(
		protocol.Varuint64(buf, &pk.EntityRuntimeID),
		binary.Read(buf, binary.LittleEndian, &flags),
	); err != nil {
		return err
	}
	var v int32
	if flags&moveFlagHasX != 0 {
		if err := protocol.Varint32(buf, &v); err != nil {
			return err
		}
		pk.DeltaPosition[0] = math.Float32frombits(uint32(v))
	}
	if flags&moveFlagHasY != 0 {
		if err := protocol.Varint32(buf, &v); err != nil {
			return err
		}
		pk.DeltaPosition[1] = math.Float32frombits(uint32(v))
	}
	if flags&moveFlagHasZ != 0 {
		if err := protocol.Varint32(buf, &v); err != nil {
			return err
		}
		pk.DeltaPosition[2] = math.Float32frombits(uint32(v))
	}
	var w byte
	if flags&moveFlagHasRotX != 0 {
		if err := binary.Read(buf, binary.LittleEndian, &w); err != nil {
			return err
		}
		pk.DeltaRotation[0] = float32(w) * (360.0 / 256.0)
	}
	if flags&moveFlagHasRotY != 0 {
		if err := binary.Read(buf, binary.LittleEndian, &w); err != nil {
			return err
		}
		pk.DeltaRotation[1] = float32(w) * (360.0 / 256.0)
	}
	if flags&moveFlagHasRotZ != 0 {
		if err := binary.Read(buf, binary.LittleEndian, &w); err != nil {
			return err
		}
		pk.DeltaRotation[2] = float32(w) * (360.0 / 256.0)
	}
	return nil
}
