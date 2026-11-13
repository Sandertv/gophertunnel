package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// VirtualContainerOpen is sent by the server to open a virtual container screen that is
// not tied to any block or entity.
type VirtualContainerOpen struct {
	// WindowID is the ID of the inventory window being opened.
	WindowID uint8
	// SessionID is the ID that identifies this container. It is later used by the
	// VirtualContainerClose packet to identify which virtual container was closed.
	SessionID uint64
	// Title is the title of the virtual container. It can be up to 256 bytes.
	// A localization key may be specified to display a localized title in the client.
	Title string
	// Content is the initial content of the virtual container. The length of the slice
	// determines the size of the container, which can have up to 54 slots.
	Content []protocol.ItemInstance
}

// ID ...
func (*VirtualContainerOpen) ID() uint32 {
	return IDVirtualContainerOpen
}

func (pk *VirtualContainerOpen) Marshal(io protocol.IO) {
	io.Uint8(&pk.WindowID)
	io.Varuint64(&pk.SessionID)
	io.String(&pk.Title)
	count := uint32(len(pk.Content))
	io.Varuint32(&count)
	if count < 1 || count > 54 {
		io.InvalidValue(count, "content length", "must be between 1 and 54")
	}
	protocol.FuncSliceOfLen(io, count, &pk.Content, io.ItemInstance)
}
