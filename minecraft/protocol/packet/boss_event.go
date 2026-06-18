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
	BossEventColourPink = iota
	BossEventColourBlue
	BossEventColourRed
	BossEventColourGreen
	BossEventColourYellow
	BossEventColourPurple
	BossEventColourRebeccaPurple
	BossEventColourWhite
)

const (
	BossEventOverlayProgress = iota
	BossEventOverlayNotched6
	BossEventOverlayNotched10
	BossEventOverlayNotched12
	BossEventOverlayNotched20
)

// BossEvent is sent by the server to make a specific 'boss event' occur in the
// world. It includes features such as showing a boss bar to the player and
// turning the sky dark.
type BossEvent struct {
	// BossEntityUniqueID is the unique ID of the boss entity that the boss
	// event sent involves. By default, the health percentage and title of the
	// boss bar depend on the health and name tag of this entity. If
	// BossEntityUniqueID is the same as the client's entity unique ID, its
	// HealthPercentage and BossBarTitle can be freely altered.
	BossEntityUniqueID int64
	// PlayerUniqueID is the unique ID of the player that is registered to or
	// unregistered from the boss fight.
	PlayerUniqueID int64
	// EventType is the type of the event. It is one of the BossEvent constants above.
	EventType uint8
	// BossBarTitle is the title shown above the boss bar. It may be set to set
	// a different title if the BossEntityUniqueID matches the client's entity
	// unique ID.
	BossBarTitle string
	// FilteredBossBarTitle is a filtered version of BossBarTitle with all the
	// profanity removed. The client will use this over BossBarTitle if this
	// field is not empty and they have the "Filter Profanity" setting enabled.
	FilteredBossBarTitle string
	// HealthPercentage is the percentage of health that is shown in the boss
	// bar (0.0-1.0). The HealthPercentage may be set to a specific value if the
	// BossEntityUniqueID matches the client's entity unique ID.
	HealthPercentage float32
	// Colour is the colour of the boss bar that is shown when a player is
	// subscribed. It is one of the BossEventColour constants listed above.
	Colour uint8
	// Overlay is the overlay of the boss bar that is shown on top of the boss
	// bar when a player is subscribed. It is one of the BossEventOverlay
	// constants listed above.
	Overlay uint8
}

// ID ...
func (*BossEvent) ID() uint32 {
	return IDBossEvent
}

func (pk *BossEvent) Marshal(io protocol.IO) {
	io.Varint64(&pk.BossEntityUniqueID)
	io.Varint64(&pk.PlayerUniqueID)
	io.Uint8(&pk.EventType)
	io.String(&pk.BossBarTitle)
	io.String(&pk.FilteredBossBarTitle)
	io.Float32(&pk.HealthPercentage)
	io.Uint8(&pk.Colour)
	io.Uint8(&pk.Overlay)
}
