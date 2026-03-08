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

// LocatorBarWaypointPayload represents a waypoint entry in the locator bar packet.
type LocatorBarWaypointPayload struct {
	// GroupHandle is the UUID handle for the waypoint group.
	GroupHandle uuid.UUID
	// Payload contains the waypoint data.
	Payload ServerWaypointPayload
	// ActionFlag determines the action for this waypoint. It is one of the WaypointAction constants.
	ActionFlag uint8
}

// Marshal encodes/decodes a LocatorBarWaypointPayload.
func (x *LocatorBarWaypointPayload) Marshal(r IO) {
	r.UUID(&x.GroupHandle)
	Single(r, &x.Payload)
	r.Uint8(&x.ActionFlag)
}

// ServerWaypointPayload contains waypoint data with optional fields controlled by a bitmask.
type ServerWaypointPayload struct {
	// UpdateFlag is a bitmask that controls which optional fields are present.
	UpdateFlag uint32
	// Visible indicates if the waypoint is visible.
	Visible bool
	// Position is the world position of the waypoint.
	Position mgl32.Vec3
	// DimensionID is the dimension the waypoint is in.
	DimensionID int32
	// TextureID is the texture identifier for the waypoint icon.
	TextureID uint32
	// Colour is the colour of the waypoint.
	Colour int32
	// ClientPositionAuthority indicates if the client has position authority.
	ClientPositionAuthority bool
	// ActorUniqueID is the unique ID of an actor the waypoint is attached to.
	ActorUniqueID int64
}

// Marshal encodes/decodes a ServerWaypointPayload.
func (x *ServerWaypointPayload) Marshal(r IO) {
	r.Uint32(&x.UpdateFlag)
	if x.UpdateFlag&WaypointUpdateFlagVisible != 0 {
		r.Bool(&x.Visible)
	}
	if x.UpdateFlag&WaypointUpdateFlagPosition != 0 {
		r.Vec3(&x.Position)
		r.Varint32(&x.DimensionID)
	}
	if x.UpdateFlag&WaypointUpdateFlagTextureID != 0 {
		r.Uint32(&x.TextureID)
	}
	if x.UpdateFlag&WaypointUpdateFlagColour != 0 {
		r.Int32(&x.Colour)
	}
	if x.UpdateFlag&WaypointUpdateFlagClientPositionAuthority != 0 {
		r.Bool(&x.ClientPositionAuthority)
	}
	if x.UpdateFlag&WaypointUpdateFlagActorUniqueID != 0 {
		r.Varint64(&x.ActorUniqueID)
	}
}
