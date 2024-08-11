package franchise

import (
	"context"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/franchise/internal/test"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/playfab"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"
	"math/rand"
	"strconv"
	"testing"
)

func TestToken(t *testing.T) {
	d, err := Discover(protocol.CurrentVersion)
	if err != nil {
		t.Fatalf("discover environments: %s", err)
	}
	a := new(AuthorizationEnvironment)
	if err := d.Environment(a, EnvironmentTypeProduction); err != nil {
		t.Fatalf("decode environment: %s", err)
	}

	src := test.TokenSource(t, "internal/test/auth.tok", auth.TokenSource, func(old *oauth2.Token) (new *oauth2.Token, err error) {
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

	conf := &TokenConfig{
		Device: &DeviceConfig{
			ApplicationType: ApplicationTypeMinecraftPE,
			Capabilities:    []string{CapabilityRayTracing},
			GameVersion:     protocol.CurrentVersion,
			ID:              uuid.New(),
			Memory:          strconv.FormatUint(rand.Uint64(), 10),
			Platform:        PlatformWindows10,
			PlayFabTitleID:  a.PlayFabTitleID,
			StorePlatform:   StorePlatformUWPStore,
			Type:            DeviceTypeWindows10,
		},
		User: &UserConfig{
			Language:     language.English,
			LanguageCode: language.AmericanEnglish,
			RegionCode:   region.String(),
			Token:        identity.SessionTicket,
			TokenType:    TokenTypePlayFab,
		},
		Environment: a,
	}

	tok, err := conf.Token()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", tok)
}
