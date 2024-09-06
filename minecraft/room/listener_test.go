package room

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/df-mc/go-playfab"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/auth/xal"
	"github.com/sandertv/gophertunnel/minecraft/franchise"
	"github.com/sandertv/gophertunnel/minecraft/franchise/signaling"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/oauth2"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestListen(t *testing.T) {
	discovery, err := franchise.Discover(protocol.CurrentVersion)
	if err != nil {
		t.Fatalf("error retrieving discovery: %s", err)
	}
	a := new(franchise.AuthorizationEnvironment)
	if err := discovery.Environment(a, franchise.EnvironmentTypeProduction); err != nil {
		t.Fatalf("error reading environment for authorization: %s", err)
	}
	s := new(signaling.Environment)
	if err := discovery.Environment(s, franchise.EnvironmentTypeProduction); err != nil {
		t.Fatalf("error reading environment for signaling: %s", err)
	}

	tok, err := readToken("../franchise/internal/test/auth.tok", auth.TokenSource)
	if err != nil {
		t.Fatalf("error reading token: %s", err)
	}
	src := auth.RefreshTokenSource(tok)

	i := franchise.PlayFabIdentityProvider{
		Environment: a,
		IdentityProvider: playfab.XBLIdentityProvider{
			TokenSource: xal.RefreshTokenSource(src, playfab.RelyingParty),
		},
	}

	d := signaling.Dialer{
		NetworkID: rand.Uint64(),
	}

	dial, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	signals, err := d.DialContext(dial, i, s)
	if err != nil {
		t.Fatalf("error dialing signaling: %s", err)
	}
	t.Cleanup(func() {
		if err := signals.Close(); err != nil {
			t.Fatalf("error closing signaling: %s", err)
		}
	})

	x := xal.RefreshTokenSource(src, "http://xboxlive.com")
	xt, err := x.Token()
	if err != nil {
		t.Fatalf("error requesting token: %s", err)
	}

	var p SessionPublishConfig
	announcer := p.New(x)

	status := DefaultStatus()
	status.OwnerID = xt.DisplayClaims().XUID

	minecraft.RegisterNetwork("room", Network{
		Network: minecraft.NetherNet{
			Signaling: signals,
		},
		ListenConfig: ListenConfig{
			StatusProvider: NewStatusProvider(status),
		},
		Announcer: announcer,
	})

	// The most of the code below has been copied from minecraft/example_listener_test.go.

	// Create a minecraft.Listener with a specific name to be displayed as MOTD in the server list.
	name := "MOTD of this server"
	cfg := minecraft.ListenConfig{
		StatusProvider: minecraft.NewStatusProvider(name, "Gophertunnel"),
	}

	listener, err := cfg.Listen("room", strconv.FormatUint(d.NetworkID, 10))
	if err != nil {
		t.Fatalf("error listening: %s", err)
	}
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Fatalf("error closing listener: %s", err)
		}
	})

	for {
		netConn, err := listener.Accept()
		if err != nil {
			return
		}
		c := netConn.(*minecraft.Conn)
		if err := c.StartGame(minecraft.GameData{
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
			// Allow inviting player into the world.
			GamePublishSetting: 3,
		}); err != nil {
			t.Fatalf("error starting game: %s", err)
		}

		go func() {
			defer func() {
				if err := c.Close(); err != nil {
					t.Errorf("error closing connection: %s", err)
				}
			}()
			for {
				pk, err := c.ReadPacket()
				if err != nil {
					// No output for errors which has occurred during decoding a packet,
					// since minecraft.Conn does not return net.ErrClosed.
					return
				}
				switch pk := pk.(type) {
				case *packet.Text:
					if pk.Message == "Close" {
						if err := listener.Disconnect(c, "Connection closed"); err != nil {
							t.Errorf("error closing connection: %s", err)
						}
						if err := listener.Close(); err != nil {
							t.Errorf("error closing listener: %s", err)
						}
					}
				}
			}
		}()
	}
}

func readToken(path string, src oauth2.TokenSource) (t *oauth2.Token, err error) {
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
