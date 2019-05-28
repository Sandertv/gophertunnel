package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	_ = iota
	_
	EntityEventHurt
	EntityEventDeath
	EntityEventArmSwing
	_
	EntityEventTameFail
	EntityEventTameSucceed
	EntityEventShakeDry
	EntityEventUseItem
	EntityEventEatGrass
	EntityEventFishHookBubble
	EntityEventFishHookPosition
	EntityEventFishHook
	EntityEventFishHookTease
	EntityEventSquidInkCloud
	EntityEventZombieVillagerCure
	_
	EntityEventRespawn
	EntityEventIronGolemOfferFlower
	EntityEventIronGolemWithdrawFlower
	EntityEventLookingForPartner
	_
	_
	EntityEventWitchSpell
	EntityEventFirework
	_
	EntityEventSilverfishSpawn
	_
	EntityEventWitchDrinkPotion
	EntityEventWitchThrowPotion
	EntityEventMinecartTNTPrimeFuse
	_
	_
	EntityEventPlayerAddXPLevels
	EntityEventElderGuardianCurse
	EntityEventAgentArmSwing
	EntityEventEnderDragonDeath
	EntityEventDustParticles
	EntityEventArrowShake
	// ...
	EntityEventEatingItem = 57
)

const (
	EntityEventBabyAnimalFeed = iota + 60
	EntityEventDeathSmokeCloud
	EntityEventCompleteTrade
	EntityEventRemoveLeash
	_
	EntityEventConsumeToken
	EntityEventPlayerCheckTreasureHunterAchievement
	EntityEventEntitySpawn
	EntityEventDragonBreath
	EntityEventItemEntityMerge
)

// EntityEvent is sent by the server when a particular event happens that has to do with an entity. Some of
// these events are entity-specific, for example a wolf shaking itself dry, but others are used for each
// entity, such as dying.
type EntityEvent struct {
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
func (*EntityEvent) ID() uint32 {
	return IDEntityEvent
}

// Marshal ...
func (pk *EntityEvent) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	_ = binary.Write(buf, binary.LittleEndian, pk.EventType)
	_ = protocol.WriteVarint32(buf, pk.EventData)
}

// Unmarshal ...
func (pk *EntityEvent) Unmarshal(buf *bytes.Buffer) error {
	return ChainErr(
		protocol.Varuint64(buf, &pk.EntityRuntimeID),
		binary.Read(buf, binary.LittleEndian, &pk.EventType),
		protocol.Varint32(buf, &pk.EventData),
	)
}
