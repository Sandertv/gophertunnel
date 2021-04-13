package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AddPlayer is sent by the server to the client to make a player entity show up client-side. It is one of the
// few entities that cannot be sent using the AddActor packet.
type AddPlayer struct {
	// UUID is the UUID of the player. It is the same UUID that the client sent in the Login packet at the
	// start of the session. A player with this UUID must exist in the player list (built up using the
	// PlayerList packet), for it to show up in-game.
	UUID uuid.UUID
	// Username is the name of the player. This username is the username that will be set as the initial
	// name tag of the player.
	Username string
	// EntityUniqueID is the unique ID of the player. The unique ID is a value that remains consistent across
	// different sessions of the same world, but most servers simply fill the runtime ID of the player out for
	// this field.
	EntityUniqueID int64
	// EntityRuntimeID is the runtime ID of the player. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// PlatformChatID is an identifier only set for particular platforms when chatting (presumably only for
	// Nintendo Switch). It is otherwise an empty string, and is used to decide which players are able to
	// chat with each other.
	PlatformChatID string
	// Position is the position to spawn the player on. If the player is on a distance that the viewer cannot
	// see it, the player will still show up if the viewer moves closer.
	Position mgl32.Vec3
	// Velocity is the initial velocity the player spawns with. This velocity will initiate client side
	// movement of the player.
	Velocity mgl32.Vec3
	// Pitch is the vertical rotation of the player. Facing straight forward yields a pitch of 0. Pitch is
	// measured in degrees.
	Pitch float32
	// Yaw is the horizontal rotation of the player. Yaw is also measured in degrees.
	Yaw float32
	// HeadYaw is the same as Yaw, except that it applies specifically to the head of the player. A different
	// value for HeadYaw than Yaw means that the player will have its head turned.
	HeadYaw float32
	// HeldItem is the item that the player is holding. The item is shown to the viewer as soon as the player
	// itself shows up. Needless to say that this field is rather pointless, as additional packets still must
	// be sent for armour to show up.
	HeldItem protocol.ItemInstance
	// EntityMetadata is a map of entity metadata, which includes flags and data properties that alter in
	// particular the way the player looks. Flags include ones such as 'on fire' and 'sprinting'.
	// The metadata values are indexed by their property key.
	EntityMetadata map[uint32]interface{}
	// Flags is a set of flags that specify certain properties of the player, such as whether or not it can
	// fly and/or move through blocks.
	Flags uint32
	// CommandPermissionLevel is a set of permissions that specify what commands a player is allowed to execute.
	CommandPermissionLevel uint32
	// ActionPermissions is, much like Flags, a set of flags that specify actions that the player is allowed
	// to undertake, such as whether it is allowed to edit blocks, open doors etc.
	ActionPermissions uint32
	// PermissionLevel is the permission level of the player as it shows up in the player list built up using
	// the PlayerList packet.
	PermissionLevel uint32
	// CustomStoredPermissions ...
	CustomStoredPermissions uint32
	// PlayerUniqueID is a unique identifier of the player. It appears it is not required to fill this field
	// out with a correct value. Simply writing 0 seems to work.
	PlayerUniqueID int64
	// EntityLinks is a list of entity links that are currently active on the player. These links alter the
	// way the player shows up when first spawned in terms of it shown as riding an entity. Setting these
	// links is important for new viewers to see the player is riding another entity.
	EntityLinks []protocol.EntityLink
	// DeviceID is the device ID set in one of the files found in the storage of the device of the player. It
	// may be changed freely, so it should not be relied on for anything.
	DeviceID string
	// BuildPlatform is the build platform/device OS of the player that is about to be added, as it sent in
	// the Login packet when joining.
	BuildPlatform int32
}

// ID ...
func (*AddPlayer) ID() uint32 {
	return IDAddPlayer
}

// Marshal ...
func (pk *AddPlayer) Marshal(w *protocol.Writer) {
	w.UUID(&pk.UUID)
	w.String(&pk.Username)
	w.Varint64(&pk.EntityUniqueID)
	w.Varuint64(&pk.EntityRuntimeID)
	w.String(&pk.PlatformChatID)
	w.Vec3(&pk.Position)
	w.Vec3(&pk.Velocity)
	w.Float32(&pk.Pitch)
	w.Float32(&pk.Yaw)
	w.Float32(&pk.HeadYaw)
	w.ItemInstance(&pk.HeldItem)
	w.EntityMetadata(&pk.EntityMetadata)
	w.Varuint32(&pk.Flags)
	w.Varuint32(&pk.CommandPermissionLevel)
	w.Varuint32(&pk.ActionPermissions)
	w.Varuint32(&pk.PermissionLevel)
	w.Varuint32(&pk.CustomStoredPermissions)
	w.Int64(&pk.PlayerUniqueID)
	protocol.WriteEntityLinks(w, &pk.EntityLinks)
	w.String(&pk.DeviceID)
	w.Int32(&pk.BuildPlatform)
}

// Unmarshal ...
func (pk *AddPlayer) Unmarshal(r *protocol.Reader) {
	r.UUID(&pk.UUID)
	r.String(&pk.Username)
	r.Varint64(&pk.EntityUniqueID)
	r.Varuint64(&pk.EntityRuntimeID)
	r.String(&pk.PlatformChatID)
	r.Vec3(&pk.Position)
	r.Vec3(&pk.Velocity)
	r.Float32(&pk.Pitch)
	r.Float32(&pk.Yaw)
	r.Float32(&pk.HeadYaw)
	r.ItemInstance(&pk.HeldItem)
	r.EntityMetadata(&pk.EntityMetadata)
	r.Varuint32(&pk.Flags)
	r.Varuint32(&pk.CommandPermissionLevel)
	r.Varuint32(&pk.ActionPermissions)
	r.Varuint32(&pk.PermissionLevel)
	r.Varuint32(&pk.CustomStoredPermissions)
	r.Int64(&pk.PlayerUniqueID)
	protocol.EntityLinks(r, &pk.EntityLinks)
	r.String(&pk.DeviceID)
	r.Int32(&pk.BuildPlatform)
}
