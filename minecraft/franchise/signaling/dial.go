package signaling

import (
	"context"
	"fmt"
	"github.com/coder/websocket"
	"github.com/df-mc/go-nethernet"
	"github.com/df-mc/go-playfab"
	"github.com/sandertv/gophertunnel/minecraft/auth/xal"
	"github.com/sandertv/gophertunnel/minecraft/franchise"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"golang.org/x/oauth2"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
)

// Dialer provides methods and fields to establish a Conn to a signaling service.
// It allows specifying options for the connection and handles various authentication
// and environment configuration.
type Dialer struct {
	// Options specifies the options for dialing the signaling service over
	// a WebSocket connection. If nil, a new *websocket.DialOptions will be
	// created. Note that the [websocket.DialOptions.HTTPClient] and its Transport
	// will be overridden with a [franchise.Transport] for authorization.
	Options *websocket.DialOptions

	// NetworkID specifies a unique ID for the network. If set to zero, a random
	// value will be automatically set from [rand.Uint64]. It is included in the URI
	// for establishing a WebSocket connection.
	NetworkID uint64

	// Log is used to logging messages at various levels. If nil, the default
	// [slog.Logger] will be set from [slog.Default].
	Log *slog.Logger
}

// DialContext establishes a Conn to the signaling service using the [oauth2.TokenSource] for
// authentication and authorization with franchise services. It obtains the necessary [franchise.Discovery]
// and [Environment] needed, then calls DialWithIdentityAndEnvironment internally. It is the
// method that is typically used when no configuration of identity and environment is required.
func (d Dialer) DialContext(ctx context.Context, src oauth2.TokenSource) (*Conn, error) {
	discovery, err := franchise.Discover(protocol.CurrentVersion)
	if err != nil {
		return nil, fmt.Errorf("obtain discovery: %w", err)
	}
	a := new(franchise.AuthorizationEnvironment)
	if err := discovery.Environment(a, franchise.EnvironmentTypeProduction); err != nil {
		return nil, fmt.Errorf("obtain environment for %s: %w", a.EnvironmentName(), err)
	}
	s := new(Environment)
	if err := discovery.Environment(s, franchise.EnvironmentTypeProduction); err != nil {
		return nil, fmt.Errorf("obtain environment for %s: %w", s.EnvironmentName(), err)
	}

	return d.DialWithIdentityAndEnvironment(ctx, franchise.PlayFabIdentityProvider{
		Environment: a,
		IdentityProvider: playfab.XBLIdentityProvider{
			TokenSource: xal.RefreshTokenSource(src, playfab.RelyingParty),
		},
	}, s)
}

// DialWithIdentityAndEnvironment establishes a Conn to the signaling service using the [franchise.IdentityProvider]
// for authorization and the [Environment] for creating the URI of an internal WebSocket connection. It appends 'ws/v1.0/signaling'
// with the NetworkID to the service URI from the Environment. It sets up necessary options and logging if not provided, and
// dials a [websocket.Conn] using [websocket.Dial]. The [context.Context] may be used to cancel the connection if necessary as
// soon as possible.
func (d Dialer) DialWithIdentityAndEnvironment(ctx context.Context, i franchise.IdentityProvider, env *Environment) (*Conn, error) {
	if d.Options == nil {
		d.Options = &websocket.DialOptions{}
	}
	if d.Options.HTTPClient == nil {
		d.Options.HTTPClient = &http.Client{}
	}
	if d.NetworkID == 0 {
		d.NetworkID = rand.Uint64()
	}
	if d.Log == nil {
		d.Log = slog.Default()
	}

	var (
		hasTransport bool
		base         = d.Options.HTTPClient.Transport
	)
	if base != nil {
		_, hasTransport = base.(*franchise.Transport)
	}
	if !hasTransport {
		d.Options.HTTPClient.Transport = &franchise.Transport{
			IdentityProvider: i,
			Base:             base,
		}
	}

	u, err := url.Parse(env.ServiceURI)
	if err != nil {
		return nil, fmt.Errorf("parse service URI: %w", err)
	}

	c, _, err := websocket.Dial(ctx, u.JoinPath("/ws/v1.0/signaling", strconv.FormatUint(d.NetworkID, 10)).String(), d.Options)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		conn: c,
		d:    d,

		credentialsReceived: make(chan struct{}),

		closed: make(chan struct{}),

		notifiers: make(map[uint32]nethernet.Notifier),
	}
	go conn.read()
	return conn, nil
}
