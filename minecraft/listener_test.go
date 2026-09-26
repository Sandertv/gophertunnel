package minecraft

import (
	"io"
	"log/slog"
	"net"
	"testing"
)

func TestCreateConnPreservesAcceptedProtocolsBackingArray(t *testing.T) {
	backing := []Protocol{proto{}, nil}
	group := &ListenerGroup{}
	group.playerCount.Store(1)
	listener := &Listener{
		cfg: ListenConfig{
			AcceptedProtocols: backing[:1],
			ErrorLog:          slog.New(slog.NewTextHandler(io.Discard, nil)),
			StatusProvider:    NewStatusProvider("test", ""),
			MaximumPlayers:    1,
		},
		group: group,
	}
	server, client := net.Pipe()
	drained := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, client)
		close(drained)
	}()
	t.Cleanup(func() {
		_ = server.Close()
		_ = client.Close()
		<-drained
	})

	// A full listener closes synchronously, so this test needs no login traffic.
	listener.createConn(server)
	if backing[1] != nil {
		t.Fatal("creating a connection changed the caller's backing array")
	}
}
