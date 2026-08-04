package protocol

import (
	"image/color"

	"github.com/google/uuid"
)

const (
	PlayerListActionAdd = iota
	PlayerListActionRemove
)

const (
	PlayerActionStartBreak = iota
	PlayerActionAbortBreak
	PlayerActionStopBreak
	PlayerActionGetUpdatedBlock
	PlayerActionDropItem
	PlayerActionStartSleeping
	PlayerActionStopSleeping
	PlayerActionRespawn
	PlayerActionJump
	PlayerActionStartSprint
	PlayerActionStopSprint
	PlayerActionStartSneak
	PlayerActionStopSneak
	PlayerActionCreativePlayerDestroyBlock
	PlayerActionDimensionChangeDone
	PlayerActionStartGlide
	PlayerActionStopGlide
	PlayerActionBuildDenied
	PlayerActionCrackBreak
	PlayerActionChangeSkin
	PlayerActionSetEnchantmentSeed
	PlayerActionStartSwimming
	PlayerActionStopSwimming
	PlayerActionStartSpinAttack
	PlayerActionStopSpinAttack
	PlayerActionStartBuildingBlock
	PlayerActionPredictDestroyBlock
	PlayerActionContinueDestroyBlock
	PlayerActionStartItemUseOn
	PlayerActionStopItemUseOn
	PlayerActionHandledTeleport
	PlayerActionMissedSwing
	PlayerActionStartCrawling
	PlayerActionStopCrawling
	PlayerActionStartFlying
	PlayerActionStopFlying
	PlayerActionReceivedServerData
	PlayerActionStartUsingItem
	PlayerActionInternalUpdate
	PlayerActionCount
)

// PlayerListEntry is an entry found in the PlayerList packet. It represents a single player using the UUID
// found in the entry, and contains several properties such as the skin.
type PlayerListEntry struct {
	// ActionType is the action to execute upon the player list.
	ActionType byte
	// UUID is the UUID of the player as sent in the Login packet when the client joined the server. It must
	// match this UUID exactly for the correct XBOX Live icon to show up in the list.
	UUID uuid.UUID
	// EntityUniqueID is the unique entity ID of the player. This ID typically stays consistent during the
	// lifetime of a world, but servers often send the runtime ID for this.
	EntityUniqueID int64
	// Username is the username that is shown in the player list of the player that obtains a PlayerList
	// packet with this entry. It does not have to be the same as the actual username of the player.
	Username string
	// XUID is the XBOX Live user ID of the player, which will remain consistent as long as the player is
	// logged in with the XBOX Live account.
	XUID string
	// PlatformChatID is an identifier only set for particular platforms when chatting (presumably only for
	// Nintendo Switch). It is otherwise an empty string, and is used to decide which players are able to
	// chat with each other.
	PlatformChatID string
	// BuildPlatform is the platform of the player as sent by that player in the Login packet.
	BuildPlatform int32
	// Skin is the skin of the player that should be added to the player list. Once sent here, it will not
	// have to be sent again.
	Skin Skin
	// Teacher is a Minecraft: Education Edition field. It specifies if the player to be added to the player
	// list is a teacher.
	Teacher bool
	// Host specifies if the player that is added to the player list is the host of the game.
	Host bool
	// SubClient specifies if the player that is added to the player list is a sub-client of another player.
	SubClient bool
	// PlayerColour is the colour of the player that is shown in UI elements, currently only used for the
	// player locator bar.
	PlayerColour color.RGBA
}

// Marshal encodes/decodes a PlayerListEntry.
func (x *PlayerListEntry) Marshal(r IO) {
	variant := uint32(0)
	if x.ActionType == PlayerListActionAdd {
		variant = 1
	}
	r.Varuint32(&variant)

	legacyAction := x.ActionType
	r.Uint8(&legacyAction)
	x.ActionType = PlayerListActionRemove
	if variant == 1 {
		x.ActionType = PlayerListActionAdd
	} else if variant != 0 {
		r.UnknownEnumOption(variant, "player list entry variant")
	}
	if legacyAction != x.ActionType {
		r.InvalidValue(legacyAction, "player list action type", "does not match entry variant")
	}
	r.UUID(&x.UUID)
	if x.ActionType == PlayerListActionRemove {
		return
	}

	r.Varint64(&x.EntityUniqueID)
	r.String(&x.Username)
	r.String(&x.XUID)
	r.String(&x.PlatformChatID)
	r.Int32(&x.BuildPlatform)
	Single(r, &x.Skin)
	r.Bool(&x.Teacher)
	r.Bool(&x.Host)
	r.Bool(&x.SubClient)
	r.BEARGB(&x.PlayerColour)
}

// PlayerMovementSettings represents the different server authoritative movement settings. These control how
// the client will provide input to the server.
type PlayerMovementSettings struct {
	// RewindHistorySize is the amount of history to keep at maximum.
	RewindHistorySize int32
	// ServerAuthoritativeBlockBreaking specifies if block breaking should be sent through
	// packet.PlayerAuthInput or not.
	ServerAuthoritativeBlockBreaking bool
}

// PlayerMoveSettings reads/writes PlayerMovementSettings x to/from IO r.
func PlayerMoveSettings(r IO, x *PlayerMovementSettings) {
	r.Varint32(&x.RewindHistorySize)
	r.Bool(&x.ServerAuthoritativeBlockBreaking)
}

// PlayerBlockAction ...
type PlayerBlockAction struct {
	// Action is the action to be performed, and is one of the constants listed above.
	Action int32
	// BlockPos is the position of the block that was interacted with.
	BlockPos BlockPos
	// Face is the face of the block that was interacted with.
	Face int32
}

// Marshal encodes/decodes a PlayerBlockAction.
func (x *PlayerBlockAction) Marshal(r IO) {
	r.Varint32(&x.Action)
	r.BlockPos(&x.BlockPos)
	r.Varint32(&x.Face)
}

// PlayerArmourDamageEntry represents an entry for a single piece of armour that should be damaged.
type PlayerArmourDamageEntry struct {
	// ArmourSlot is the index of the armour slot to damage.
	ArmourSlot int32
	// Damage is the amount of damage to apply to the armour in the specified slot.
	Damage int16
}

// Marshal encodes/decodes a PlayerArmourDamageEntry.
func (x *PlayerArmourDamageEntry) Marshal(r IO) {
	r.Varint32(&x.ArmourSlot)
	r.Int16(&x.Damage)
}

type TeleportData struct {
	// TeleportCause specifies why the teleport occurred. See the TeleportCause constants in the packet package.
	// TeleportData is present only when MovePlayer uses teleport mode.
	TeleportCause int32
	// TeleportSourceEntityType is the entity type that caused the teleportation, for example an ender pearl.
	TeleportSourceEntityType int32
}

func (x *TeleportData) Marshal(r IO) {
	r.Int32(&x.TeleportCause)
	r.Int32(&x.TeleportSourceEntityType)
}
