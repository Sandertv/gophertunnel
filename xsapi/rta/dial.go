package rta

import (
	"context"
	"github.com/sandertv/gophertunnel/xsapi"
	"log/slog"
	"net/http"
	"nhooyr.io/websocket"
	"slices"
	"time"
)

// Dialer represents the options for establishing a Conn with real-time activity services with DialContext or Dial.
type Dialer struct {
	Options  *websocket.DialOptions
	ErrorLog *slog.Logger
}

// Dial calls DialContext with a 15 seconds timeout.
func (d Dialer) Dial(src xsapi.TokenSource) (*Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	return d.DialContext(ctx, src)
}

// DialContext establishes a connection with real-time activity service. A context.Context is used to control the
// scene real-timely. An authorization token may be used for configuring an HTTP header to Options. An error may be
// returned during the dial of websocket connection.
func (d Dialer) DialContext(ctx context.Context, src xsapi.TokenSource) (*Conn, error) {
	if d.ErrorLog == nil {
		d.ErrorLog = slog.Default()
	}
	if d.Options == nil {
		d.Options = &websocket.DialOptions{}
	}
	if !slices.Contains(d.Options.Subprotocols, subprotocol) {
		d.Options.Subprotocols = append(d.Options.Subprotocols, subprotocol)
	}
	if d.Options.HTTPHeader == nil {
		d.Options.HTTPHeader = make(http.Header)
	}

	if d.Options.HTTPClient == nil {
		d.Options.HTTPClient = &http.Client{}
	}
	var (
		hasTransport bool
		base         = d.Options.HTTPClient.Transport
	)
	if base != nil {
		_, hasTransport = base.(*xsapi.Transport)
	}
	if !hasTransport {
		d.Options.HTTPClient.Transport = &xsapi.Transport{
			Source: src,
			Base:   base,
		}
	}

	c, _, err := websocket.Dial(ctx, connectURL, d.Options)
	if err != nil {
		return nil, err
	}
	background, cancel := context.WithCancelCause(context.Background())
	conn := &Conn{
		conn:          c,
		log:           d.ErrorLog,
		ctx:           background,
		subscriptions: make(map[uint32]*Subscription),
	}
	for i := 0; i < cap(conn.expected); i++ {
		conn.expected[i] = make(map[uint32]chan<- *handshake)
	}
	go conn.read(cancel)
	return conn, nil
}

const (
	// connectURL is the URL used to establish a websocket connection with real-time activity services. It is
	// generally present at websocket.Dial with other websocket.DialOptions, specifically along with subprotocol.
	connectURL = "wss://rta.xboxlive.com/connect"
	// subprotocol is the subprotocol used with connectURL, to establish a websocket connection.
	subprotocol = "rta.xboxlive.com.V2"
)
