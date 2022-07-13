package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ActorEventJump = iota + 1
	ActorEventHurt
	ActorEventDeath
	ActorEventStartAttacking
	ActorEventStopAttacking
	ActorEventTamingFailed
	ActorEventTamingSucceeded
	ActorEventShakeWetness
	ActorEventUseItem
	ActorEventEatGrass
	ActorEventFishhookBubble
	ActorEventFishhookFishPosition
	ActorEventFishhookHookTime
	ActorEventFishhookTease
	ActorEventSquidFleeing
	ActorEventZombieConverting
	ActorEventPlayAmbient
	ActorEventSpawnAlive
	ActorEventStartOfferFlower
	ActorEventStopOfferFlower
	ActorEventLoveHearts
	ActorEventVillagerAngry
	ActorEventVillagerHappy
	ActorEventWitchHatMagic
	ActorEventFireworksExplode
	ActorEventInLoveHearts
	ActorEventSilverfishMergeAnimation
	ActorEventGuardianAttackSound
	ActorEventDrinkPotion
	ActorEventThrowPotion
	ActorEventCartWithPrimeTNT
	ActorEventPrimeCreeper
	ActorEventAirSupply
	ActorEventAddPlayerLevels
	ActorEventGuardianMiningFatigue
	ActorEventAgentSwingArm
	ActorEventDragonStartDeathAnim
	ActorEventGroundDust
	ActorEventShake
)

const (
	ActorEventFeed = iota + 57
	_
	_
	ActorEventBabyEat
	ActorEventInstantDeath
	ActorEventNotifyTrade
	ActorEventLeashDestroyed
	ActorEventCaravanUpdated
	ActorEventTalismanActivate
	ActorEventUpdateStructureFeature
	ActorEventPlayerSpawnedMob
	ActorEventPuke
	ActorEventUpdateStackSize
	ActorEventStartSwimming
	ActorEventBalloonPop
	ActorEventTreasureHunt
	ActorEventSummonAgent
	ActorEventFinishedChargingCrossbow
	ActorEventLandedOnGround
	ActorEventActorGrowUp
	ActorEventVibrationDetected
	ActorEventDrinkMilk
)

// ActorEvent is sent by the server when a particular event happens that has to do with an entity. Some of
// these events are entity-specific, for example a wolf shaking itself dry, but others are used for each
// entity, such as dying.
type ActorEvent struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// EventType is the ID of the event to be called. It is one of the constants that can be found above.
	EventType byte
	// EventData is optional data associated with a particular event. The data has a different function for
	// different events, however most events don't use this field at all.
	EventData int32
}

// ID ...
func (*ActorEvent) ID() uint32 {
	return IDActorEvent
}

// Marshal ...
func (pk *ActorEvent) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Uint8(&pk.EventType)
	w.Varint32(&pk.EventData)
}

// Unmarshal ...
func (pk *ActorEvent) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.Uint8(&pk.EventType)
	r.Varint32(&pk.EventData)
}
