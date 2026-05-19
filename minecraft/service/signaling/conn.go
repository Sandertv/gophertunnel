package signaling

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/df-mc/go-nethernet"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/service/signaling/internal"
)

// Conn implements a [nethernet.Signaling] over a WebSocket connection.
type Conn struct {
	// conn is the underlying websocket connection with signaling service.
	conn *websocket.Conn
	// d is the Dialer used to dial this Conn.
	d Dialer

	// credentials holds ICE credentials received from messages with MessageTypeCredentials.
	// credentials might be changed if it has exceeded the expiration time.
	credentials atomic.Pointer[nethernet.Credentials]
	// credentialsReceived is a channel that is closed when credentials is received for the first time.
	credentialsReceived chan struct{}

	// once ensures that closure of the Conn occurs only once.
	once sync.Once
	// ctx is the background context for the Conn.
	// It is canceled when an error is returned from the underlying websocket connection.
	ctx    context.Context
	cancel context.CancelCauseFunc

	notifier *internal.Notifier

	pending *internal.PendingMap
}

// Signal sends a [nethernet.Signal] to a network.
func (conn *Conn) Signal(ctx context.Context, signal *nethernet.Signal) error {
	id := uuid.New()
	ch := conn.pending.Add(id)
	defer conn.pending.Remove(id)

	if err := conn.write(Message{
		Type: MessageTypeSignal,
		To:   signal.NetworkID,
		Data: signal.String(),
		ID:   id,
	}); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-conn.ctx.Done():
		return context.Cause(conn.ctx)
	case err := <-ch:
		return err
	}
}

// Notify registers a channel to receive incoming NetherNet signals.
//
// The returned stop function unregisters the channel and closes it. Callers must not close
// the channel themselves.
func (conn *Conn) Notify(signals chan<- *nethernet.Signal) (stop func()) {
	return conn.notifier.Register(signals)
}

// complete resolves the expectation registered for the outbound Message with
// the given ID.
// It is called when the signaling service sends a completion frame for that
// message ID, such as [MessageTypeDelivered] or [MessageTypeError].
func (conn *Conn) complete(id uuid.UUID, err error) {
	if !conn.pending.Done(id, err) {
		conn.d.Log.Warn("unexpected message ID", slog.String("id", id.String()), slog.Any("result", err))
	}
}

// Credentials blocks until [nethernet.Credentials] are received from the server or the [context.Context]
// is done. It returns a [nethernet.Credentials] or an error if the Conn is closed or the [context.Context]
// is canceled or exceeded a deadline.
func (conn *Conn) Credentials(ctx context.Context) (*nethernet.Credentials, error) {
	select {
	case <-conn.ctx.Done():
		return nil, context.Cause(conn.ctx)
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-conn.credentialsReceived:
		return conn.credentials.Load(), nil
	}
}

func (conn *Conn) PongData(b []byte) {
}

// NetworkID returns the network ID of the Conn. It may be specified from [Dialer.NetworkID], otherwise a random
// value will be automatically set from [rand.Uint64] in set up during [Dialer.DialContext]. It is utilized by
// [nethernet.Listener] and [nethernet.Dialer] to obtain its local network ID to listen.
func (conn *Conn) NetworkID() string {
	return conn.d.NetworkID
}

// Close closes the Conn and unregisters any notifiers. It ensures that the Conn is closed only once.
// It unregisters all notifiers registered on the Conn with notifying [nethernet.ErrSignalingStopped].
func (conn *Conn) Close() (err error) {
	return conn.close(net.ErrClosed)
}

// Context returns the background context of the [Conn].
// It is canceled if the underlying WebSocket connection was closed.
func (conn *Conn) Context() context.Context {
	return conn.ctx
}

func (conn *Conn) close(cause error) (err error) {
	conn.once.Do(func() {
		conn.d.Log.Debug("closing connection", slog.Any("cause", cause))

		_ = conn.notifier.Close()

		conn.cancel(cause)
		err = conn.conn.Close(websocket.StatusNormalClosure, "")
	})
	return err
}

// read continuously reads messages from the WebSocket connection and handles them.
// It also sends a Message of MessageTypePing at 15 seconds intervals to keep the
// Conn alive. It goes as a background goroutine of the Conn and handles different
// types of messages: credentials, signals, and errors. It closes the Conn if it
// encounters an error or when the Conn is closed.
func (conn *Conn) read() {
	go func() {
		ticker := time.NewTicker(time.Second * 15)
		defer ticker.Stop()

		for {
			select {
			case <-conn.ctx.Done():
				return
			case <-ticker.C:
				if err := conn.write(Message{
					Type: MessageTypePing,
				}); err != nil {
					_ = conn.close(fmt.Errorf("background: write ping: %w", err))
					return
				}
			}
		}
	}()

	for {
		var message Message
		if err := wsjson.Read(context.Background(), conn.conn, &message); err != nil {
			_ = conn.close(err)
			return
		}
		log := conn.d.Log.With(slog.Any("message", message))
		if message.ID == uuid.Nil {
			log.Warn("received message without an ID", slog.Any("message", message))
			continue
		}
		switch message.Type {
		case MessageTypeCredentials:
			if message.From != "Server" {
				log.Warn("received credentials from non-Server", slog.Any("message", message))
				continue
			}
			var credentials nethernet.Credentials
			if err := json.Unmarshal([]byte(message.Data), &credentials); err != nil {
				log.Error("error decoding credentials", slog.Any("error", err))
				continue
			}
			closeChan := conn.credentials.Load() == nil
			conn.credentials.Store(&credentials)
			if closeChan {
				close(conn.credentialsReceived)
			}
		case MessageTypeSignal:
			signal := &nethernet.Signal{}
			if err := signal.UnmarshalText([]byte(message.Data)); err != nil {
				log.Error("error decoding signal", slog.Any("error", err))
				continue
			}
			signal.NetworkID = message.From

			conn.notifier.Signal(signal)
		case MessageTypeError:
			err := &Error{}
			if err2 := json.Unmarshal([]byte(message.Data), err); err2 != nil {
				log.Error("error decoding error", slog.Any("error", err2))
				continue
			}
			log.Debug("received error", slog.Any("message", message))
			conn.complete(message.ID, err)
		case MessageTypeDelivered:
			var status MessageStatus
			if err := json.Unmarshal([]byte(message.Data), &status); err != nil {
				log.Error("error decoding message status", slog.Any("message", message), slog.Any("error", err))
				continue
			}
			if status.DeliveredOn.IsZero() {
				log.Warn("delivery time is not included in message data")
			}
			conn.complete(message.ID, nil)
		case MessageTypeAccepted:
			continue
		default:
			log.Warn("received message for unknown type")
		}
	}
}

// write encodes the given Message and sends it over the WebSocket connection. It uses a background context
// to avoid issues with context cancellation affecting the connection. An error may be returned if the message
// could not be sent.
func (conn *Conn) write(message Message) error {
	return wsjson.Write(context.Background(), conn.conn, message)
}
