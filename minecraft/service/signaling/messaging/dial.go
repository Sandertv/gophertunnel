package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/coder/websocket"
	"github.com/creachadair/jrpc2"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/service"
	"github.com/sandertv/gophertunnel/minecraft/service/signaling"
)

// Dialer specifies options for connecting to the messaging service.
type Dialer struct {
	// Environment is the Environment of the 'signaling' service used to connect to messaging service.
	// If nil, it will be automatically resolved from the discovery data returned from [service.Default].
	Environment *signaling.Environment
	// HTTPClient is the HTTP client used for WebSocket handshake messages and [Environment] discovery.
	HTTPClient *http.Client
	// Log is the logger used to log messages at various levels.
	Log *slog.Logger
	// NetworkID specifies a unique ID for the NetherNet network. If zero, a random value will
	// be automatically set from [rand.Uint64]. When listening on friend worlds, this value
	// must match the NetworkID advertised in [p2p.Connection.NetherNetID] in order to successfully
	// negotiate with vanilla clients.
	NetworkID string
}

// Dial connects to the messaging service with a 15 seconds timeout.
func (d Dialer) Dial(src service.TokenSource) (*Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	return d.DialContext(ctx, src)
}

// DialContext connects to the messaging service using the provided [service.TokenSource] for authorization.
func (d Dialer) DialContext(ctx context.Context, src service.TokenSource) (*Conn, error) {
	if d.Environment == nil {
		discovery, err := service.Default(ctx)
		if err != nil {
			return nil, fmt.Errorf("discover network services: %w", err)
		}
		d.Environment = new(signaling.Environment)
		if err := discovery.Environment(d.Environment); err != nil {
			return nil, fmt.Errorf("resolve environment for %q: %w", d.Environment.ServiceName(), err)
		}
	}
	if d.HTTPClient == nil {
		d.HTTPClient = http.DefaultClient
	}
	if d.Log == nil {
		d.Log = slog.Default()
	}
	if d.NetworkID == "" {
		d.NetworkID = strconv.FormatUint(rand.Uint64(), 10)
	}

	token, err := src.ServiceToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("request service token: %w", err)
	}

	opts := &websocket.DialOptions{
		HTTPHeader: make(http.Header),
		HTTPClient: d.HTTPClient,
	}
	opts.HTTPHeader.Set("Authorization", token.AuthorizationHeader)
	requestURL := d.Environment.ServiceURI.JoinPath("/ws/v1.0/messaging/connect")
	c, _, err := websocket.Dial(ctx, requestURL.String(), opts)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		conn: c,
		d:    d,
		pmid: token.Claims.PlayerMessagingID,

		notifiers: make(map[uint32]notifier),
		expected:  make(map[uuid.UUID]chan error),
	}
	conn.ctx, conn.cancel = context.WithCancelCause(context.Background())
	conn.client = jrpc2.NewClient(&websocketChannel{c}, &jrpc2.ClientOptions{
		Logger: func(text string) {
			d.Log.Debug(text, slog.String("src", "jrpc2"))
		},
		OnNotify: func(request *jrpc2.Request) {
			d.Log.Warn("notification received", slog.Group("request",
				slog.String("id", request.ID()),
				slog.String("params", request.ParamString()),
			))
		},
		OnCallback: func(ctx context.Context, request *jrpc2.Request) (v any, err error) {
			defer func() {
				if err2 := recover(); err2 != nil {
					v, err = nil, nil
				}
			}()
			v, err = conn.handleCallback(ctx, request)
			if err != nil {
				// Returning non-nil error may cause the connection to stale
				// so we catch it here and always return nil instead.
				conn.d.Log.Error("error handling server message",
					slog.GroupAttrs("request",
						slog.String("id", request.ID()),
						slog.String("method", request.Method()),
						slog.String("params", request.ParamString()),
					),
					slog.Any("error", err),
				)
				return nil, nil
			}
			return v, nil
		},
	})
	go conn.ping()
	return conn, nil
}

// websocketChannel is an implementation of [channel.Channel] over [websocket.Conn].
// It is used to transmit JSON-RPC messages over a WebSocket connection in [jrpc2.Client].
type websocketChannel struct{ *websocket.Conn }

// Send writes the provided bytes to the WebSocket connection.
func (ch *websocketChannel) Send(b []byte) error {
	return ch.Write(context.Background(), websocket.MessageText, b)
}

// Recv blocks until a new data is received in the WebSocket connection
// and returns the payload.
func (ch *websocketChannel) Recv() ([]byte, error) {
	_, msg, err := ch.Read(context.Background())
	return msg, err
}

// Close closes the underlying WebSocket connection.
func (ch *websocketChannel) Close() error {
	return ch.Conn.Close(websocket.StatusNormalClosure, "")
}
