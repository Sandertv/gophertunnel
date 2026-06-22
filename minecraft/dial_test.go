package minecraft

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"testing"

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

func TestReadChainIdentityDataRejectsShortChain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
	}{
		{name: "empty", data: `{"chain":[]}`},
		{name: "one entry", data: `{"chain":["root"]}`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := readChainIdentityData([]byte(tt.data)); err == nil {
				t.Fatal("expected short chain error")
			}
		})
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
