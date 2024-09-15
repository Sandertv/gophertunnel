package room

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/auth/xal"
	"github.com/sandertv/gophertunnel/minecraft/franchise/signaling"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/oauth2"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestListen demonstrates a world displayed in the friend list.
func TestListen(t *testing.T) {
	tok, err := readToken("../franchise/internal/test/auth.tok", auth.TokenSource)
	if err != nil {
		t.Fatalf("error reading token: %s", err)
	}
	src := auth.RefreshTokenSource(tok)

	d := signaling.Dialer{
		NetworkID: rand.Uint64(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	signals, err := d.DialContext(ctx, src)
	if err != nil {
		t.Fatalf("error dialing signaling: %s", err)
	}
	t.Cleanup(func() {
		if err := signals.Close(); err != nil {
			t.Errorf("error closing signaling: %s", err)
		}
	})

	x, err := xal.RefreshTokenSource(src, "http://xboxlive.com").Token()
	if err != nil {
		t.Fatal(err)
	}

	status := DefaultStatus()
	status.OwnerID = x.DisplayClaims().XUID
	minecraft.RegisterNetwork("room", Network{
		Network: minecraft.NetherNet{
			Signaling: signals,
		},
		ListenConfig: ListenConfig{
			Announcer: &XBLAnnouncer{
				TokenSource: xal.RefreshTokenSource(src, "http://xboxlive.com"),
			},
			StatusProvider: NewStatusProvider(status),
		},
	})

	l, err := minecraft.Listen("room", strconv.FormatUint(d.NetworkID, 10))
	if err != nil {
		t.Fatalf("error listening: %s", err)
	}
	t.Cleanup(func() {
		if err := l.Close(); err != nil {
			t.Errorf("error closing listener: %s", err)
		}
	})

	for {
		n, err := l.Accept()
		if err != nil {
			return
		}

		conn := n.(*minecraft.Conn)
		if err := conn.StartGame(minecraft.GameData{
			WorldName:       "NetherNet - room.TestListen",
			WorldSeed:       rand.Int63(),
			EntityUniqueID:  rand.Int63(),
			EntityRuntimeID: rand.Uint64(),
			PlayerGameMode:  1,
			WorldGameMode:   1,
			// Allow inviting players to the world.
			GamePublishSetting: status.BroadcastSetting,
			Time:               rand.Int63(),
		}); err != nil {
			t.Errorf("error starting game: %s", err)
		}

		// Try reading and decoding deferred packets.
		go func() {
			for {
				pk, err := conn.ReadPacket()
				if err != nil {
					if !strings.Contains(err.Error(), net.ErrClosed.Error()) {
						t.Errorf("error decoding packet: %s", err)
					}
					if err := conn.Close(); err != nil {
						t.Errorf("error closing connection: %s", err)
					}
					return
				}

				switch pk := pk.(type) {
				case *packet.Text:
					if pk.TextType == packet.TextTypeChat && strings.EqualFold(pk.Message, "Close") {
						if err := conn.Close(); err != nil {
							t.Errorf("error closing connection: %s", err)
						}
						if err := l.Close(); err != nil {
							t.Errorf("error closing listener: %s", err)
						}
						return
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
