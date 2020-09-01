package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ContainerOpen is sent by the server to open a container client-side. This container must be physically
// present in the world, for the packet to have any effect. Unlike Java Edition, Bedrock Edition requires that
// chests for example must be present and in range to open its inventory.
type ContainerOpen struct {
	// WindowID is the ID representing the window that is being opened. It may be used later to close the
	// container using a ContainerClose packet.
	WindowID byte
	// ContainerType is the type ID of the container that is being opened when opening the container at the
	// position of the packet. It depends on the block/entity, and could, for example, be the window type of
	// a chest or a hopper, but also a horse inventory.
	ContainerType byte
	// ContainerPosition is the position of the container opened. The position must point to a block entity
	// that actually has a container. If that is not the case, the window will not be opened and the packet
	// will be ignored, if a valid ContainerEntityUniqueID has not also been provided.
	ContainerPosition protocol.BlockPos
	// ContainerEntityUniqueID is the unique ID of the entity container that was opened. It is only used if
	// the ContainerType is one that points to an entity, for example a horse.
	ContainerEntityUniqueID int64
}

// ID ...
func (*ContainerOpen) ID() uint32 {
	return IDContainerOpen
}

// Marshal ...
func (pk *ContainerOpen) Marshal(w *protocol.Writer) {
	w.Uint8(&pk.WindowID)
	w.Uint8(&pk.ContainerType)
	w.UBlockPos(&pk.ContainerPosition)
	w.Varint64(&pk.ContainerEntityUniqueID)
}

// Unmarshal ...
func (pk *ContainerOpen) Unmarshal(r *protocol.Reader) {
	r.Uint8(&pk.WindowID)
	r.Uint8(&pk.ContainerType)
	r.UBlockPos(&pk.ContainerPosition)
	r.Varint64(&pk.ContainerEntityUniqueID)
}
