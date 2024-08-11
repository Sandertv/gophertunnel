package signaling

import (
	"context"
	"encoding/json"
	"github.com/sandertv/gophertunnel/minecraft/franchise/internal"
	"github.com/sandertv/gophertunnel/minecraft/nethernet"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Conn struct {
	conn *websocket.Conn
	ctx  context.Context
	d    Dialer

	credentials atomic.Pointer[nethernet.Credentials]
	ready       chan struct{}

	once sync.Once

	signals chan *nethernet.Signal
}

func (c *Conn) WriteSignal(signal *nethernet.Signal) error {
	return c.write(Message{
		Type: MessageTypeSignal,
		To:   json.Number(strconv.FormatUint(signal.NetworkID, 10)),
		Data: signal.String(),
	})
}

func (c *Conn) ReadSignal() (*nethernet.Signal, error) {
	select {
	case s := <-c.signals:
		return s, nil
	case <-c.ctx.Done():
		return nil, context.Cause(c.ctx)
	}
}

func (c *Conn) Credentials() (*nethernet.Credentials, error) {
	select {
	case <-c.ctx.Done():
		return nil, context.Cause(c.ctx)
	case <-c.ready:
		return c.credentials.Load(), nil
	}
}

func (c *Conn) ping() {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.write(Message{
				Type: MessageTypeRequestPing,
			}); err != nil {
				c.d.Log.Error("error writing ping", internal.ErrAttr(err))
			}
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Conn) read(cancel context.CancelCauseFunc) {
	for {
		var message Message
		if err := wsjson.Read(context.Background(), c.conn, &message); err != nil {
			cancel(err)
			return
		}
		switch message.Type {
		case MessageTypeCredentials:
			if message.From != "Server" {
				c.d.Log.Warn("received credentials from non-Server", "message", message)
				continue
			}
			var credentials nethernet.Credentials
			if err := json.Unmarshal([]byte(message.Data), &credentials); err != nil {
				c.d.Log.Error("error decoding credentials", internal.ErrAttr(err))
				continue
			}
			c.credentials.Store(&credentials)
			close(c.ready)
		case MessageTypeSignal:
			s := &nethernet.Signal{}
			if err := s.UnmarshalText([]byte(message.Data)); err != nil {
				c.d.Log.Error("error decoding signal", internal.ErrAttr(err))
				continue
			}
			var err error
			s.NetworkID, err = strconv.ParseUint(message.From, 10, 64)
			if err != nil {
				c.d.Log.Error("error parsing network ID of signal", internal.ErrAttr(err))
				continue
			}
			c.signals <- s
		default:
			c.d.Log.Warn("received message for unknown type", "message", message)
		}
	}
}

func (c *Conn) write(m Message) error {
	return wsjson.Write(context.Background(), c.conn, m)
}

func (c *Conn) Close() (err error) {
	c.once.Do(func() {
		err = c.conn.Close(websocket.StatusNormalClosure, "")
	})
	return err
}
