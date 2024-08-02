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
	AbilityPrivilegedBuilder
	AbilityCount
)

const (
	AbilityLayerTypeCustomCache = iota
	AbilityLayerTypeBase
	AbilityLayerTypeSpectator
	AbilityLayerTypeCommands
	AbilityLayerTypeEditor
	AbilityLayerTypeLoadingScreen
)

const (
	AbilityBaseFlySpeed  = 0.05
	AbilityBaseWalkSpeed = 0.1
)

// AbilityData represents various data about the abilities of a player, such as ability layers or permissions.
type AbilityData struct {
	// EntityUniqueID is a unique identifier of the player. It appears it is not required to fill this field
	// out with a correct value. Simply writing 0 seems to work.
	EntityUniqueID int64
	// PlayerPermissions is the permission level of the player as it shows up in the player list built up using
	// the PlayerList packet.
	PlayerPermissions byte
	// CommandPermissions is a set of permissions that specify what commands a player is allowed to execute.
	CommandPermissions byte
	// Layers contains all ability layers and their potential values. This should at least have one entry, being the
	// base layer.
	Layers []AbilityLayer
}

// Marshal encodes/decodes an AbilityData.
func (x *AbilityData) Marshal(r IO) {
	r.Int64(&x.EntityUniqueID)
	r.Uint8(&x.PlayerPermissions)
	r.Uint8(&x.CommandPermissions)
	SliceUint8Length(r, &x.Layers)
}

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

// Marshal encodes/decodes an AbilityLayer.
func (x *AbilityLayer) Marshal(r IO) {
	r.Uint16(&x.Type)
	r.Uint32(&x.Abilities)
	r.Uint32(&x.Values)
	r.Float32(&x.FlySpeed)
	r.Float32(&x.WalkSpeed)
}
