package hitasuradebug

import (
	"encoding/json"
	"log/slog"
	"maps"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/df-mc/go-xsapi/v2/mpsd"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/p2p"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/service/signaling/messaging"
	"github.com/sandertv/gophertunnel/minecraft/service/test"
)

func TestListen(t *testing.T) {
	user := test.PleaseRemoveThisBeforePRingSebWontAllowThis(t)

	networkID := rand.Uint64()
	m, err := messaging.Dialer{
		NetworkID: strconv.FormatUint(networkID, 10),
	}.DialContext(t.Context(), user.Minecraft())
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	world := p2p.World{
		Joinability:      p2p.JoinabilityFriends,
		HostName:         user.XSAPI().UserInfo().GamerTag,
		OwnerID:          user.XSAPI().UserInfo().XUID,
		Version:          protocol.CurrentVersion,
		LevelID:          "89r05rCXMlM=",
		WorldName:        "nonce test",
		WorldType:        "Creative",
		Protocol:         protocol.CurrentProtocol,
		MemberCount:      1,
		MaxMemberCount:   8,
		BroadcastSetting: p2p.BroadcastSettingFriendsOfFriends,
		TransportLayer:   p2p.TransportLayerNetherNet,
		Nonces:           make(map[string]string),
		SupportedConnections: []p2p.Connection{
			{
				Type:              p2p.ConnectionTypeSignalingOverJSONRPC,
				NetherNetID:       json.Number(strconv.FormatUint(networkID, 10)),
				PlayerMessagingID: m.PlayerMessagingID(),
			},
		},
	}

	minecraft.RegisterNetwork("nethernet", func(l *slog.Logger) minecraft.Network {
		return minecraft.NetherNet{Signaling: m, Log: l}
	})

	l, err := minecraft.ListenConfig{}.Listen("nethernet", "")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	session, err := user.XSAPI().MPSD().Publish(t.Context(), mpsd.SessionReference{
		TemplateName:    "MinecraftLobby",
		ServiceConfigID: auth.ServiceConfigID,
	}, mpsd.PublishConfig{
		CustomProperties: encodeJSON(t, world),
		JoinRestriction:  world.BroadcastSetting.JoinRestriction(),
		ReadRestriction:  world.BroadcastSetting.ReadRestriction(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	session.Handle(&sessionHandler{t: t, w: world})

	timer := time.AfterFunc(time.Minute, func() {
		if err := l.Close(); err != nil {
			t.Fatalf("error closing listener: %s", err)
		}
	})
	defer timer.Stop()

	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		t.Log(conn.(*minecraft.Conn).IdentityData())
	}
}

type sessionHandler struct {
	t testing.TB

	w  p2p.World
	mu sync.Mutex
}

func (s *sessionHandler) HandleSessionChange(session *mpsd.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.t.Log("handle session change called")

	var changed bool
	for _, member := range session.Members() {
		xuid := member.Constants.System.XUID
		if xuid == s.w.OwnerID {
			continue
		}
		if _, ok := s.w.Nonces[xuid]; !ok {
			s.w.Nonces[xuid] = "842321107f785505"
			s.t.Logf("set nonce for %s", xuid)
			changed = true
		}
	}
	maps.DeleteFunc(s.w.Nonces, func(xuid string, _ string) bool {
		_, ok := session.MemberByXUID(xuid)
		return !ok
	})
	if changed {
		if err := session.SetCustomProperties(s.t.Context(), encodeJSON(s.t, s.w)); err != nil {
			s.t.Errorf("error updating custom properties: %s", err)
		}
	}
}

func encodeJSON(t testing.TB, v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("error encoding JSON: %s", err)
	}
	return b
}
