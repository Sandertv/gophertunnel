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

func (d Dialer) DialContext(ctx context.Context, src franchise.IdentityProvider, env *Environment) (*Conn, error) {
	if d.Options == nil {
		d.Options = &websocket.DialOptions{}
	}
	if d.Options.HTTPClient == nil {
		d.Options.HTTPClient = &http.Client{}
	}
	if d.Options.HTTPHeader == nil {
		d.Options.HTTPHeader = make(http.Header) // TODO(lactyy): Move to *franchise.Transport
	}
	if d.NetworkID == 0 {
		d.NetworkID = rand.Uint64()
	}
	if d.Log == nil {
		d.Log = slog.Default()
	}
	/*var hasTransport bool
	if base := d.Options.HTTPClient.Transport; base != nil {
		_, hasTransport = base.(*franchise.Transport)
	}
	if !hasTransport {
		d.Options.HTTPClient.Transport = &franchise.Transport{
			Source: src,
			Base:   d.Options.HTTPClient.Transport,
		}
	}*/

	// TODO(lactyy): Move to *franchise.Transport
	conf, err := src.TokenConfig()
	if err != nil {
		return nil, fmt.Errorf("request token config: %w", err)
	}
	t, err := conf.Token()
	if err != nil {
		return nil, fmt.Errorf("request token: %w", err)
	}
	d.Options.HTTPHeader.Set("Authorization", t.AuthorizationHeader)

	u, err := url.Parse(env.ServiceURI)
	if err != nil {
		return nil, fmt.Errorf("parse service URI: %w", err)
	}

	c, _, err := websocket.Dial(ctx, u.JoinPath("/ws/v1.0/signaling/", strconv.FormatUint(d.NetworkID, 10)).String(), d.Options)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		conn:    c,
		d:       d,
		signals: make(chan *nethernet.Signal),
		ready:   make(chan struct{}),
	}
	var cancel context.CancelCauseFunc
	conn.ctx, cancel = context.WithCancelCause(context.Background())

	go conn.read(cancel)
	go conn.ping()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-conn.ready:
		return conn, nil
	}
}
