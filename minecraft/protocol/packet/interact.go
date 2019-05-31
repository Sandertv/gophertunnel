package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	_ = iota
	_
	_
	InteractActionLeaveVehicle
	InteractActionMouseOverEntity
	_
	InteractActionOpenInventory
)

// Interact is sent by the client when it interacts with another entity in some way. It used to be used for
// normal entity and block interaction, but this is no longer the case now.
type Interact struct {
	// Action type is the ID of the action that was executed by the player. It is one of the constants that
	// may be found above.
	ActionType byte
	// TargetEntityRuntimeID is the runtime ID of the entity that the player interacted with. This is empty
	// for the InteractActionOpenInventory action type.
	TargetEntityRuntimeID uint64
	// MouseOverPosition was the position relative to the entity moused over over which the player hovered
	// with its mouse/touch. It is only set if ActionType is InteractActionMouseOverEntity.
	MouseOverPosition mgl32.Vec3
}

// ID ...
func (*Interact) ID() uint32 {
	return IDInteract
}

// Marshal ...
func (pk *Interact) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.ActionType)
	_ = protocol.WriteVaruint64(buf, pk.TargetEntityRuntimeID)
	switch pk.ActionType {
	case InteractActionMouseOverEntity:
		_ = protocol.WriteVec3(buf, pk.MouseOverPosition)
	}
}

// Unmarshal ...
func (pk *Interact) Unmarshal(buf *bytes.Buffer) error {
	if err := chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.ActionType),
		protocol.Varuint64(buf, &pk.TargetEntityRuntimeID),
	); err != nil {
		return err
	}
	switch pk.ActionType {
	case InteractActionMouseOverEntity:
		return protocol.Vec3(buf, &pk.MouseOverPosition)
	}
	return nil
}
