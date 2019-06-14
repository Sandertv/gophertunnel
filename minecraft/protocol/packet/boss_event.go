package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	BossEventShow = iota
	// BossEventRegisterPlayer is sent by the client to the server to request being shown the boss bar.
	BossEventRegisterPlayer
	BossEventHide
	// BossEventUnregisterPlayer is sent by the client to request the removal of the boss bar.
	BossEventUnregisterPlayer
	BossEventHealthPercentage
	BossEventTitle
	BossEventAppearanceProperties
	BossEventTexture
)

// BossEvent is sent by the server to make a specific 'boss event' occur in the world. It includes features
// such as showing a boss bar to the player and turning the sky dark.
type BossEvent struct {
	// BossEntityUniqueID is the unique ID of the boss entity that the boss event sent involves. The health
	// percentage and title of the boss bar depend on the health and name tag of this entity.
	BossEntityUniqueID int64
	// EventType is the type of the event. The fields written depend on the event type set, and some event
	// types are sent by the client, whereas others are sent by the server. The event type is one of the
	// constants above.
	EventType uint32
	// PlayerUniqueID is the unique ID of the player that is registered to or unregistered from the boss
	// fight. It is set if EventType is either BossEventRegisterPlayer or BossEventUnregisterPlayer.
	PlayerUniqueID int64
	// BossBarTitle is the title shown above the boss bar. It currently does not function, and instead uses
	// the name tag of the boss entity at all times. It is only set if the EventType is BossEventShow or
	// BossEventTitle.
	BossBarTitle string
	// HealthPercentage is the percentage of health that is shown in the boss bar. It currently does not
	// function, and instead uses the health percentage of the boss entity at all times. It is only set if the
	// EventType is BossEventShow or BossEventHealthPercentage.
	HealthPercentage float32
	// UnknownInt16: Might be something to do with the sky darkening...
	UnknownInt16 int16
	// Colour is the colour of the boss bar that is shown when a player is subscribed. It currently does not
	// function. It is only set if the EventType is BossEventShow, BossEventAppearanceProperties or
	// BossEventTexture.
	Colour uint32
	// Overlay is the overlay of the boss bar that is shown on top of the boss bar when a player is
	// subscribed. It currently does not function. It is only set if the EventType is BossEventShow,
	// BossEventAppearanceProperties or BossEventTexture.
	Overlay uint32
}

// ID ...
func (*BossEvent) ID() uint32 {
	return IDBossEvent
}

// Marshal ...
func (pk *BossEvent) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint64(buf, pk.BossEntityUniqueID)
	_ = protocol.WriteVaruint32(buf, pk.EventType)
	switch pk.EventType {
	case BossEventShow:
		_ = protocol.WriteString(buf, pk.BossBarTitle)
		_ = protocol.WriteFloat32(buf, pk.HealthPercentage)
		_ = binary.Write(buf, binary.LittleEndian, pk.UnknownInt16)
		_ = protocol.WriteVaruint32(buf, pk.Colour)
		_ = protocol.WriteVaruint32(buf, pk.Overlay)
	case BossEventRegisterPlayer, BossEventUnregisterPlayer:
		_ = protocol.WriteVarint64(buf, pk.PlayerUniqueID)
	case BossEventHide:
		// No extra payload for this event type.
	case BossEventHealthPercentage:
		_ = protocol.WriteFloat32(buf, pk.HealthPercentage)
	case BossEventTitle:
		_ = protocol.WriteString(buf, pk.BossBarTitle)
	case BossEventAppearanceProperties:
		_ = binary.Write(buf, binary.LittleEndian, pk.UnknownInt16)
		_ = protocol.WriteVaruint32(buf, pk.Colour)
		_ = protocol.WriteVaruint32(buf, pk.Overlay)
	case BossEventTexture:
		_ = protocol.WriteVaruint32(buf, pk.Colour)
		_ = protocol.WriteVaruint32(buf, pk.Overlay)
	default:
		panic(fmt.Sprintf("invalid boss event type %v", pk.EventType))
	}
}

// Unmarshal ...
func (pk *BossEvent) Unmarshal(buf *bytes.Buffer) error {
	if err := chainErr(
		protocol.Varint64(buf, &pk.BossEntityUniqueID),
		protocol.Varuint32(buf, &pk.EventType),
	); err != nil {
		return err
	}
	switch pk.EventType {
	case BossEventShow:
		return chainErr(
			protocol.String(buf, &pk.BossBarTitle),
			protocol.Float32(buf, &pk.HealthPercentage),
			binary.Read(buf, binary.LittleEndian, &pk.UnknownInt16),
			protocol.Varuint32(buf, &pk.Colour),
			protocol.Varuint32(buf, &pk.Overlay),
		)
	case BossEventRegisterPlayer, BossEventUnregisterPlayer:
		return protocol.Varint64(buf, &pk.PlayerUniqueID)
	case BossEventHide:
		// No extra payload for this boss event type.
		return nil
	case BossEventHealthPercentage:
		return protocol.Float32(buf, &pk.HealthPercentage)
	case BossEventTitle:
		return protocol.String(buf, &pk.BossBarTitle)
	case BossEventAppearanceProperties:
		return chainErr(
			binary.Read(buf, binary.LittleEndian, &pk.UnknownInt16),
			protocol.Varuint32(buf, &pk.Colour),
			protocol.Varuint32(buf, &pk.Overlay),
		)
	case BossEventTexture:
		return chainErr(
			protocol.Varuint32(buf, &pk.Colour),
			protocol.Varuint32(buf, &pk.Overlay),
		)
	default:
		return fmt.Errorf("unknown boss event type %v", pk.EventType)
	}
}
