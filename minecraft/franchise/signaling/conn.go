package signaling

import (
	"context"
	"encoding/json"
	"github.com/sandertv/gophertunnel/minecraft/franchise/internal"
	"github.com/sandertv/gophertunnel/minecraft/nethernet"
	"net"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Conn struct {
	conn *websocket.Conn
	d    Dialer

	credentials         atomic.Pointer[nethernet.Credentials]
	credentialsReceived chan struct{}

	once   sync.Once
	closed chan struct{}

	signals chan *nethernet.Signal
}

func (c *Conn) WriteSignal(signal *nethernet.Signal) error {
	return c.write(Message{
		Type: MessageTypeSignal,
		To:   json.Number(strconv.FormatUint(signal.NetworkID, 10)),
		Data: signal.String(),
	})
}

func (c *Conn) ReadSignal(cancel <-chan struct{}) (*nethernet.Signal, error) {
	select {
	case <-cancel:
		return nil, nethernet.ErrSignalingCanceled
	case <-c.closed:
		return nil, net.ErrClosed
	case s := <-c.signals:
		return s, nil
	}
}

func (c *Conn) Credentials() (*nethernet.Credentials, error) {
	select {
	case <-c.closed:
		return nil, net.ErrClosed
	default:
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
		case <-c.closed:
			return
		}
	}
}

func (c *Conn) read() {
	for {
		var message Message
		if err := wsjson.Read(context.Background(), c.conn, &message); err != nil {
			_ = c.Close()
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
			previous := c.credentials.Load()
			c.credentials.Store(&credentials)
			if previous == nil {
				close(c.credentialsReceived)
			}
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
		close(c.closed)
		close(c.signals)
		err = c.conn.Close(websocket.StatusNormalClosure, "")
	})
	return err
}
