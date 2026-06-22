package minecraft

import (
	"errors"
	"net"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

var errBufferTooSmall = errors.New("a message sent was larger than the buffer used to receive the message into")

// wrap wraps the error passed into a net.OpError with the op as operation and returns it, or nil if the error
// passed is nil.
func (conn *Conn) wrap(err error, op string) error {
	if err == nil {
		return nil
	}
	return &net.OpError{
		Op:     op,
		Net:    "minecraft",
		Source: conn.LocalAddr(),
		Addr:   conn.RemoteAddr(),
		Err:    err,
	}
}

// DisconnectError is an error returned by operations from Conn when the connection is closed by the other
// end through a packet.Disconnect. It is wrapped in a net.OpError and may be obtained using
// errors.Unwrap(net.OpError).
type DisconnectError string

// Error returns the message held in the packet.Disconnect.
func (d DisconnectError) Error() string {
	return string(d)
}

// DisconnectPacketError is returned when the other end closes the connection
// through a packet.Disconnect. It preserves the original disconnect packet
// fields while still unwrapping to DisconnectError for older callers.
type DisconnectPacketError struct {
	Reason                  int32
	HideDisconnectionScreen bool
	Message                 string
	FilteredMessage         string
	DisplayMessage          string
}

// Error returns the message that should be shown for the disconnect.
func (d *DisconnectPacketError) Error() string {
	if d == nil {
		return "<nil>"
	}
	if d.DisplayMessage != "" {
		return d.DisplayMessage
	}
	if d.Message != "" {
		return d.Message
	}
	return "Disconnected"
}

// Unwrap returns the legacy DisconnectError so existing errors.As checks keep
// working.
func (d *DisconnectPacketError) Unwrap() error {
	return DisconnectError(d.Error())
}

// Packet returns a copy of the original disconnect packet fields.
func (d *DisconnectPacketError) Packet() *packet.Disconnect {
	if d == nil {
		return nil
	}
	return &packet.Disconnect{
		Reason:                  d.Reason,
		HideDisconnectionScreen: d.HideDisconnectionScreen,
		Message:                 d.Message,
		FilteredMessage:         d.FilteredMessage,
	}
}
