package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Disconnect may be sent by the server to disconnect the client using an optional message to send as the
// disconnect screen.
type Disconnect struct {
	// HideDisconnectionScreen specifies if the disconnection screen should be hidden when the client is
	// disconnected, meaning it will be sent directly to the main menu.
	HideDisconnectionScreen bool
	// Message is an optional message to show when disconnected. This message is only written if the
	// HideDisconnectionScreen field is set to true.
	Message string
}

// ID ...
func (*Disconnect) ID() uint32 {
	return IDDisconnect
}

// Marshal ...
func (pk *Disconnect) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.HideDisconnectionScreen)
	if !pk.HideDisconnectionScreen {
		_ = protocol.WriteString(buf, pk.Message)
	}
}

// Unmarshal ...
func (pk *Disconnect) Unmarshal(buf *bytes.Buffer) error {
	if err := binary.Read(buf, binary.LittleEndian, &pk.HideDisconnectionScreen); err != nil {
		return err
	}
	if !pk.HideDisconnectionScreen {
		return protocol.String(buf, &pk.Message)
	}
	return nil
}
