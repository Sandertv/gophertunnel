package signaling

import (
	"context"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/franchise"
	"github.com/sandertv/gophertunnel/minecraft/nethernet"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
	"strconv"
)

type Dialer struct {
	Options   *websocket.DialOptions
	NetworkID uint64
	Log       *slog.Logger
}

func (d Dialer) DialContext(ctx context.Context, i franchise.IdentityProvider, env *Environment) (*Conn, error) {
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

	c, _, err := websocket.Dial(ctx, u.JoinPath("/ws/v1.0/signaling/", strconv.FormatUint(d.NetworkID, 10)).String(), d.Options)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		conn: c,
		d:    d,

		credentialsReceived: make(chan struct{}),

		closed: make(chan struct{}),

		signals: make(chan *nethernet.Signal),
	}
	go conn.read()
	go conn.ping()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-conn.credentialsReceived:
		return conn, nil
	}
}
