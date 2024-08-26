package franchise

import (
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/franchise/internal/test"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/playfab"
	"github.com/sandertv/gophertunnel/xsapi/xal"
	"testing"
)

func TestToken(t *testing.T) {
	discovery, err := Discover(protocol.CurrentVersion)
	if err != nil {
		t.Fatalf("error retrieving discovery: %s", err)
	}
	a := new(AuthorizationEnvironment)
	if err := discovery.Environment(a, EnvironmentTypeProduction); err != nil {
		t.Fatalf("error reading environment for authorization: %s", err)
	}

	tok, err := test.ReadToken("internal/test/auth.tok", auth.TokenSource)
	if err != nil {
		t.Fatalf("error reading token: %s", err)
	}
	src := auth.RefreshTokenSource(tok)

	prov := PlayFabIdentityProvider{
		Environment: a,
		IdentityProvider: playfab.XBLIdentityProvider{
			TokenSource: xal.RefreshTokenSource(src, "http://playfab.xboxlive.com/"),
		},
	}

	conf, err := prov.TokenConfig()
	if err != nil {
		t.Fatalf("error requesting token config: %s", err)
	}

	token, err := conf.Token()
	if err != nil {
		t.Fatalf("error requesting token: %s", err)
	}

	t.Logf("%#v", token)
}
