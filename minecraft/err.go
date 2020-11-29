package minecraft

import (
	"errors"
	"net"
)

var (
	// TODO: Change this to net.ErrClosed in 1.16.
	errClosed         = errors.New("use of closed network connection")
	errBufferTooSmall = errors.New("a message sent was larger than the buffer used to receive the message into")
	errListenerClosed = errors.New("use of closed listener")
)

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
