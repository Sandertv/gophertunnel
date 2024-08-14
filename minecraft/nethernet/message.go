package nethernet

import (
	"fmt"
	"io"
)

// TODO: Probably the structure of remote messages sent in both ReliableDataChannel and UnreliableDataChannel
// are changed since whenever, and the specification might be outdated. We need to reverse that too.

func (c *Conn) handleMessage(b []byte) error {
	if len(b) < 2 {
		return io.ErrUnexpectedEOF
	}
	segments := b[0]
	data := b[1:]

	if c.promisedSegments > 0 && c.promisedSegments-1 != segments {
		return fmt.Errorf("invalid promised segments: expected %d, got %d", c.promisedSegments-1, segments)
	}
	c.promisedSegments = segments

	c.buf.Write(data)

	if c.promisedSegments > 0 {
		return nil
	}

	c.packets <- c.buf.Bytes()
	c.buf.Reset()

	return nil
}
