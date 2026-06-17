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
	// closeCredentialsReceived ensures that the credentialsReceived channel is closed only once.
	closeCredentialsReceived sync.Once

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
	message := Message{
		Type: MessageTypeSignal,
		To:   signal.NetworkID,
		Data: signal.String(),
		ID:   id,
	}
	if signal.Type != nethernet.SignalTypeOffer || conn.d.IgnoreDeliveryNotification {
		return conn.write(ctx, message)
	}

	ch := conn.pending.Add(id)
	defer conn.pending.Remove(id)

	if err := conn.write(ctx, message); err != nil {
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

// Notify returns a channel that receives incoming NetherNet signals.
//
// The returned stop function unregisters and closes the channel.
func (conn *Conn) Notify() (<-chan *nethernet.Signal, func()) {
	return conn.notifier.Register()
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

// close cancels the background context of the Conn and closes the
// underlying WebSocket connection. Any blocking methods on Conn
// will return the given cause as the error.
func (conn *Conn) close(cause error) (err error) {
	conn.once.Do(func() {
		conn.d.Log.Debug("closing connection", slog.Any("cause", cause))

		conn.cancel(cause)
		_ = conn.notifier.Close()

		err = conn.conn.Close(websocket.StatusNormalClosure, "")
	})
	return err
}

// ping starts periodically sending ping messages at the specified interval.
// On failure, it closes the Conn immediately with the cause.
func (conn *Conn) ping(frequency time.Duration) {
	if frequency <= 0 {
		frequency = DefaultPingFrequency
	}
	ticker := time.NewTicker(frequency)
	defer ticker.Stop()

	for {
		select {
		case <-conn.ctx.Done():
			return
		case <-ticker.C:
			if err := conn.write(conn.ctx, Message{
				Type: MessageTypePing,
			}); err != nil {
				_ = conn.close(fmt.Errorf("background: write ping: %w", err))
				return
			}
		}
	}
}

// read continuously reads messages from the WebSocket connection and handles them.
// It also sends a Message of MessageTypePing at 15 seconds intervals to keep the
// Conn alive. It goes as a background goroutine of the Conn and handles different
// types of messages: credentials, signals, and errors. It closes the Conn if it
// encounters an error or when the Conn is closed.
func (conn *Conn) read() {
	for {
		var message Message
		if err := wsjson.Read(context.Background(), conn.conn, &message); err != nil {
			_ = conn.close(err)
			return
		}
		conn.handleMessage(message)
	}
}

// handleMessage handles a message received from the WebSocket signaling service.
func (conn *Conn) handleMessage(message Message) {
	log := conn.d.Log.With(slog.Any("message", message))
	switch message.Type {
	case MessageTypeCredentials:
		if message.From != "Server" {
			log.Warn("received credentials from non-Server", slog.Any("message", message))
			return
		}
		var credentials nethernet.Credentials
		if err := json.Unmarshal([]byte(message.Data), &credentials); err != nil {
			log.Error("error decoding credentials", slog.Any("error", err))
			return
		}
		conn.credentials.Store(&credentials)
		conn.closeCredentialsReceived.Do(func() {
			close(conn.credentialsReceived)
		})
	case MessageTypeSignal:
		signal := &nethernet.Signal{}
		if err := signal.UnmarshalText([]byte(message.Data)); err != nil {
			log.Error("error decoding signal", slog.Any("error", err))
			return
		}
		signal.NetworkID = message.From
		conn.notifier.Signal(signal)
	case MessageTypeError:
		if message.ID == uuid.Nil {
			log.Warn("received message without an ID", slog.Any("message", message))
			return
		}
		err := &Error{}
		if err2 := json.Unmarshal([]byte(message.Data), err); err2 != nil {
			log.Error("error decoding error", slog.Any("error", err2))
			return
		}
		log.Debug("received error", slog.Any("message", message))
		conn.complete(message.ID, err)
	case MessageTypeDelivered:
		if conn.d.IgnoreDeliveryNotification {
			return
		}
		if message.ID == uuid.Nil {
			log.Warn("received message without an ID", slog.Any("message", message))
			return
		}
		var status MessageStatus
		if err := json.Unmarshal([]byte(message.Data), &status); err != nil {
			log.Error("error decoding message status", slog.Any("message", message), slog.Any("error", err))
			return
		}
		if status.DeliveredOn.IsZero() {
			log.Warn("delivery time is not included in message data")
		}
		conn.complete(message.ID, nil)
	case MessageTypeAccepted:
		if message.ID == uuid.Nil {
			log.Warn("received message without an ID", slog.Any("message", message))
		}
	default:
		log.Warn("received message for unknown type")
	}
}

// write encodes the given Message and sends it over the WebSocket connection.
// An error may be returned if the message could not be sent.
func (conn *Conn) write(ctx context.Context, message Message) error {
	return wsjson.Write(ctx, conn.conn, message)
}
