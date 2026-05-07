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

	// notifyCount counts the total notifiers registered for the Conn.
	// It is increased when a [nethernet.Notifier] is registered and used
	// and used as the ID for [nethernet.Notifier] for storing them to notifiers.
	notifyCount uint32
	// notifiers is a map whose keys are the IDs and whose values are [nethernet.Notifier]
	// registered for use in the Conn from [Conn.Notify] when dialing or listening.
	notifiers map[uint32]notifier
	// notifiersMu guards notifyCount and notifiers to ensure concurrent safety on [Conn.Notify].
	notifiersMu sync.Mutex

	// expected is a map whose keys are IDs associated with messages sent to the remote network, and whose
	// values are channels that may be used to signal an error when a message of MessageTypeError or
	// MessageTypeDelivered is received.
	expected map[uuid.UUID]chan error
	// expectedMu should be held when expected is in access for ensuring concurrent safety.
	expectedMu sync.Mutex
}

// Signal sends a [nethernet.Signal] to a network.
func (conn *Conn) Signal(ctx context.Context, signal *nethernet.Signal) error {
	id := uuid.New()
	if err := conn.write(Message{
		Type: MessageTypeSignal,
		To:   signal.NetworkID,
		Data: signal.String(),
		ID:   id,
	}); err != nil {
		return err
	}

	ch := conn.expect(id)
	defer conn.release(id)

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
	conn.notifiersMu.Lock()
	i := conn.notifyCount
	n := notifier{
		// Buffer notifications so packet handling never blocks under lock.
		in:   make(chan *nethernet.Signal, 64),
		out:  signals,
		stop: make(chan struct{}),
	}
	conn.notifiers[i] = n
	conn.notifyCount++
	conn.notifiersMu.Unlock()

	go func() {
		defer close(signals)
		for {
			select {
			case <-n.stop:
				return
			case sig, ok := <-n.in:
				if !ok {
					return
				}
				select {
				case <-n.stop:
					return
				case n.out <- sig:
				}
			}
		}
	}()

	return func() {
		conn.notifiersMu.Lock()
		conn.stop(i)
		conn.notifiersMu.Unlock()
	}
}

// stop stops notifying signals on the notifier with the corresponding ID. The ID
// is internally assigned for the notifier and contained in the stop function returned
// by [Conn.Notify]. It should not be called by anywhere else.
func (conn *Conn) stop(i uint32) {
	n, ok := conn.notifiers[i]
	if !ok {
		return
	}
	delete(conn.notifiers, i)
	close(n.stop)
	close(n.in)
}

// notifier holds a buffered input channel and a caller-provided output
// channel for relaying incoming signals to a [nethernet.Listener].
type notifier struct {
	in   chan *nethernet.Signal
	out  chan<- *nethernet.Signal
	stop chan struct{}
}

// expect registers interest in the completion of the outbound Message with
// the given ID and returns a channel that is completed by [Conn.complete].
// The channel receives nil when the signaling service reports
// [MessageTypeDelivered], or a non-nil error when it reports
// [MessageTypeError].
func (conn *Conn) expect(id uuid.UUID) <-chan error {
	c := make(chan error)
	conn.expectedMu.Lock()
	conn.expected[id] = c
	conn.expectedMu.Unlock()
	return c
}

// release stops tracking the outbound Message with the given ID and closes
// its expectation channel if it is still registered.
// It is typically deferred after [Conn.expect] once waiting for delivery is
// no longer needed.
func (conn *Conn) release(id uuid.UUID) {
	conn.expectedMu.Lock()
	ch, ok := conn.expected[id]
	if ok {
		close(ch)
	}
	delete(conn.expected, id)
	conn.expectedMu.Unlock()
}

// complete resolves the expectation registered for the outbound Message with
// the given ID.
// It is called when the signaling service sends a completion frame for that
// message ID, such as [MessageTypeDelivered] or [MessageTypeError].
func (conn *Conn) complete(id uuid.UUID, err error) {
	conn.expectedMu.Lock()
	ch, ok := conn.expected[id]
	if !ok {
		conn.expectedMu.Unlock()
		conn.d.Log.Warn("unexpected message ID", slog.Group("message",
			slog.String("id", id.String())))
		return
	}
	ch <- err
	conn.expectedMu.Unlock()
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

		conn.notifiersMu.Lock()
		for i := range conn.notifiers {
			conn.stop(i)
		}
		conn.notifiersMu.Unlock()

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

			conn.notifiersMu.Lock()
			for _, n := range conn.notifiers {
				select {
				case n.in <- signal:
				default:
					// Drop when notifier is backed up to avoid deadlocks and keep packet processing moving.
					log.Debug("dropping signal due to notifier being backed up", slog.String("signal", signal.String()))
				}
			}
			conn.notifiersMu.Unlock()
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
