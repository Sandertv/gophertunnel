package signaling

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/coder/websocket"
	"github.com/df-mc/go-nethernet"
	"github.com/sandertv/gophertunnel/minecraft/service"
	"github.com/sandertv/gophertunnel/minecraft/service/signaling/internal"
	"golang.org/x/oauth2"
)

// Dialer specifies options for connecting to the signaling service.
type Dialer struct {
	// Environment is the environment used for connecting to the signaling service.
	// If nil, it will be derived from [service.Default].
	Environment ConfigurationProvider
	// HTTPClient is the HTTP client used during WebSocket handshake.
	HTTPClient *http.Client
	// Log is the logger used to log messages at various levels.
	// If nil, it will be set from [slog.Default].
	Log *slog.Logger
	// NetworkID specifies a unique ID for the network. If zero, a random value will
	// be automatically set from [rand.Uint64]. It is included in the URI for establishing
	// a WebSocket connection.
	NetworkID string
	// IgnoreDeliveryNotification disables waiting for delivery confirmation after
	// sending a signal to a remote peer.
	//
	// By default, [Conn.Signal] blocks until the remote peer sends back a
	// delivery confirmation message to confirm receipt. Enabling this field
	// causes Signal to return as soon as the signaling service accepts the
	// message, without waiting for acknowledgement by the remote peer.
	//
	// The signaling service appears to be sending delivery confirmation only for
	// telemetry so this field is unlikely to affect the actual WebRTC negotiation.
	IgnoreDeliveryNotification bool
}

// Dial connects to the signaling service with a 15 seconds timeout.
func (d Dialer) Dial(src service.TokenSource) (*Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	return d.DialContext(ctx, src)
}

// DialContext connects to the signaling service using the provided [service.TokenSource] for authorization.
// The given [context.Context] is used to control the deadline for discovery, authorization, and the WebSocket
// handshake. The returned Conn may still need to wait for initial ICE credentials in [Conn.Credentials].
func (d Dialer) DialContext(ctx context.Context, src service.TokenSource) (*Conn, error) {
	if d.HTTPClient == nil {
		d.HTTPClient = http.DefaultClient
	}
	if c, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); !ok || c == nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, d.HTTPClient)
	}
	if d.Environment == nil {
		discovery, err := service.Default(ctx)
		if err != nil {
			return nil, fmt.Errorf("discover network services: %w", err)
		}
		env := new(Environment)
		if err := discovery.Environment(env); err != nil {
			return nil, fmt.Errorf("resolve environment for %q: %w", env.ServiceName(), err)
		}
		d.Environment = env
	}
	if d.Log == nil {
		d.Log = slog.Default()
	}
	if d.NetworkID == "" {
		d.NetworkID = strconv.FormatUint(rand.Uint64(), 10)
	}

	cfg, err := d.Environment.Configuration(ctx, d.HTTPClient, src)
	if err != nil {
		return nil, fmt.Errorf("request configuration: %w", err)
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
	requestURL := cfg.ServiceURI.JoinPath(
		"/ws/v1.0/signaling",
		d.NetworkID,
	)
	c, _, err := websocket.Dial(ctx, requestURL.String(), opts)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		conn: c,
		d:    d,

		credentialsReceived: make(chan struct{}),

		notifiers: make(map[uint32]nethernet.Notifier),
		pending:   internal.NewPendingMap(),
	}
	conn.ctx, conn.cancel = context.WithCancelCause(context.Background())
	go conn.read()
	go conn.ping(cfg.PingFrequency)
	return conn, nil
}
