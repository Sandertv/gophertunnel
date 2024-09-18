package signaling

import (
	"context"
	"encoding/json"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/df-mc/go-nethernet"
	"github.com/sandertv/gophertunnel/minecraft/franchise/internal"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Conn implements a [nethernet.Signaling] over a WebSocket connection.
//
// A Conn may be established using the methods of Dialer with either
// a [franchise.IdentityProvider] and an [Environment] or an [oauth2.TokenSource]
// for authorization.
//
// A Conn can be utilized with [nethernet.ListenConfig.Listen] or [nethernet.Dialer.DialContext].
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

// Signal sends a [nethernet.Signal] to a network.
func (c *Conn) Signal(signal *nethernet.Signal) error {
	return c.write(Message{
		Type: MessageTypeSignal,
		To:   signal.NetworkID,
		Data: signal.String(),
	})
}

// Notify registers a [nethernet.Notifier] to receive notifications of signals and errors. It returns
// a function to stop receiving notifications on the [nethernet.Notifier].
func (c *Conn) Notify(n nethernet.Notifier) (stop func()) {
	c.notifiersMu.Lock()
	i := c.notifyCount
	c.notifiers[i] = n
	c.notifyCount++
	c.notifiersMu.Unlock()

	return c.stopFunc(i, n)
}

// stopFunc returns a function to be returned by [Conn.Notify], which stops receiving notifications
// on the Notifier by unregistering them on the Conn with notifying [nethernet.ErrSignalingStopped]
// as an error through [nethernet.Notifier.NotifyError].
func (c *Conn) stopFunc(i uint32, n nethernet.Notifier) func() {
	return func() {
		n.NotifyError(nethernet.ErrSignalingStopped)

		c.notifiersMu.Lock()
		delete(c.notifiers, i)
		c.notifiersMu.Unlock()
	}
}

// Credentials blocks until [nethernet.Credentials] are received from the server or the [context.Context]
// is done. It returns a [nethernet.Credentials] or an error if the Conn is closed or the [context.Context]
// is canceled or exceeded a deadline.
func (c *Conn) Credentials(ctx context.Context) (*nethernet.Credentials, error) {
	select {
	case <-c.closed:
		return nil, net.ErrClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return c.credentials.Load(), nil
	}
}

// NetworkID returns the network ID of the Conn. It may be specified from [Dialer.NetworkID], otherwise a random
// value will be automatically set from [rand.Uint64] in set up during [Dialer.DialContext]. It is utilized by
// [nethernet.Listener] and [nethernet.Dialer] to obtain its local network ID to listen.
func (c *Conn) NetworkID() uint64 {
	return c.d.NetworkID
}

// Close closes the Conn and unregisters any notifiers. It ensures that the Conn is closed only once.
// It unregisters all notifiers registered on the Conn with notifying [nethernet.ErrSignalingStopped].
func (c *Conn) Close() (err error) {
	c.once.Do(func() {
		c.notifiersMu.Lock()
		for _, n := range c.notifiers {
			n.NotifyError(nethernet.ErrSignalingStopped)
		}
		clear(c.notifiers)
		c.notifiersMu.Unlock()

		close(c.closed)
		err = c.conn.Close(websocket.StatusNormalClosure, "")
	})
	return err
}

// read continuously reads messages from the WebSocket connection and handles them.
// It also sends a Message of MessageTypePing at 15 seconds intervals to keep the
// Conn alive. It goes as a background goroutine of the Conn and handles different
// types of messages: credentials, signals, and errors. It closes the Conn if it
// encounters an error or when the Conn is closed.
func (c *Conn) read() {
	go func() {
		ticker := time.NewTicker(time.Second * 15)
		defer ticker.Stop()

		for {
			select {
			case <-c.closed:
				return
			case <-ticker.C:
				if err := c.write(Message{
					Type: MessageTypePing,
				}); err != nil {
					c.d.Log.Error("error writing ping", internal.ErrAttr(err))
					return
				}
			}
		}
	}()
	defer c.Close()

	for {
		var message Message
		if err := wsjson.Read(context.Background(), c.conn, &message); err != nil {
			return
		}
		switch message.Type {
		case MessageTypeCredentials:
			if message.From != "Server" {
				c.d.Log.Warn("received credentials from non-Server", slog.Any("message", message))
				continue
			}
			var credentials nethernet.Credentials
			if err := json.Unmarshal([]byte(message.Data), &credentials); err != nil {
				c.d.Log.Error("error decoding credentials", internal.ErrAttr(err))
				continue
			}
			notifyCredentials := c.credentials.Load() == nil
			c.credentials.Store(&credentials)
			if notifyCredentials {
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
		case MessageTypeError:
			var err Error
			if err2 := json.Unmarshal([]byte(message.Data), &err); err2 != nil {
				c.d.Log.Error("error decoding error", internal.ErrAttr(err2))
				continue
			}

			c.notifiersMu.Lock()
			for _, n := range c.notifiers {
				n.NotifyError(&err)
			}
			c.notifiersMu.Unlock()
		default:
			c.d.Log.Warn("received message for unknown type", slog.Any("message", message))
		}
	}
}

// write encodes the given Message and sends it over the WebSocket connection. It uses a background context
// to avoid issues with context cancellation affecting the connection. An error may be returned if the message
// could not be sent.
func (c *Conn) write(message Message) error {
	return wsjson.Write(context.Background(), c.conn, message)
}
