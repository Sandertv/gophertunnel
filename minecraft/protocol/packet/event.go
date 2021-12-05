package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// TODO: Support the last seven new events.
const (
	EventTypeAchievementAwarded = iota
	EventTypeEntityInteract
	EventTypePortalBuilt
	EventTypePortalUsed
	EventTypeMobKilled
	EventTypeCauldronUsed
	EventTypePlayerDied
	EventTypeBossKilled
	EventTypeAgentCommand
	EventTypeAgentCreated // Unused for whatever reason?
	EventTypePatternRemoved
	EventTypeSlashCommandExecuted
	EventTypeFishBucketed
	EventTypeMobBorn
	EventTypePetDied
	EventTypeCauldronInteract
	EventTypeComposterInteract
	EventTypeBellUsed
	EventTypeEntityDefinitionTrigger
	EventTypeRaidUpdate
	EventTypeMovementAnomaly
	EventTypeMovementCorrected
	EventTypeExtractHoney
	EventTypeTargetBlockHit
	EventTypePiglinBarter
	EventTypePlayerWaxedOrUnwaxedCopper
	EventTypeCodeBuilderRuntimeAction
	EventTypeCodeBuilderScoreboard
	EventTypeStriderRiddenInLavaInOverworld
	EventTypeSneakCloseToSculkSensor
)

// Event is sent by the server to send an event with additional data. It is typically sent to the client for
// telemetry reasons, much like the SimpleEvent packet.
// TODO: Figure out what UsePlayerID is for.
type Event struct {
	// EntityRuntimeID is the runtime ID of the player. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// EventType is the type of the event to be called. It is one of the constants that may be found above.
	EventType int32
	// UsePlayerID ...
	UsePlayerID byte
	// EventData is the parsed event data.
	EventData protocol.EventData
}

// ID ...
func (*Event) ID() uint32 {
	return IDEvent
}

// Marshal ...
func (pk *Event) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Varint32(&pk.EventType)
	w.Uint8(&pk.UsePlayerID)

	pk.EventData.Marshal(w)
}

// Unmarshal ...
func (pk *Event) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.Varint32(&pk.EventType)
	r.Uint8(&pk.UsePlayerID)

	switch pk.EventType {
	case EventTypeAchievementAwarded:
		pk.EventData = &protocol.AchievementAwardedEventData{}
	case EventTypeEntityInteract:
		pk.EventData = &protocol.EntityInteractEventData{}
	case EventTypePortalBuilt:
		pk.EventData = &protocol.PortalBuiltEventData{}
	case EventTypePortalUsed:
		pk.EventData = &protocol.PortalUsedEventData{}
	case EventTypeMobKilled:
		pk.EventData = &protocol.MobKilledEventData{}
	case EventTypeCauldronUsed:
		pk.EventData = &protocol.CauldronUsedEventData{}
	case EventTypePlayerDied:
		pk.EventData = &protocol.PlayerDiedEventData{}
	case EventTypeBossKilled:
		pk.EventData = &protocol.BossKilledEventData{}
	case EventTypeAgentCommand:
		pk.EventData = &protocol.AgentCommandEventData{}
	case EventTypePatternRemoved:
		pk.EventData = &protocol.PatternRemovedEventData{}
	case EventTypeSlashCommandExecuted:
		pk.EventData = &protocol.SlashCommandExecutedEventData{}
	case EventTypeFishBucketed:
		pk.EventData = &protocol.FishBucketedEventData{}
	case EventTypeMobBorn:
		pk.EventData = &protocol.MobBornEventData{}
	case EventTypePetDied:
		pk.EventData = &protocol.PetDiedEventData{}
	case EventTypeCauldronInteract:
		pk.EventData = &protocol.CauldronInteractEventData{}
	case EventTypeComposterInteract:
		pk.EventData = &protocol.ComposterInteractEventData{}
	case EventTypeBellUsed:
		pk.EventData = &protocol.BellUsedEventData{}
	case EventTypeEntityDefinitionTrigger:
		pk.EventData = &protocol.EntityDefinitionTriggerEventData{}
	case EventTypeRaidUpdate:
		pk.EventData = &protocol.RaidUpdateEventData{}
	case EventTypeMovementAnomaly:
		pk.EventData = &protocol.MovementAnomalyEventData{}
	case EventTypeMovementCorrected:
		pk.EventData = &protocol.MovementCorrectedEventData{}
	case EventTypeExtractHoney:
		pk.EventData = &protocol.ExtractHoneyEventData{}
	}

	pk.EventData.Unmarshal(r)
}
