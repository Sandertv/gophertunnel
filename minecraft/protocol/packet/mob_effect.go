package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	MobEffectAdd = iota + 1
	MobEffectModify
	MobEffectRemove
)

const (
	EffectSpeed = iota + 1
	EffectSlowness
	EffectHaste
	EffectMiningFatigue
	EffectStrength
	EffectInstantHealth
	EffectInstantDamage
	EffectJumpBoost
	EffectNausea
	EffectRegeneration
	EffectResistance
	EffectFireResistance
	EffectWaterBreathing
	EffectInvisibility
	EffectBlindness
	EffectNightVision
	EffectHunger
	EffectWeakness
	EffectPoison
	EffectWither
	EffectHealthBoost
	EffectAbsorption
	EffectSaturation
	EffectLevitation
	EffectFatalPoison
	EffectConduitPower
	EffectSlowFalling
)

// MobEffect is sent by the server to apply an effect to the player, for example an effect like poison. It may
// also be used to modify existing effects, or removing them completely.
type MobEffect struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// Operation is the operation of the packet. It is either MobEffectAdd, MobEffectModify or MobEffectRemove
	// and specifies the result of the packet client-side.
	Operation byte
	// EffectType is the ID of the effect to be added, removed or modified. It is one of the constants that
	// may be found above.
	EffectType int32
	// Amplifier is the amplifier of the effect. Take note that the amplifier is not the same as the effect's
	// level. The level is usually one higher than the amplifier, and the amplifier can actually be negative
	// to reverse the behaviour effect.
	Amplifier int32
	// Particles specifies if viewers of the entity that gets the effect shows particles around it. If set to
	// false, no particles are emitted around the entity.
	Particles bool
	// Duration is the duration of the effect in seconds. After the duration has elapsed, the effect will be
	// removed automatically client-side.
	Duration int32
	// Tick is the server tick at which the packet was sent. It is used in relation to CorrectPlayerMovePrediction.
	Tick uint64
}

// ID ...
func (*MobEffect) ID() uint32 {
	return IDMobEffect
}

func (pk *MobEffect) Marshal(io protocol.IO) {
	io.Varuint64(&pk.EntityRuntimeID)
	io.Uint8(&pk.Operation)
	io.Varint32(&pk.EffectType)
	io.Varint32(&pk.Amplifier)
	io.Bool(&pk.Particles)
	io.Varint32(&pk.Duration)
	io.Varuint64(&pk.Tick)
}
