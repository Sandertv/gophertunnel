package protocol

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
	r.Varint64(&x.RiddenEntityUniqueID)
	r.Varint64(&x.RiderEntityUniqueID)
	r.Uint8(&x.Type)
	r.Bool(&x.Immediate)
	r.Bool(&x.RiderInitiated)
	r.Float32(&x.VehicleAngularVelocity)
}
