package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	MoveModeNormal = iota
	MoveModeReset
	MoveModeTeleport
	MoveModeRotation
)

const (
	TeleportCauseUnknown = iota
	TeleportCauseProjectile
	TeleportCauseChorusFruit
	TeleportCauseCommand
	TeleportCauseBehaviour
)

// MovePlayer is sent by players to send their movement to the server, and by the server to update the
// movement of player entities to other players.
type MovePlayer struct {
	// EntityRuntimeID is the runtime ID of the player. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// Position is the position to spawn the player on. If the player is on a distance that the viewer cannot
	// see it, the player will still show up if the viewer moves closer.
	Position mgl32.Vec3
	// Pitch is the vertical rotation of the player. Facing straight forward yields a pitch of 0. Pitch is
	// measured in degrees.
	Pitch float32
	// Yaw is the horizontal rotation of the player. Yaw is also measured in degrees.
	Yaw float32
	// HeadYaw is the same as Yaw, except that it applies specifically to the head of the player. A different
	// value for HeadYaw than Yaw means that the player will have its head turned.
	HeadYaw float32
	// Mode is the mode of the movement. It specifies the way the player's movement should be shown to other
	// players. It is one of the constants above.
	Mode byte
	// OnGround specifies if the player is considered on the ground. Note that proxies or hacked clients could
	// fake this to always be true, so it should not be taken for granted.
	OnGround bool
	// RiddenEntityRuntimeID is the runtime ID of the entity that the player might currently be riding. If not
	// riding, this should be left 0.
	RiddenEntityRuntimeID uint64
	// TeleportCause is written only if Mode is MoveModeTeleport. It specifies the cause of the teleportation,
	// which is one of the constants above.
	TeleportCause int32
	// TeleportSourceEntityType is the entity type that caused the teleportation, for example an ender pearl.
	TeleportSourceEntityType int32
}

// ID ...
func (*MovePlayer) ID() uint32 {
	return IDMovePlayer
}

// Marshal ...
func (pk *MovePlayer) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	_ = protocol.WriteVec3(buf, pk.Position)
	_ = protocol.WriteFloat32(buf, pk.Pitch)
	_ = protocol.WriteFloat32(buf, pk.Yaw)
	_ = protocol.WriteFloat32(buf, pk.HeadYaw)
	_ = binary.Write(buf, binary.LittleEndian, pk.Mode)
	_ = binary.Write(buf, binary.LittleEndian, pk.OnGround)
	_ = protocol.WriteVaruint64(buf, pk.RiddenEntityRuntimeID)
	if pk.Mode == MoveModeTeleport {
		_ = binary.Write(buf, binary.LittleEndian, pk.TeleportCause)
		_ = binary.Write(buf, binary.LittleEndian, pk.TeleportSourceEntityType)
	}
}

// Unmarshal ...
func (pk *MovePlayer) Unmarshal(buf *bytes.Buffer) error {
	if err := chainErr(
		protocol.Varuint64(buf, &pk.EntityRuntimeID),
		protocol.Vec3(buf, &pk.Position),
		protocol.Float32(buf, &pk.Pitch),
		protocol.Float32(buf, &pk.Yaw),
		protocol.Float32(buf, &pk.HeadYaw),
		binary.Read(buf, binary.LittleEndian, &pk.Mode),
		binary.Read(buf, binary.LittleEndian, &pk.OnGround),
		protocol.Varuint64(buf, &pk.RiddenEntityRuntimeID),
	); err != nil {
		return err
	}
	if pk.Mode == MoveModeTeleport {
		return chainErr(
			binary.Read(buf, binary.LittleEndian, &pk.TeleportCause),
			binary.Read(buf, binary.LittleEndian, &pk.TeleportSourceEntityType),
		)
	}
	return nil
}
