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
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/service"
)

type Dialer struct {
	Environment *Environment
	HTTPClient  *http.Client
	Log         *slog.Logger
	NetworkID   string
}

// Dial connects to the signaling service.
func (d Dialer) Dial(src service.TokenSource) (*Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	return d.DialContext(ctx, src)
}

// DialContext connects to the signaling service using the provided [service.TokenSource] for authorization.
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

		notifiers: make(map[uint32]notifier),
		expected:  make(map[uuid.UUID]chan error),
	}
	conn.ctx, conn.cancel = context.WithCancelCause(context.Background())
	go conn.read()
	return conn, nil
}

type Environment struct {
	ServiceURI *url.URL `json:"serviceUri"`
	TurnURI    string   `json:"turnUri"`
	StunURI    string   `json:"stunUri"`
}

func (e *Environment) ServiceName() string {
	return "signaling"
}

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
