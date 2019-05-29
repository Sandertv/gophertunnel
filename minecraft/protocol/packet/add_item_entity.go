package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AddItemEntity is sent by the server to the client to make an item entity show up. It is one of the few
// entities that cannot be sent using the AddEntity packet
type AddItemEntity struct {
	// EntityUniqueID is the unique ID of the entity. The unique ID is a value that remains consistent across
	// different sessions of the same world, but most servers simply fill the runtime ID of the entity out for
	// this field.
	EntityUniqueID int64
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// Item is the item that is spawned. It must have a valid ID for it to show up client-side. If it is not
	// a valid item, the client will crash when coming near.
	Item protocol.ItemStack
	// Position is the position to spawn the entity on. If the entity is on a distance that the player cannot
	// see it, the entity will still show up if the player moves closer.
	Position mgl32.Vec3
	// Velocity is the initial velocity the entity spawns with. This velocity will initiate client side
	// movement of the entity.
	Velocity mgl32.Vec3
	// EntityMetadata is a map of entity metadata, which includes flags and data properties that alter in
	// particular the way the entity looks. Flags include ones such as 'on fire' and 'sprinting'.
	// The metadata values are indexed by their property key.
	EntityMetadata map[uint32]interface{}
	// FromFishing specifies if the item was obtained by fishing it up using a fishing rod. It is not clear
	// why the client needs to know this.
	FromFishing bool
}

// ID ...
func (*AddItemEntity) ID() uint32 {
	return IDAddItemEntity
}

// Marshal ...
func (pk *AddItemEntity) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint64(buf, pk.EntityUniqueID)
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	_ = protocol.WriteItem(buf, pk.Item)
	_ = protocol.WriteVec3(buf, pk.Position)
	_ = protocol.WriteVec3(buf, pk.Velocity)
	_ = protocol.WriteEntityMetadata(buf, pk.EntityMetadata)
	_ = binary.Write(buf, binary.LittleEndian, pk.FromFishing)
}

// Unmarshal ...
func (pk *AddItemEntity) Unmarshal(buf *bytes.Buffer) error {
	pk.EntityMetadata = map[uint32]interface{}{}
	return chainErr(
		protocol.Varint64(buf, &pk.EntityUniqueID),
		protocol.Varuint64(buf, &pk.EntityRuntimeID),
		protocol.Item(buf, &pk.Item),
		protocol.Vec3(buf, &pk.Position),
		protocol.Vec3(buf, &pk.Velocity),
		protocol.EntityMetadata(buf, &pk.EntityMetadata),
		binary.Read(buf, binary.LittleEndian, &pk.FromFishing),
	)
}
