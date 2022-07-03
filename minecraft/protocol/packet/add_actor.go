package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AddActor is sent by the server to the client to spawn an entity to the player. It is used for every entity
// except other players, for which the AddPlayer packet is used.
type AddActor struct {
	// EntityUniqueID is the unique ID of the entity. The unique ID is a value that remains consistent across
	// different sessions of the same world, but most servers simply fill the runtime ID of the entity out for
	// this field.
	EntityUniqueID int64
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// EntityType is the string entity type of the entity, for example 'minecraft:skeleton'. A list of these
	// entities may be found online.
	EntityType string
	// Position is the position to spawn the entity on. If the entity is on a distance that the player cannot
	// see it, the entity will still show up if the player moves closer.
	Position mgl32.Vec3
	// Velocity is the initial velocity the entity spawns with. This velocity will initiate client side
	// movement of the entity.
	Velocity mgl32.Vec3
	// Pitch is the vertical rotation of the entity. Facing straight forward yields a pitch of 0. Pitch is
	// measured in degrees.
	Pitch float32
	// Yaw is the horizontal rotation of the entity. Yaw is also measured in degrees.
	Yaw float32
	// HeadYaw is the same as Yaw, except that it applies specifically to the head of the entity. A different value for
	// HeadYaw than Yaw means that the entity will have its head turned.
	HeadYaw float32
	// BodyYaw is the same as Yaw, except that it applies specifically to the body of the entity. A different value for
	// BodyYaw than HeadYaw means that the entity will have its body turned, although it is unclear what the difference
	// between BodyYaw and Yaw is.
	BodyYaw float32
	// Attributes is a slice of attributes that the entity has. It includes attributes such as its health,
	// movement speed, etc.
	Attributes []protocol.Attribute
	// EntityMetadata is a map of entity metadata, which includes flags and data properties that alter in
	// particular the way the entity looks. Flags include ones such as 'on fire' and 'sprinting'.
	// The metadata values are indexed by their property key.
	EntityMetadata map[uint32]any
	// EntityLinks is a list of entity links that are currently active on the entity. These links alter the
	// way the entity shows up when first spawned in terms of it shown as riding an entity. Setting these
	// links is important for new viewers to see the entity is riding another entity.
	EntityLinks []protocol.EntityLink
}

// ID ...
func (*AddActor) ID() uint32 {
	return IDAddActor
}

// Marshal ...
func (pk *AddActor) Marshal(w *protocol.Writer) {
	w.Varint64(&pk.EntityUniqueID)
	w.Varuint64(&pk.EntityRuntimeID)
	w.String(&pk.EntityType)
	w.Vec3(&pk.Position)
	w.Vec3(&pk.Velocity)
	w.Float32(&pk.Pitch)
	w.Float32(&pk.Yaw)
	w.Float32(&pk.HeadYaw)
	w.Float32(&pk.BodyYaw)
	protocol.WriteInitialAttributes(w, &pk.Attributes)
	w.EntityMetadata(&pk.EntityMetadata)
	protocol.WriteEntityLinks(w, &pk.EntityLinks)
}

// Unmarshal ...
func (pk *AddActor) Unmarshal(r *protocol.Reader) {
	r.Varint64(&pk.EntityUniqueID)
	r.Varuint64(&pk.EntityRuntimeID)
	r.String(&pk.EntityType)
	r.Vec3(&pk.Position)
	r.Vec3(&pk.Velocity)
	r.Float32(&pk.Pitch)
	r.Float32(&pk.Yaw)
	r.Float32(&pk.HeadYaw)
	r.Float32(&pk.BodyYaw)
	protocol.InitialAttributes(r, &pk.Attributes)
	r.EntityMetadata(&pk.EntityMetadata)
	protocol.EntityLinks(r, &pk.EntityLinks)
}
