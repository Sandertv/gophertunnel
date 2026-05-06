package hitasuradebug

import (
	"log/slog"
	"math/rand"
	"testing"

	"github.com/df-mc/go-nethernet"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/p2p"
	"github.com/sandertv/gophertunnel/minecraft/service/signaling"
	"github.com/sandertv/gophertunnel/minecraft/service/signaling/messaging"
	"github.com/sandertv/gophertunnel/minecraft/service/test"
)

func TestDialP2P(t *testing.T) {
	user := test.PleaseRemoveThisBeforePRingSebWontAllowThis(t)

	client := p2p.NewClient(user.XSAPI())
	worlds, err := client.Worlds(t.Context())
	if err != nil {
		t.Fatalf("error searching for open worlds: %s", err)
	}
	t.Log(worlds)
	if len(worlds) == 0 {
		t.Fatalf("no open worlds")
	}

	world := worlds[0]
	connection := world.SupportedConnections[0]

	var s nethernet.Signaling
	switch connection.Type {
	case p2p.ConnectionTypeSignalingOverJSONRPC:
		var d messaging.Dialer
		conn, err := d.DialContext(t.Context(), user.Minecraft())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		s = conn
	case p2p.ConnectionTypeSignalingOverWebSocket:
		var d signaling.Dialer
		conn, err := d.DialContext(t.Context(), user.Minecraft())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		s = conn
	default:
		t.Fatalf("invalid connection type: %d", connection.Type)
	}

	minecraft.RegisterNetwork("nethernet", func(l *slog.Logger) minecraft.Network {
		return minecraft.NetherNet{Signaling: s, Log: l}
	})

	conn, err := minecraft.Dialer{
		XBLClient: user.XSAPI(),
	}.DialContext(t.Context(), "nethernet", connection.Address())
	if err != nil {
		t.Fatalf("error dialing: %s", err)
	}
	defer conn.Close()

	if err := conn.DoSpawn(); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidNetworkID(t *testing.T) {
	user := test.PleaseRemoveThisBeforePRingSebWontAllowThis(t)

	var d messaging.Dialer
	conn, err := d.DialContext(t.Context(), user.Minecraft())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("error closing messaging conn: %s", err)
		}
	}()

	peer, _ := webrtc.NewPeerConnection(webrtc.Configuration{})
	sdp, _ := peer.CreateOffer(nil)

	t.Log(conn.Signal(t.Context(), &nethernet.Signal{
		Type:         nethernet.SignalTypeOffer,
		ConnectionID: rand.Uint64(),
		Data:         sdp.SDP,
		NetworkID:    uuid.NewString(),
	}))
}
