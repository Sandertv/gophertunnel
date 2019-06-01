package packet

import (
	"bytes"
	"encoding/binary"
)

// ContainerClose is sent by the server to close a container the player currently has opened, which was opened
// using the ContainerOpen packet.
type ContainerClose struct {
	// WindowID is the ID representing the window of the container that should be closed. It must be equal to
	// the one sent in the ContainerOpen packet to close the designated window.
	WindowID byte
}

// ID ...
func (*ContainerClose) ID() uint32 {
	return IDContainerClose
}

// Marshal ...
func (pk *ContainerClose) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.WindowID)
}

// Unmarshal ...
func (pk *ContainerClose) Unmarshal(buf *bytes.Buffer) error {
	return binary.Read(buf, binary.LittleEndian, &pk.WindowID)
}
