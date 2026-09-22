package protocol

import "github.com/go-gl/mathgl/mgl32"

const (
	// EntityLinkRemove is set to remove the link between two entities.
	EntityLinkRemove = iota
	// EntityLinkRider is set for entities that have control over the entity they're riding, such as in a
	// minecart.
	EntityLinkRider
	// EntityLinkPassenger is set for entities being a passenger of a vehicle they enter, such as the back
	// sit of a boat.
	EntityLinkPassenger
)

// EntityLink is a link between two entities, typically being one entity riding another.
type EntityLink struct {
	// RiddenEntityUniqueID is the entity unique ID of the entity that is being ridden. For a player sitting
	// in a boat, this is the unique ID of the boat.
	RiddenEntityUniqueID int64
	// RiderEntityUniqueID is the entity unique ID of the entity that is riding. For a player sitting in a
	// boat, this is the unique ID of the player.
	RiderEntityUniqueID int64
	// Type is one of the types above. It specifies the way the entity is linked to another entity.
	Type byte
	// Immediate is set to immediately dismount an entity from another. This should be set when the mount of
	// an entity is killed.
	Immediate bool
	// RiderInitiated specifies if the link was created by the rider, for example the player starting to ride
	// a horse by itself. This is generally true in vanilla environment for players.
	RiderInitiated bool
	// VehicleAngularVelocity is the angular velocity of the vehicle that the rider is riding.
	VehicleAngularVelocity float32
}

// Marshal encodes/decodes a single entity link.
func (x *EntityLink) Marshal(r IO) {
	r.ActorUniqueID(&x.RiddenEntityUniqueID)
	r.ActorUniqueID(&x.RiderEntityUniqueID)
	r.Uint8(&x.Type)
	r.Bool(&x.Immediate)
	r.Bool(&x.RiderInitiated)
	r.Float32(&x.VehicleAngularVelocity)
}

const (
	PassengerOfBlockEmoteStanding = iota
	PassengerOfBlockEmoteRiding
	PassengerOfBlockEmoteLaying
)

// PassengerOfBlockData holds the data of a block that an entity is riding.
type PassengerOfBlockData struct {
	// BlockPosition is the position of the block that the entity is riding.
	BlockPosition BlockPos
	// Offset is the offset of the entity relative to the block position. Each component must be in the range
	// -5 to 5.
	Offset mgl32.Vec3
	// Rotation is the rotation of the entity while riding the block, in degrees from 0 to 360.
	Rotation float32
	// RotationLimit is the limit of the rotation of the entity while riding the block, in degrees from 0 to
	// 360.
	RotationLimit float32
	// EmoteType is the pose the entity takes while riding the block. It is one of the PassengerOfBlockEmote
	// constants above.
	EmoteType uint8
}

// Marshal encodes/decodes a PassengerOfBlockData.
func (x *PassengerOfBlockData) Marshal(r IO) {
	r.BlockPos(&x.BlockPosition)
	r.Vec3(&x.Offset)
	r.Float32(&x.Rotation)
	r.Float32(&x.RotationLimit)
	r.Uint8(&x.EmoteType)
}
