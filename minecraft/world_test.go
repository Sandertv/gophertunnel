package minecraft

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/kr/pretty"
	"github.com/pion/sdp/v3"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/franchise"
	"github.com/sandertv/gophertunnel/minecraft/franchise/signaling"
	"github.com/sandertv/gophertunnel/minecraft/nethernet"
	"net"
	"os"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/playfab"
	"github.com/sandertv/gophertunnel/xsapi"
	"github.com/sandertv/gophertunnel/xsapi/mpsd"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

// TestListen demonstrates a world displayed in the friend list.
func TestWorld(t *testing.T) {
	discovery, err := franchise.Discover(protocol.CurrentVersion)
	if err != nil {
		t.Fatalf("discover: %s", err)
	}
	a := new(franchise.AuthorizationEnvironment)
	if err := discovery.Environment(a, franchise.EnvironmentTypeProduction); err != nil {
		t.Fatalf("decode environment: %s", err)
	}

	src := TokenSource(t, "franchise/internal/test/auth.tok", auth.TokenSource, func(old *oauth2.Token) (new *oauth2.Token, err error) {
		return auth.RefreshTokenSource(old).Token()
	})
	x, err := auth.RequestXBLToken(context.Background(), src, "http://xboxlive.com")
	if err != nil {
		t.Fatalf("error requesting XBL token: %s", err)
	}
	playfabXBL, err := auth.RequestXBLToken(context.Background(), src, "http://playfab.xboxlive.com/")
	if err != nil {
		t.Fatalf("error requesting XBL token: %s", err)
	}

	identity, err := playfab.Login{
		Title:         "20CA2",
		CreateAccount: true,
	}.WithXBLToken(playfabXBL).Login()
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

	s := new(signaling.Environment)
	if err := discovery.Environment(s, franchise.EnvironmentTypeProduction); err != nil {
		t.Fatalf("decode environment: %s", err)
	}
	sd := signaling.Dialer{
		NetworkID: rand.Uint64(),
	}
	signalingConn, err := sd.DialContext(context.Background(), tokenConfigSource(func() (*franchise.TokenConfig, error) {
		return conf, nil
	}), s)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := signalingConn.Close(); err != nil {
			t.Errorf("clean up: error closing: %s", err)
		}
	})

	var (
		displayClaims = x.AuthorizationToken.DisplayClaims.UserInfo[0]
		name          = strings.ToUpper(uuid.NewString()) // The name of the session.
	)
	custom, err := json.Marshal(map[string]any{
		"Joinability":             "joinable_by_friends",
		"hostName":                displayClaims.GamerTag,
		"ownerId":                 displayClaims.XUID,
		"rakNetGUID":              "",
		"version":                 "1.21.2",
		"levelId":                 "lhhPZjgNAQA=",
		"worldName":               name,
		"worldType":               "Creative",
		"protocol":                686,
		"MemberCount":             1,
		"MaxMemberCount":          8,
		"BroadcastSetting":        3,
		"LanGame":                 true,
		"isEditorWorld":           false,
		"TransportLayer":          2, // Zero means RakNet, and two means NetherNet.
		"WebRTCNetworkId":         sd.NetworkID,
		"OnlineCrossPlatformGame": true,
		"CrossPlayDisabled":       false,
		"TitleId":                 0,
		"SupportedConnections": []map[string]any{
			{
				"ConnectionType":  3,
				"HostIpAddress":   "",
				"HostPort":        0,
				"NetherNetId":     sd.NetworkID,
				"WebRTCNetworkId": sd.NetworkID,
				"RakNetGUID":      "UNASSIGNED_RAKNET_GUID",
			},
		},
	})
	if err != nil {
		t.Fatalf("error encoding custom properties: %s", err)
	}
	pub := mpsd.PublishConfig{
		Description: &mpsd.SessionDescription{
			Properties: &mpsd.SessionProperties{
				System: &mpsd.SessionPropertiesSystem{
					JoinRestriction: mpsd.SessionRestrictionFollowed,
					ReadRestriction: mpsd.SessionRestrictionFollowed,
				},
				Custom: custom,
			},
		},
	}
	session, err := pub.PublishContext(context.Background(), &tokenSource{
		x: x,
	}, mpsd.SessionReference{
		ServiceConfigID: uuid.MustParse("4fc10100-5f7a-4470-899b-280835760c07"),
		TemplateName:    "MinecraftLobby",
		Name:            name,
	})
	if err != nil {
		t.Fatalf("error publishing session: %s", err)
	}
	t.Cleanup(func() {
		if err := session.Close(); err != nil {
			t.Errorf("error closing session: %s", err)
		}
	})

	t.Logf("Network ID: %d", sd.NetworkID)
	t.Logf("Session Name: %q", name)

	RegisterNetwork("nethernet", &network{
		networkID: sd.NetworkID,
		signaling: signalingConn,
	})

	l, err := Listen("nethernet", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := l.Close(); err != nil {
			t.Fatal(err)
		}
	})

	for {
		conn, err := l.Accept()
		if err != nil {
			t.Fatal(err)
		}
		c := conn.(*Conn)
		_ = c.StartGame(GameData{
			WorldName:         "NetherNet",
			WorldSeed:         0,
			Difficulty:        0,
			EntityUniqueID:    rand.Int63(),
			EntityRuntimeID:   rand.Uint64(),
			PlayerGameMode:    1,
			PlayerPosition:    mgl32.Vec3{},
			WorldSpawn:        protocol.BlockPos{},
			WorldGameMode:     1,
			Time:              rand.Int63(),
			PlayerPermissions: 2,
		})
	}
}

