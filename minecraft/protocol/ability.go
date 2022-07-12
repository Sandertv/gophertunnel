package protocol

const (
	AbilityBuild = 1 << iota
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

const (
	AbilityLayerTypeCustomCache = iota
	AbilityLayerTypeBase
	AbilityLayerTypeSpectator
	AbilityLayerTypeCommands
)

const (
	AbilityBaseFlySpeed  = 0.05
	AbilityBaseWalkSpeed = 0.1
)

// AbilityLayer represents the abilities of a specific layer, such as the base layer or the spectator layer.
type AbilityLayer struct {
	// Type represents the type of the layer. This is one of the AbilityLayerType constants defined above.
	Type uint16
	// Abilities is a set of abilities that are enabled for the layer. This is one of the Ability constants defined
	// above.
	Abilities uint32
	// Values is a set of values that are associated with the enabled abilities, representing the values of the
	// abilities.
	Values uint32
	// FlySpeed is the default fly speed of the layer.
	FlySpeed float32
	// WalkSpeed is the default walk speed of the layer.
	WalkSpeed float32
}

// SerializedLayer reads/writes a AbilityLayer x using IO r.
func SerializedLayer(r IO, x *AbilityLayer) {
	r.Uint16(&x.Type)
	r.Uint32(&x.Abilities)
	r.Uint32(&x.Values)
	r.Float32(&x.FlySpeed)
	r.Float32(&x.WalkSpeed)
}
