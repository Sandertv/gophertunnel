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
}

// ID ...
func (*MobEffect) ID() uint32 {
	return IDMobEffect
}

// Marshal ...
func (pk *MobEffect) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Uint8(&pk.Operation)
	w.Varint32(&pk.EffectType)
	w.Varint32(&pk.Amplifier)
	w.Bool(&pk.Particles)
	w.Varint32(&pk.Duration)
}

// Unmarshal ...
func (pk *MobEffect) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.Uint8(&pk.Operation)
	r.Varint32(&pk.EffectType)
	r.Varint32(&pk.Amplifier)
	r.Bool(&pk.Particles)
	r.Varint32(&pk.Duration)
}