func TestDecodeOffer(t *testing.T) {
	d := &sdp.SessionDescription{}
	if err := d.UnmarshalString("v=0\r\no=- 8735254407289596231 2 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0\r\na=extmap-allow-mixed\r\na=msid-semantic: WMS\r\nm=application 9 UDP/DTLS/SCTP webrtc-datachannel\r\nc=IN IP4 0.0.0.0\r\na=ice-ufrag:gMX+\r\na=ice-pwd:4SN4mwDq5k9Q2LwCiMqxacaM\r\na=ice-options:trickle\r\na=fingerprint:sha-256 B2:35:F2:64:66:B3:73:B3:BB:8D:EE:AF:D8:96:6C:29:9C:A9:E8:94:B3:67:E1:B9:77:8C:18:19:EA:29:7D:12\r\na=setup:actpass\r\na=mid:0\r\na=sctp-port:5000\r\na=max-message-size:262144\r\n"); err != nil {
		t.Fatal(err)
	}
	pretty.Println(d)
}

type network struct {
	networkID uint64
	signaling nethernet.Signaling
}

func (network) DialContext(context.Context, string) (net.Conn, error) {
	panic("not implemented (yet)")
}

func (network) PingContext(context.Context, string) ([]byte, error) {
	panic("not implemented (yet)")
}

func (n network) Listen(string) (NetworkListener, error) {
	var c nethernet.ListenConfig
	return c.Listen(n.networkID, n.signaling)
}

func (network) Encrypted() bool { return true }

// tokenSource is an implementation of xsapi.TokenSource that simply returns a *auth.XBLToken.
type tokenSource struct{ x *auth.XBLToken }

func (t *tokenSource) Token() (xsapi.Token, error) {
	return &token{t.x}, nil
}

type token struct {
	*auth.XBLToken
}

func (t *token) DisplayClaims() xsapi.DisplayClaims {
	return t.AuthorizationToken.DisplayClaims.UserInfo[0]
}

type tokenConfigSource func() (*franchise.TokenConfig, error)

func (f tokenConfigSource) TokenConfig() (*franchise.TokenConfig, error) { return f() }

func TokenSource(t *testing.T, path string, src oauth2.TokenSource, hooks ...RefreshTokenFunc) *oauth2.Token {
	tok, err := readTokenSource(path, src)
	if err != nil {
		t.Fatalf("error reading token: %s", err)
	}
	for _, h := range hooks {
		tok, err = h(tok)
		if err != nil {
			t.Fatalf("error refreshing token: %s", err)
		}
	}
	return tok
}

type RefreshTokenFunc func(old *oauth2.Token) (new *oauth2.Token, err error)

func readTokenSource(path string, src oauth2.TokenSource) (t *oauth2.Token, err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t, err = src.Token()
		if err != nil {
			return nil, fmt.Errorf("obtain token: %w", err)
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if err := json.NewEncoder(f).Encode(t); err != nil {
			return nil, fmt.Errorf("encode: %w", err)
		}
		return t, nil
	} else if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&t); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return t, nil
}
