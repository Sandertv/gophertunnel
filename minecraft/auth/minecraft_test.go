package auth

import (
	"context"
	"net/url"
	"testing"

	"github.com/df-mc/go-xsapi/v2/xal/xsts"
)

func TestMinecraftTokenSignerUsesMinecraftRelyingParty(t *testing.T) {
	t.Parallel()

	wantToken := &xsts.Token{}
	src := &fakeXSTSSource{token: wantToken}
	token, _, err := (MinecraftTokenSigner{Source: src}).TokenAndSignature(context.Background(), &url.URL{
		Scheme: "https",
		Host:   "multiplayer.minecraft.net",
		Path:   "/authentication",
	})
	if err != nil {
		t.Fatalf("TokenAndSignature: %v", err)
	}
	if token != wantToken {
		t.Fatalf("token = %p, want %p", token, wantToken)
	}
	if src.relyingParty != MinecraftRelyingParty {
		t.Fatalf("relying party = %q, want %q", src.relyingParty, MinecraftRelyingParty)
	}
}

type fakeXSTSSource struct {
	relyingParty string
	token        *xsts.Token
}

func (f *fakeXSTSSource) XSTSToken(_ context.Context, relyingParty string) (*xsts.Token, error) {
	f.relyingParty = relyingParty
	return f.token, nil
}
