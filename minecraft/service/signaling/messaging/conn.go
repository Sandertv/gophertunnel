package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/creachadair/jrpc2"
	"github.com/df-mc/go-nethernet"
	"github.com/google/uuid"
)

// Conn implements a [nethernet.Signaling] over a JSON-RPC communication channel over WebSocket connection.
type Conn struct {
	// conn is the underlying websocket connection with signaling service.
	conn   *websocket.Conn
	client *jrpc2.Client
	// d is the Dialer used to dial this Conn.
	d Dialer
	// pmid is the Player Messaging ID extracted from JWT claims.
	pmid uuid.UUID

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

	// credentials is the last known credentials received from the server.
	// This is cached until it expires to prevent excessive call.
	credentials *nethernet.Credentials
	// credentialsExpiry is the expiry of the credentials, estimated from [nethernet.Credentials.ExpirationInSeconds].
	credentialsExpiry time.Time
	// credentialsMu guards credentials and credentialsExpiry from concurrent read/write access.
	credentialsMu sync.Mutex
}

// Signal sends a [nethernet.Signal] to a network.
func (conn *Conn) Signal(ctx context.Context, signal *nethernet.Signal) error {
	messagingID, err := uuid.Parse(signal.NetworkID)
	if err != nil {
		return fmt.Errorf("signaling/messaging: recipient ID must be a UUID: %q", signal.NetworkID)
	}
	id := uuid.New()
	if err := conn.send(ctx, id, map[string]any{
		"params": map[string]any{
			"netherNetId": conn.d.NetworkID,
			"message":     signal.String(),
		},
		"jsonrpc": "2.0",
		"method":  MethodSignalingWebRTC,
	}, messagingID); err != nil {
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

// send submits the JSON-RPC message encapsulated in an envelope that contains the recipient ID.
func (conn *Conn) send(ctx context.Context, id uuid.UUID, inner any, messagingID uuid.UUID) error {
	data, err := json.Marshal(inner)
	if err != nil {
		return fmt.Errorf("encode inner message: %w", err)
	}
	_, err = conn.client.Call(ctx, MethodSignalingSendMessage, map[string]any{
		"toPlayerId": messagingID,
		"messageId":  id,
		"message":    string(data),
	})
	if err != nil {
		return fmt.Errorf("call %q: %w", MethodSignalingSendMessage, err)
	}
	return nil
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

// expect registers interest in the completion of the outbound signaling
// message identified by id.
// The returned channel is resolved by [Conn.complete] when the matching
// delivery notification is received.
func (conn *Conn) expect(id uuid.UUID) <-chan error {
	c := make(chan error)
	conn.expectedMu.Lock()
	conn.expected[id] = c
	conn.expectedMu.Unlock()
	return c
}

// release stops tracking the outbound signaling message identified by id
// and closes its expectation channel if it is still registered.
// It is typically deferred after [Conn.expect] once waiting is no longer
// needed.
func (conn *Conn) release(id uuid.UUID) {
	conn.expectedMu.Lock()
	ch, ok := conn.expected[id]
	if ok {
		close(ch)
	}
	delete(conn.expected, id)
	conn.expectedMu.Unlock()
}

// complete resolves the expectation registered for the outbound signaling
// message identified by id.
// It is called when the matching JSON-RPC callback indicates that the
// remote side has acknowledged the message, or when message processing
// needs to report an error for that ID.
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
	conn.credentialsMu.Lock()
	defer conn.credentialsMu.Unlock()

	if conn.credentials != nil && conn.credentialsExpiry.Before(time.Now()) {
		return conn.credentials, nil
	}

	var credentials nethernet.Credentials
	if err := conn.client.CallResult(ctx, MethodSignalingCredentials, map[string]any{}, &credentials); err != nil {
		return nil, fmt.Errorf("call %q: %w", MethodSignalingCredentials, err)
	}

	conn.credentials = &credentials
	conn.credentialsExpiry = time.Now().Add(time.Duration(credentials.ExpirationInSeconds) * time.Second)

	return conn.credentials, nil
}

// PongData ...
func (conn *Conn) PongData(b []byte) {
}

// NetworkID returns the network ID of the Conn. It may be specified from [Dialer.NetworkID], otherwise a random
// value will be automatically set from [rand.Uint64] in set up during [Dialer.DialContext]. It is utilized by
// [nethernet.Listener] and [nethernet.Dialer] to obtain its local network ID to listen.
func (conn *Conn) NetworkID() string {
	return conn.pmid.String()
}

// PlayerMessagingID returns the player messaging ID of the current authenticated user.
func (conn *Conn) PlayerMessagingID() uuid.UUID {
	return conn.pmid
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

// close cancels the background context of the Conn and closes the underlying WebSocket connection.
func (conn *Conn) close(cause error) (err error) {
	conn.once.Do(func() {
		conn.d.Log.Debug("closing connection", slog.Any("cause", cause))

		conn.notifiersMu.Lock()
		for i := range conn.notifiers {
			conn.stop(i)
		}
		conn.notifiersMu.Unlock()

		conn.cancel(cause)
		err = conn.client.Close()
	})
	return err
}

// handleCallback handles an JSON-RPC request method called by the server.
// It is used as the [jrpc2.ClientOptions.OnCallback] handler in the client.
func (conn *Conn) handleCallback(ctx context.Context, request *jrpc2.Request) (any, error) {
	switch request.Method() {
	case MethodSystemPong:
		return nil, nil
	case MethodSignalingReceiveMessage:
		return conn.handleMessage(ctx, request)
	default:
		return nil, fmt.Errorf("received message for unknown method: %q", request.Method())
	}
}

// handleMessage handles a message received from the server.
// It decodes every batch message included in the given request
// and calls handleInnerMessage.
func (conn *Conn) handleMessage(ctx context.Context, request *jrpc2.Request) (v any, err error) {
	var params []*envelope
	if err := request.UnmarshalParams(&params); err != nil {
		return nil, fmt.Errorf("decode parameters: %w", err)
	}
	for _, e := range params {
		if e == nil {
			err = errors.Join(err, errors.New("signaling/messaging: nil envelope"))
			continue
		}
		if err2 := conn.handleInnerMessage(ctx, e); err2 != nil {
			err = errors.Join(err, err2)
		}
	}
	return nil, err
}

// envelope wraps a JSON-RPC message delivered through Player Messaging.
type envelope struct {
	// From identifies the Player Messaging ID of the sender.
	From uuid.UUID
	// Message is the inner JSON-RPC request sent by the remote network.
	Message *jrpc2.ParsedRequest
	// ID is the unique message ID associated to this message.
	// It is used to track the delivery status of a message.
	ID uuid.UUID `json:"Id"`
}

// UnmarshalJSON decodes an envelope whose Message field is encoded on the
// wire as a JSON string containing a nested JSON-RPC request.
func (m *envelope) UnmarshalJSON(b []byte) error {
	type Alias envelope
	data := struct {
		*Alias
		Message string
	}{Alias: (*Alias)(m)}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(data.Message), &m.Message); err != nil {
		return fmt.Errorf("decode message: %w", err)
	}
	return nil
}

// handleInnerMessage handles an inner message encapsulated in the given envelope.
func (conn *Conn) handleInnerMessage(ctx context.Context, envelope *envelope) error {
	switch envelope.Message.Method {
	case MethodSignalingDeliveryNotification:
		var params struct {
			MessageID uuid.UUID `json:"messageId"`
		}
		if err := json.Unmarshal(envelope.Message.Params, &params); err != nil {
			return fmt.Errorf("decode request parameters: %w", err)
		}
		if params.MessageID == uuid.Nil {
			return errors.New("invalid request parameters")
		}
		conn.complete(params.MessageID, nil)
		return nil
	case MethodSignalingWebRTC:
		var params struct {
			NetherNetID string `json:"netherNetId"`
			Message     string `json:"message"`
		}
		if err := json.Unmarshal(envelope.Message.Params, &params); err != nil {
			return fmt.Errorf("decode request parameters: %w", err)
		}
		signal := &nethernet.Signal{NetworkID: envelope.From.String()}
		if err := signal.UnmarshalText([]byte(params.Message)); err != nil {
			return fmt.Errorf("decode signal: %w", err)
		}
		conn.notifiersMu.Lock()
		for _, n := range conn.notifiers {
			select {
			case n.in <- signal:
			default:
				// Drop when notifier is backed up to avoid deadlocks and keep packet processing moving.
				conn.d.Log.Debug("dropping signal due to notifier being backed up", slog.String("signal", signal.String()))
			}
		}
		conn.notifiersMu.Unlock()

		if err := conn.send(ctx, uuid.New(), map[string]any{
			"params": map[string]any{
				"messageId": envelope.ID,
			},
			"jsonrpc": "2.0",
			"method":  MethodSignalingDeliveryNotification,
		}, envelope.From); err != nil {
			return fmt.Errorf("acknowledge message: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown inner request method: %q", envelope.Message.Method)
	}
}

// ping starts calling [MethodSystemPing] at 50 seconds interval.
// If the ping failed, it closes the Conn immediately with the cause.
func (conn *Conn) ping() {
	ticker := time.NewTicker(time.Second * 50)
	defer ticker.Stop()

	for {
		select {
		case <-conn.ctx.Done():
			return
		case <-ticker.C:
			if _, err := conn.client.Call(conn.ctx, MethodSystemPing, map[string]any{}); err != nil {
				conn.d.Log.Error("error pinging", slog.Any("error", err))
				_ = conn.close(fmt.Errorf("call %q: %w", MethodSystemPing, err))
				return
			}
		}
	}
}

const (
	// MethodSystemPing is the JSON-RPC method name used by the client to
	// ping the server and keep the connection alive.
	MethodSystemPing = "System_Ping_v1_0"
	// MethodSystemPong is called by the server in response to [MethodSystemPing] to ping the client.
	// The client must respond with a nil value in order to complete the request.
	MethodSystemPong = "System_Pong_v1_0"

	// MethodSignalingCredentials is the JSON-RPC method name used by the
	// client to request [nethernet.Credentials] for Microsoft's STUN/TURN
	// servers.
	MethodSignalingCredentials = "Signaling_TurnAuth_v1_0"
	// MethodSignalingReceiveMessage is the JSON-RPC method name used by
	// the server to deliver one or more envelopes received from remote peers.
	MethodSignalingReceiveMessage = "Signaling_ReceiveMessage_v1_0"
	// MethodSignalingSendMessage is the JSON-RPC method name used by the
	// client to send an inner signaling message to a remote peer.
	MethodSignalingSendMessage = "Signaling_SendClientMessage_v1_0"

	// MethodSignalingDeliveryNotification is the JSON-RPC method name used
	// by an inner message to acknowledge receipt of an earlier message identified
	// by its outer envelope ID.
	MethodSignalingDeliveryNotification = "Signaling_DeliveryNotification_V1_0"
	// MethodSignalingWebRTC is the JSON-RPC method name used by an inner
	// message that carries a NetherNet WebRTC signaling payload.
	MethodSignalingWebRTC = "Signaling_WebRtc_v1_0"
)
