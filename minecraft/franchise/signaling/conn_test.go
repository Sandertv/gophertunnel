package signaling

import (
	"context"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/franchise"
	"github.com/sandertv/gophertunnel/minecraft/franchise/internal/test"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/xsapi/xal"
	"testing"
	"time"
)

func TestDial(t *testing.T) {
	discovery, err := franchise.Discover(protocol.CurrentVersion)
	if err != nil {
		t.Fatalf("error retrieving discovery: %s", err)
	}

	a := new(franchise.AuthorizationEnvironment)
	if err := discovery.Environment(a, franchise.EnvironmentTypeProduction); err != nil {
		t.Fatalf("error reading environment for authorization: %s", err)
	}
	s := new(Environment)
	if err := discovery.Environment(s, franchise.EnvironmentTypeProduction); err != nil {
		t.Fatalf("error reading environment for signaling: %s", err)
	}

	tok, err := test.ReadToken("../internal/test/auth.tok", auth.TokenSource)
	if err != nil {
		t.Fatalf("error reading token: %s", err)
	}
	src := auth.RefreshTokenSource(tok)

	refresh, cancel := context.WithCancel(context.Background())
	defer cancel()
	prov := franchise.PlayFabXBLIdentityProvider{
		Environment: a,
		TokenSource: xal.RefreshTokenSourceContext(refresh, src, "http://playfab.xboxlive.com/"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	var d Dialer
	conn, err := d.DialContext(ctx, prov, s)
	if err != nil {
		t.Fatalf("error dialing: %s", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("error closing conn: %s", err)
		}
	})

	credentials, err := conn.Credentials()
	if err != nil {
		t.Fatalf("error obtaining credentials: %s", err)
	}
	if credentials == nil {
		t.Fatal("credentials is nil")
	}
	t.Logf("credentials obtained: %#v", credentials)
}
