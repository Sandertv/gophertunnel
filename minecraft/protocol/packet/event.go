package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	EventAchievementAwarded = iota
	EventEntityInteract
	EventPortalBuilt
	EventPortalUsed
	EventMobKilled
	EventCauldronUsed
	EventPlayerDeath
	EventBossKilled
	EventAgentCommand
	EventAgentCreated
	EventBannerPatternRemoved
	EventCommandExecuted
	EventFishBucketed
)

// Event is sent by the server to send an event with additional data. It is typically sent to the client for
// telemetry reasons, much like the SimpleEvent packet.
type Event struct {
	// EntityRuntimeID is the runtime ID of the player. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// EventType is the type of the event to be called. It is one of the constants that may be found above.
	EventType int32
	// Unknown ... TODO: Find out what this is for.
	Unknown byte
}

// ID ...
func (*Event) ID() uint32 {
	return IDEvent
}

// Marshal ...
func (pk *Event) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	_ = protocol.WriteVarint32(buf, pk.EventType)
	_ = binary.Write(buf, binary.LittleEndian, pk.Unknown)

	switch pk.EventType {
	// TODO: Figure out which events carry additional fields.
	}
}

// Unmarshal ...
func (pk *Event) Unmarshal(buf *bytes.Buffer) error {
	if err := chainErr(
		protocol.Varuint64(buf, &pk.EntityRuntimeID),
		protocol.Varint32(buf, &pk.EventType),
		binary.Read(buf, binary.LittleEndian, &pk.Unknown),
	); err != nil {
		return err
	}
	switch pk.EventType {
	// TODO: Figure out which events carry additional fields.
	}

	return nil
}
