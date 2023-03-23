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

func (pk *RequestAbility) Marshal(io protocol.IO) {
	io.Varint32(&pk.Ability)
	io.AbilityValue(&pk.Value)
}
