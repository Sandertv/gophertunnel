package signaling

import (
	"context"
	"encoding/json"
	"github.com/df-mc/go-nethernet"
	"github.com/sandertv/gophertunnel/minecraft/franchise/internal"
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

	notifyCount uint32
	notifiers   map[uint32]nethernet.Notifier
	notifiersMu sync.Mutex
}

func (c *Conn) Signal(signal *nethernet.Signal) error {
	return c.write(Message{
		Type: MessageTypeSignal,
		To:   json.Number(strconv.FormatUint(signal.NetworkID, 10)),
		Data: signal.String(),
	})
}

func (c *Conn) Notify(cancel <-chan struct{}, n nethernet.Notifier) {
	c.notifiersMu.Lock()
	i := c.notifyCount
	c.notifiers[i] = n
	c.notifyCount++
	c.notifiersMu.Unlock()

	go c.notify(cancel, n, i)
}

func (c *Conn) notify(cancel <-chan struct{}, n nethernet.Notifier, i uint32) {
	select {
	case <-c.closed:
		n.NotifyError(net.ErrClosed)
	case <-cancel:
		n.NotifyError(nethernet.ErrSignalingCanceled)
	}

	c.notifiersMu.Lock()
	delete(c.notifiers, i)
	c.notifiersMu.Unlock()
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
			signal := &nethernet.Signal{}
			if err := signal.UnmarshalText([]byte(message.Data)); err != nil {
				c.d.Log.Error("error decoding signal", internal.ErrAttr(err))
				continue
			}
			var err error
			signal.NetworkID, err = strconv.ParseUint(message.From, 10, 64)
			if err != nil {
				c.d.Log.Error("error parsing network ID of signal", internal.ErrAttr(err))
				continue
			}

			c.notifiersMu.Lock()
			for _, n := range c.notifiers {
				n.NotifySignal(signal)
			}
			c.notifiersMu.Unlock()
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
		err = c.conn.Close(websocket.StatusNormalClosure, "")
	})
	return err
}
