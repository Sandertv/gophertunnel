package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ContainerClose is sent by the server to close a container the player currently has opened, which was opened
// using the ContainerOpen packet, or by the client to tell the server it closed a particular container, such
// as the crafting grid.
type ContainerClose struct {
	// WindowID is the ID representing the window of the container that should be closed. It must be equal to
	// the one sent in the ContainerOpen packet to close the designated window.
	WindowID byte
	// ContainerType is the type of container that the server is trying to close. This is used to validate on
	// the client side whether or not the server's close request is valid.
	ContainerType byte
	// ServerSide determines whether or not the container was force-closed by the server. If this value is
	// not set correctly, the client may ignore the packet and respond with a PacketViolationWarning.
	ServerSide bool
}

// ID ...
func (*ContainerClose) ID() uint32 {
	return IDContainerClose
}

func (pk *ContainerClose) Marshal(io protocol.IO) {
	io.Uint8(&pk.WindowID)
	io.Uint8(&pk.ContainerType)
	io.Bool(&pk.ServerSide)
}
