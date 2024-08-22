package nethernet

import (
	"fmt"
	"io"
)

// message represents the structure of remote messages sent in ReliableDataChannel.
type message struct {
	segments uint8
	data     []byte
}

func parseMessage(b []byte) (*message, error) {
	if len(b) < 2 {
		return nil, io.ErrUnexpectedEOF
	}
	return &message{
		segments: b[0],
		data:     b[1:],
	}, nil
}

func (c *Conn) handleMessage(b []byte) error {
	msg, err := parseMessage(b)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	if c.message.segments > 0 && c.message.segments-1 != msg.segments {
		return fmt.Errorf("invalid promised segments: expected %d, got %d", c.message.segments-1, msg.segments)
	}
	c.message.segments = msg.segments

	c.message.data = append(c.message.data, msg.data...)

	if c.message.segments > 0 {
		return nil
	}

	c.packets <- c.message.data
	c.message.data = nil

	return nil
}
