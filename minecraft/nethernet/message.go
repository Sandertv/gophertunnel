package nethernet

import (
	"fmt"
	"github.com/pion/webrtc/v4"
	"github.com/sandertv/gophertunnel/minecraft/nethernet/internal"
	"io"
)

func (c *Conn) handleRemoteMessage(message webrtc.DataChannelMessage) {
	if err := c.handleMessage(message.Data); err != nil {
		c.log.Error("error handling remote message", internal.ErrAttr(err))
	}
}

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
