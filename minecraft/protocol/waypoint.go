package protocol

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

const (
	WaypointActionNone = iota
	WaypointActionAdd
	WaypointActionRemove
	WaypointActionUpdate
)

const (
	WaypointUpdateFlagVisible = 1 << iota
	WaypointUpdateFlagPosition
	WaypointUpdateFlagTextureID
	WaypointUpdateFlagColour
	WaypointUpdateFlagClientPositionAuthority
	WaypointUpdateFlagActorUniqueID
)

const (
	WaypointTextureSquare = iota + 2
	WaypointTextureCircle
	WaypointTextureSmallSquare
	WaypointTextureSmallStar
)

// LocatorBarWaypoint represents a waypoint entry in the locator bar packet.
type LocatorBarWaypoint struct {
	// GroupHandle is the UUID handle for the waypoint group.
	GroupHandle uuid.UUID
	// Waypoint contains the waypoint data.
	Waypoint Waypoint
	// Action determines the action for this waypoint. It is one of the WaypointAction constants.
	Action uint8
}

// Marshal encodes/decodes a LocatorBarWaypoint.
func (x *LocatorBarWaypoint) Marshal(r IO) {
	r.UUID(&x.GroupHandle)
	Single(r, &x.Waypoint)
	r.Uint8(&x.Action)
}

// WaypointWorldPosition holds a position and dimension for a waypoint.
type WaypointWorldPosition struct {
	// Position is the world position of the waypoint.
	Position mgl32.Vec3
	// DimensionID is the dimension the waypoint is in.
	DimensionID int32
}

// Marshal encodes/decodes a WaypointWorldPosition.
func (x *WaypointWorldPosition) Marshal(r IO) {
	r.Vec3(&x.Position)
	r.Varint32(&x.DimensionID)
}

// Waypoint holds optional data for a locator bar waypoint.
type Waypoint struct {
	// UpdateFlag is a bitmask indicating which optional fields are set.
	UpdateFlag uint32
	// Visible determines whether the waypoint is shown.
	Visible Optional[bool]
	// WorldPosition is the position and dimension of the waypoint.
	WorldPosition Optional[WaypointWorldPosition]
	// TextureID is the icon texture of the waypoint. It is one of the WaypointTexture constants.
	TextureID Optional[uint32]
	// Colour is the RGB colour used to tint the waypoint icon.
	Colour Optional[int32]
	// ClientPositionAuthority determines whether the client has authority over the waypoint position.
	ClientPositionAuthority Optional[bool]
	// ActorUniqueID is the unique ID of the entity the waypoint tracks.
	ActorUniqueID Optional[int64]
}

// Marshal encodes/decodes a Waypoint.
func (x *Waypoint) Marshal(r IO) {
	r.Uint32(&x.UpdateFlag)
	OptionalFunc(r, &x.Visible, r.Bool)
	OptionalMarshaler(r, &x.WorldPosition)
	OptionalFunc(r, &x.TextureID, r.Uint32)
	OptionalFunc(r, &x.Colour, r.Int32)
	OptionalFunc(r, &x.ClientPositionAuthority, r.Bool)
	OptionalFunc(r, &x.ActorUniqueID, r.Varint64)
}
