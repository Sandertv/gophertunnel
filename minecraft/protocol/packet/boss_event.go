package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	BossEventShow = iota
	BossEventRegisterPlayer
	BossEventHide
	BossEventUnregisterPlayer
	BossEventHealthPercentage
	BossEventTitle
	BossEventAppearanceProperties
	BossEventTexture
	BossEventRequest
)

const (
	BossEventColourGrey = iota
	BossEventColourBlue
	BossEventColourRed
	BossEventColourGreen
	BossEventColourYellow
	BossEventColourPurple
	BossEventColourWhite
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
	// ScreenDarkening currently seems not to do anything.
	ScreenDarkening int16
	// Colour is the colour of the boss bar that is shown when a player is subscribed. It is only set if the
	// EventType is BossEventShow, BossEventAppearanceProperties or BossEventTexture. This is functional as
	// of 1.18 and can be any of the BossEventColour constants listed above.
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
func (pk *BossEvent) Marshal(w *protocol.Writer) {
	w.Varint64(&pk.BossEntityUniqueID)
	w.Varuint32(&pk.EventType)
	switch pk.EventType {
	case BossEventShow:
		w.String(&pk.BossBarTitle)
		w.Float32(&pk.HealthPercentage)
		w.Int16(&pk.ScreenDarkening)
		w.Varuint32(&pk.Colour)
		w.Varuint32(&pk.Overlay)
	case BossEventRegisterPlayer, BossEventUnregisterPlayer, BossEventRequest:
		w.Varint64(&pk.PlayerUniqueID)
	case BossEventHide:
		// No extra payload for this boss event type.
	case BossEventHealthPercentage:
		w.Float32(&pk.HealthPercentage)
	case BossEventTitle:
		w.String(&pk.BossBarTitle)
	case BossEventAppearanceProperties:
		w.Int16(&pk.ScreenDarkening)
		w.Varuint32(&pk.Colour)
		w.Varuint32(&pk.Overlay)
	case BossEventTexture:
		w.Varuint32(&pk.Colour)
		w.Varuint32(&pk.Overlay)
	default:
		w.UnknownEnumOption(pk.EventType, "boss event type")
	}
}

// Unmarshal ...
func (pk *BossEvent) Unmarshal(r *protocol.Reader) {
	r.Varint64(&pk.BossEntityUniqueID)
	r.Varuint32(&pk.EventType)
	switch pk.EventType {
	case BossEventShow:
		r.String(&pk.BossBarTitle)
		r.Float32(&pk.HealthPercentage)
		r.Int16(&pk.ScreenDarkening)
		r.Varuint32(&pk.Colour)
		r.Varuint32(&pk.Overlay)
	case BossEventRegisterPlayer, BossEventUnregisterPlayer, BossEventRequest:
		r.Varint64(&pk.PlayerUniqueID)
	case BossEventHide:
		// No extra payload for this boss event type.
	case BossEventHealthPercentage:
		r.Float32(&pk.HealthPercentage)
	case BossEventTitle:
		r.String(&pk.BossBarTitle)
	case BossEventAppearanceProperties:
		r.Int16(&pk.ScreenDarkening)
		r.Varuint32(&pk.Colour)
		r.Varuint32(&pk.Overlay)
	case BossEventTexture:
		r.Varuint32(&pk.Colour)
		r.Varuint32(&pk.Overlay)
	default:
		r.UnknownEnumOption(pk.EventType, "boss event type")
	}
}
