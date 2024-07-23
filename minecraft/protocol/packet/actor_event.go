package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ActorEventNone                     = 0
	ActorEventJump                     = 1
	ActorEventHurt                     = 2
	ActorEventDeath                    = 3
	ActorEventStartAttacking           = 4
	ActorEventStopAttacking            = 5
	ActorEventTamingFailed             = 6
	ActorEventTamingSucceeded          = 7
	ActorEventShakeWetness             = 8
	ActorEventEatGrass                 = 10
	ActorEventFishhookBubble           = 11
	ActorEventFishhookFishPosition     = 12
	ActorEventFishhookHookTime         = 13
	ActorEventFishhookTease            = 14
	ActorEventSquidFleeing             = 15
	ActorEventZombieConverting         = 16
	ActorEventPlayAmbient              = 17
	ActorEventSpawnAlive               = 18
	ActorEventStartOfferFlower         = 19
	ActorEventStopOfferFlower          = 20
	ActorEventLoveHearts               = 21
	ActorEventVillagerAngry            = 22
	ActorEventVillagerHappy            = 23
	ActorEventWitchHatMagic            = 24
	ActorEventFireworksExplode         = 25
	ActorEventInLoveHearts             = 26
	ActorEventSilverfishMergeAnimation = 27
	ActorEventGuardianAttackSound      = 28
	ActorEventDrinkPotion              = 29
	ActorEventThrowPotion              = 30
	ActorEventCartWithPrimeTNT         = 31
	ActorEventPrimeCreeper             = 32
	ActorEventAirSupply                = 33
	ActorEventAddPlayerLevels          = 34
	ActorEventGuardianMiningFatigue    = 35
	ActorEventAgentSwingArm            = 36
	ActorEventDragonStartDeathAnim     = 37
	ActorEventGroundDust               = 38
	ActorEventShake                    = 39
	ActorEventFeed                     = 57
	ActorEventBabyAge                  = 60
	ActorEventInstantDeath             = 61
	ActorEventNotifyTrade              = 62
	ActorEventLeashDestroyed           = 63
	ActorEventCaravanUpdated           = 64
	ActorEventTalismanActivate         = 65
	ActorEventUpdateStructureFeature   = 66
	ActorEventPlayerSpawnedMob         = 67
	ActorEventPuke                     = 68
	ActorEventUpdateStackSize          = 69
	ActorEventStartSwimming            = 70
	ActorEventBalloonPop               = 71
	ActorEventTreasureHunt             = 72
	ActorEventSummonAgent              = 73
	ActorEventFinishedChargingItem     = 74
	ActorEventActorGrowUp              = 76
	ActorEventVibrationDetected        = 77
	ActorEventDrinkMilk                = 78
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

func (pk *ActorEvent) Marshal(io protocol.IO) {
	io.Varuint64(&pk.EntityRuntimeID)
	io.Uint8(&pk.EventType)
	io.Varint32(&pk.EventData)
}
