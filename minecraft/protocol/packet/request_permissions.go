package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// RequestPermissions is a packet sent from the client to the server to request permissions that the client does not
// currently have. It can only be sent by operators and host in vanilla Minecraft.
type RequestPermissions struct {
	// EntityUniqueID is the unique ID of the player. The unique ID is unique for the entire world and is
	// often used in packets. Most servers send an EntityUniqueID equal to the EntityRuntimeID.
	EntityUniqueID int64
	// PermissionLevel is the current permission level of the player. This is one of the constants that may be found
	// in the AdventureSettings packet.
	PermissionLevel uint8
	// RequestedPermissions contains the requested permission flags.
	RequestedPermissions uint16
}

// ID ...
func (*RequestPermissions) ID() uint32 {
	return IDRequestPermissions
}

// Marshal ...
func (pk *RequestPermissions) Marshal(w *protocol.Writer) {
	w.Int64(&pk.EntityUniqueID)
	w.Uint8(&pk.PermissionLevel)
	w.Uint16(&pk.RequestedPermissions)
}

// Unmarshal ...
func (pk *RequestPermissions) Unmarshal(r *protocol.Reader) {
	r.Int64(&pk.EntityUniqueID)
	r.Uint8(&pk.PermissionLevel)
	r.Uint16(&pk.RequestedPermissions)
}
