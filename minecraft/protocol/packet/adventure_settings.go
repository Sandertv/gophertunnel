package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	AdventureFlagWorldImmutable = 1 << iota
	AdventureSettingsFlagsNoPvM
	AdventureSettingsFlagsNoMvP
	AdventureSettingsFlagsUnused
	AdventureSettingsFlagsShowNameTags
	AdventureFlagAutoJump
	AdventureFlagAllowFlight
	AdventureFlagNoClip
	AdventureFlagWorldBuilder
	AdventureFlagFlying
	AdventureFlagMuted
)

const (
	CommandPermissionLevelNormal = iota
	CommandPermissionLevelGameDirectors
	CommandPermissionLevelAdmin
	CommandPermissionLevelHost
	CommandPermissionLevelOwner
	CommandPermissionLevelInternal
)

const (
	ActionPermissionMine = 1 << iota
	ActionPermissionDoorsAndSwitches
	ActionPermissionOpenContainers
	ActionPermissionAttackPlayers
	ActionPermissionAttackMobs
	ActionPermissionOperator
	ActionPermissionTeleport
	ActionPermissionBuild
	ActionPermissionDefault
)

const (
	PermissionLevelVisitor = iota
	PermissionLevelMember
	PermissionLevelOperator
	PermissionLevelCustom
)

// AdventureSettings is sent by the server to update game-play related features, in particular permissions to
// access these features for the client. It includes allowing the player to fly, build and mine, and attack
// entities. Most of these flags should be checked server-side instead of using this packet only.
// The client may also send this packet to the server when it updates one of these settings through the
// in-game settings interface. The server should verify if the player actually has permission to update those
// settings.
type AdventureSettings struct {
	// Flags is a set of flags that specify certain properties of the player, such as whether or not it can
	// fly and/or move through blocks. It is one of the AdventureFlag constants above.
	Flags uint32
	// CommandPermissionLevel is a permission level that specifies the kind of commands that the player is
	// allowed to use. It is one of the CommandPermissionLevel constants above.
	CommandPermissionLevel uint32
	// ActionPermissions is, much like Flags, a set of flags that specify actions that the player is allowed
	// to undertake, such as whether it is allowed to edit blocks, open doors etc. It is a combination of the
	// ActionPermission constants above.
	ActionPermissions uint32
	// PermissionLevel is the permission level of the player as it shows up in the player list built up using
	// the PlayerList packet. It is one of the PermissionLevel constants above.
	PermissionLevel uint32
	// CustomStoredPermissions ...
	CustomStoredPermissions uint32
	// PlayerUniqueID is a unique identifier of the player. This must be filled out with the entity unique ID of the
	// player.
	PlayerUniqueID int64
}

// ID ...
func (*AdventureSettings) ID() uint32 {
	return IDAdventureSettings
}

// Marshal ...
func (pk *AdventureSettings) Marshal(w *protocol.Writer) {
	w.Varuint32(&pk.Flags)
	w.Varuint32(&pk.CommandPermissionLevel)
	w.Varuint32(&pk.ActionPermissions)
	w.Varuint32(&pk.PermissionLevel)
	w.Varuint32(&pk.CustomStoredPermissions)
	w.Int64(&pk.PlayerUniqueID)
}

// Unmarshal ...
func (pk *AdventureSettings) Unmarshal(r *protocol.Reader) {
	r.Varuint32(&pk.Flags)
	r.Varuint32(&pk.CommandPermissionLevel)
	r.Varuint32(&pk.ActionPermissions)
	r.Varuint32(&pk.PermissionLevel)
	r.Varuint32(&pk.CustomStoredPermissions)
	r.Int64(&pk.PlayerUniqueID)
}
