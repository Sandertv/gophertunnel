package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	AbilityBuild = iota
	AbilityMine
	AbilityDoorsAndSwitches
	AbilityOpenContainers
	AbilityAttackPlayers
	AbilityAttackMobs
	AbilityOperatorCommands
	AbilityTeleport
	AbilityInvulnerable
	AbilityFlying
	AbilityMayFly
	AbilityInstantBuild
	AbilityLightning
	AbilityFlySpeed
	AbilityWalkSpeed
	AbilityMuted
	AbilityWorldBuilder
	AbilityNoClip
	AbilityCount
)

// RequestAbility is a packet sent by the client to the server to request permission for a specific ability from the
// server. These abilities are defined above.
type RequestAbility struct {
	// Ability is the ability that the client is requesting. This is one of the constants defined in the
	// protocol/ability.go file.
	Ability int32
	// Value represents the value of the ability. This can either be a boolean or a float32, otherwise the writer/reader
	// will panic.
	Value any
}

// ID ...
func (*RequestAbility) ID() uint32 {
	return IDRequestAbility
}

// Marshal ...
func (pk *RequestAbility) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.Ability)
	switch val := pk.Value.(type) {
	case bool:
		valType, defaultVal := uint8(1), float32(0)
		w.Uint8(&valType)
		w.Bool(&val)
		w.Float32(&defaultVal)
	case float32:
		valType, defaultVal := uint8(2), false
		w.Uint8(&valType)
		w.Bool(&defaultVal)
		w.Float32(&val)
	default:
		w.InvalidValue(pk.Value, "ability value type", "must be bool or float32")
	}
}

// Unmarshal ...
func (pk *RequestAbility) Unmarshal(r *protocol.Reader) {
	valType, boolVal, floatVal := uint8(0), false, float32(0)
	r.Varint32(&pk.Ability)
	r.Uint8(&valType)
	r.Bool(&boolVal)
	r.Float32(&floatVal)
	switch valType {
	case 1:
		pk.Value = boolVal
	case 2:
		pk.Value = floatVal
	default:
		r.InvalidValue(valType, "ability value type", "must be bool or float32")
	}
}
