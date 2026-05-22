package signaling

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/coder/websocket"
	"github.com/sandertv/gophertunnel/minecraft/service"
	"github.com/sandertv/gophertunnel/minecraft/service/signaling/internal"
)

// Dialer specifies options for connecting to the signaling service.
type Dialer struct {
	// Environment is the environment used for connecting to the signaling service.
	// If nil, it will be derived from [service.Default].
	Environment *Environment
	// HTTPClient is the HTTP client used during WebSocket handshake.
	HTTPClient *http.Client
	// Log is the logger used to log messages at various levels.
	// If nil, it will be set from [slog.Default].
	Log *slog.Logger
	// NetworkID specifies a unique ID for the network. If zero, a random value will
	// be automatically set from [rand.Uint64]. It is included in the URI for establishing
	// a WebSocket connection.
	NetworkID string
}

// Dial connects to the signaling service with a 15 seconds timeout.
func (d Dialer) Dial(src service.TokenSource) (*Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	return d.DialContext(ctx, src)
}

// DialContext connects to the signaling service using the provided [service.TokenSource] for authorization.
// The given [context.Context] is used to control the deadline for the WebSocket handshake.
func (d Dialer) DialContext(ctx context.Context, src service.TokenSource) (*Conn, error) {
	if d.Environment == nil {
		discovery, err := service.Default(ctx)
		if err != nil {
			return nil, fmt.Errorf("discover network services: %w", err)
		}
		d.Environment = new(Environment)
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
	requestURL := d.Environment.ServiceURI.JoinPath(
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

		notifier: internal.NewNotifier(d.Log),
		pending:  internal.NewPendingMap(),
	}
	conn.ctx, conn.cancel = context.WithCancelCause(context.Background())
	go conn.read()
	return conn, nil
}

// Environment represents an environment for the signaling service.
type Environment struct {
	// ServiceURI is the base endpoint URL for the signaling service.
	// The scheme is typically 'wss://'.
	ServiceURI *url.URL `json:"serviceUri"`
	// TurnURI is the 'turn://' URI for Microsoft's TURN server. NetherNet
	// connections use the TURN server embedded in
	// [github.com/df-mc/go-nethernet.Credentials]
	// for actual WebRTC negotiation, so the purpose of this field is unknown.
	TurnURI string `json:"turnUri"`
	// StunURI is the 'stun://' URI for Microsoft's STUN server. NetherNet
	// connections use the STUN server embedded in
	// [github.com/df-mc/go-nethernet.Credentials]
	// for actual WebRTC negotiation, so the purpose of this field is unknown.
	StunURI string `json:"stunUri"`
}

// ServiceName always returns 'signaling' as the name of the service environment.
// It implements [service.Environment] so it can be derived using [service.Discovery.Environment].
func (e *Environment) ServiceName() string {
	return "signaling"
}

// UnmarshalJSON decodes the ServiceURI field to string then parses as URL.
// Other URI fields such as TurnURI is not validated as it is not used in
// the actual WebRTC negotiation for NetherNet connections.
func (e *Environment) UnmarshalJSON(b []byte) (err error) {
	type Alias Environment
	data := struct {
		*Alias
		ServiceURI string `json:"serviceUri"`
	}{Alias: (*Alias)(e)}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	e.ServiceURI, err = url.Parse(data.ServiceURI)
	if err != nil {
		return fmt.Errorf("parse service URI: %w", err)
	}
	return nil
}
