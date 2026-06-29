package minecraft

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/df-mc/go-xsapi/v2/xal"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
)

func TestDialAuthContextPropagatesHTTPClient(t *testing.T) {
	t.Parallel()

	client := &http.Client{}
	ctx := auth.WithContextClient(context.Background(), client)

	if got, _ := ctx.Value(xal.HTTPClient).(*http.Client); got != client {
		t.Fatalf("xal.HTTPClient = %p, want %p", got, client)
	}
	if got, _ := ctx.Value(oauth2.HTTPClient).(*http.Client); got != client {
		t.Fatalf("oauth2.HTTPClient = %p, want %p", got, client)
	}
}

func TestDialAuthContextDoesNotOverwriteHTTPClient(t *testing.T) {
	t.Parallel()

	existingXAL := &http.Client{}
	existingOAuth := &http.Client{}
	replacement := &http.Client{}
	ctx := context.WithValue(context.Background(), xal.HTTPClient, existingXAL)
	ctx = context.WithValue(ctx, oauth2.HTTPClient, existingOAuth)
	ctx = auth.WithContextClient(ctx, replacement)

	if got, _ := ctx.Value(xal.HTTPClient).(*http.Client); got != existingXAL {
		t.Fatalf("xal.HTTPClient = %p, want existing %p", got, existingXAL)
	}
	if got, _ := ctx.Value(oauth2.HTTPClient).(*http.Client); got != existingOAuth {
		t.Fatalf("oauth2.HTTPClient = %p, want existing %p", got, existingOAuth)
	}
}

func TestDialContextWithMultiplayerTokenSourceSkipsLegacySessionSetup(t *testing.T) {
	cache := auth.AndroidConfig.NewTokenCache()
	ctx := auth.WithXBLTokenCache(context.Background(), cache)
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("stop before auth discovery")
	})}

	_, err := Dialer{
		HTTPClient:  client,
		TokenSource: dialTestMultiplayerTokenSource{},
	}.DialContext(ctx, "unused", "example.com:19132")
	if err == nil {
		t.Fatal("expected DialContext to fail before network dial")
	}
	if cache.Session() != nil {
		t.Fatal("DialContext created a legacy XBL session for a MultiplayerTokenSource")
	}
}

func TestDialContextDialsOriginalAddressWithoutPinging(t *testing.T) {
	const networkID = "test-pong-port"
	dialErr := errors.New("stop after address capture")
	var dialAddress string
	pingCalled := false

	RegisterNetwork(networkID, func(*slog.Logger) Network {
		return dialTestNetwork{
			dial: func(_ context.Context, address string) (net.Conn, error) {
				dialAddress = address
				return nil, dialErr
			},
			ping: func(context.Context, string) ([]byte, error) {
				pingCalled = true
				return []byte("MCPE;InsaneSMP;800;1.21.80;0;100;123;World;Survival;1;25565;19133;"), nil
			},
		}
	})
	t.Cleanup(func() {
		UnregisterNetwork(networkID)
	})

	_, err := Dialer{}.DialContext(context.Background(), networkID, "insanesmp.net:19132")
	if !errors.Is(err, dialErr) {
		t.Fatalf("DialContext error = %v, want %v", err, dialErr)
	}
	if dialAddress != "insanesmp.net:19132" {
		t.Fatalf("dial address = %q, want insanesmp.net:19132", dialAddress)
	}
	if pingCalled {
		t.Fatal("DialContext called PingContext before dialing")
	}
}

func TestDialContextNetworkUsesExplicitNetwork(t *testing.T) {
	t.Parallel()

	dialErr := errors.New("stop after explicit network dial")
	ctx := context.WithValue(context.Background(), testContextKey{}, "explicit")
	network := dialTestNetwork{
		dial: func(gotCtx context.Context, address string) (net.Conn, error) {
			if gotCtx != ctx {
				t.Fatal("DialContextNetwork did not pass caller context to network")
			}
			if address != "nethernet-id" {
				t.Fatalf("network address = %q, want nethernet-id", address)
			}
			return nil, dialErr
		},
	}

	_, err := Dialer{}.DialContextNetwork(ctx, network, "nethernet-id")
	if !errors.Is(err, dialErr) {
		t.Fatalf("DialContextNetwork error = %v, want %v", err, dialErr)
	}
}

type dialTestNetwork struct {
	dial func(context.Context, string) (net.Conn, error)
	ping func(context.Context, string) ([]byte, error)
}

func (n dialTestNetwork) DialContext(ctx context.Context, address string) (net.Conn, error) {
	return n.dial(ctx, address)
}

func (n dialTestNetwork) PingContext(ctx context.Context, address string) ([]byte, error) {
	if n.ping != nil {
		return n.ping(ctx, address)
	}
	return nil, errors.New("not implemented")
}

func (dialTestNetwork) Listen(string) (NetworkListener, error) {
	return nil, errors.New("not implemented")
}

type dialTestMultiplayerTokenSource struct{}

func (dialTestMultiplayerTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "live", Expiry: time.Now().Add(time.Hour)}, nil
}

func (dialTestMultiplayerTokenSource) MultiplayerToken(context.Context, *ecdsa.PublicKey) (string, error) {
	return "", errors.New("unexpected multiplayer token request")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type testContextKey struct{}
