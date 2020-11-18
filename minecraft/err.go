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

// wrap wraps the error passed into a net.OpError with the op as operation and returns it, or nil if the error
// passed is nil. Additionally, the returned net.OpError returns true when err.Timeout() is called.
func (conn *Conn) wrapTimeout(err error, op string) error {
	if err == nil {
		return nil
	}
	return conn.wrap(timeoutErr{err: err}, op)
}

// timeoutErr wraps around an error and implements the net.timeout interface.
type timeoutErr struct {
	err error
}

// Error ...
func (t timeoutErr) Error() string {
	return t.err.Error()
}

// Timeout ...
func (t timeoutErr) Timeout() bool {
	return true
}
