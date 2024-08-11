package signaling

import (
	"context"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/franchise"
	"github.com/sandertv/gophertunnel/minecraft/franchise/internal/test"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/playfab"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"
	"math/rand"
	"strconv"
	"testing"
)

func TestDial(t *testing.T) {
	discovery, err := franchise.Discover(protocol.CurrentVersion)
	if err != nil {
		t.Fatalf("discover environments: %s", err)
	}
	a := new(franchise.AuthorizationEnvironment)
	if err := discovery.Environment(a, franchise.EnvironmentTypeProduction); err != nil {
		t.Fatalf("decode environment: %s", err)
	}

	src := test.TokenSource(t, "../internal/test/auth.tok", auth.TokenSource, func(old *oauth2.Token) (new *oauth2.Token, err error) {
		return auth.RefreshTokenSource(old).Token()
	})
	x, err := auth.RequestXBLToken(context.Background(), src, "http://playfab.xboxlive.com/")
	if err != nil {
		t.Fatalf("error requesting XBL token: %s", err)
	}

	identity, err := playfab.Login{
		Title:         "20CA2",
		CreateAccount: true,
	}.WithXBLToken(x).Login()
	if err != nil {
		t.Fatalf("error logging in to playfab: %s", err)
	}

	region, _ := language.English.Region()

	conf := &franchise.TokenConfig{
		Device: &franchise.DeviceConfig{
			ApplicationType: franchise.ApplicationTypeMinecraftPE,
			Capabilities:    []string{franchise.CapabilityRayTracing},
			GameVersion:     protocol.CurrentVersion,
			ID:              uuid.New(),
			Memory:          strconv.FormatUint(rand.Uint64(), 10),
			Platform:        franchise.PlatformWindows10,
			PlayFabTitleID:  a.PlayFabTitleID,
			StorePlatform:   franchise.StorePlatformUWPStore,
			Type:            franchise.DeviceTypeWindows10,
		},
		User: &franchise.UserConfig{
			Language:     language.English,
			LanguageCode: language.AmericanEnglish,
			RegionCode:   region.String(),
			Token:        identity.SessionTicket,
			TokenType:    franchise.TokenTypePlayFab,
		},
		Environment: a,
	}

	s := new(Environment)
	if err := discovery.Environment(s, franchise.EnvironmentTypeProduction); err != nil {
		t.Fatalf("decode environment: %s", err)
	}
	var d Dialer
	conn, err := d.DialContext(context.Background(), tokenConfigSource(func() (*franchise.TokenConfig, error) {
		return conf, nil
	}), s)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Errorf("clean up: error closing: %s", err)
		}
	})
}

type tokenConfigSource func() (*franchise.TokenConfig, error)

func (f tokenConfigSource) TokenConfig() (*franchise.TokenConfig, error) { return f() }
