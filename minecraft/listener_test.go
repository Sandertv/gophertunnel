package minecraft

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/internal"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestListenConfigListenNetworkUsesExplicitNetwork(t *testing.T) {
	t.Parallel()

	listener := fakeNetworkListener{addr: &net.UDPAddr{IP: net.IPv4zero, Port: 19132}}
	network := listenTestNetwork{
		listen: func(address string) (NetworkListener, error) {
			if address != "ignored-by-nethernet" {
				t.Fatalf("listen address = %q, want ignored-by-nethernet", address)
			}
			return listener, nil
		},
	}

	got, err := ListenConfig{AuthenticationDisabled: true}.ListenNetwork(network, "ignored-by-nethernet")
	if err != nil {
		t.Fatalf("ListenNetwork: %v", err)
	}
	defer got.Close()
	if got.listener != listener {
		t.Fatalf("underlying listener = %v, want explicit listener", got.listener)
	}
}

func TestListenerPublishesDisablePacketHandlingConnection(t *testing.T) {
	t.Parallel()

	client, server := net.Pipe()
	defer client.Close()

	log := slog.New(internal.DiscardHandler{})
	listener := &Listener{
		cfg: ListenConfig{
			ErrorLog:              log,
			StatusProvider:        NewStatusProvider("Minecraft Server", "Gophertunnel"),
			DisablePacketHandling: true,
		},
		listener: fakeNetworkListener{addr: &net.UDPAddr{IP: net.IPv4zero, Port: 19132}},
		incoming: make(chan *Conn, 1),
		close:    make(chan struct{}),
	}
	listener.playerCount.Store(1)

	conn := newConn(server, nil, log, proto{}, -1, true)
	conn.pool = conn.proto.Packets(true)
	conn.disablePacketHandling = true
	go listener.handleConn(conn)

	if err := writePacket(client, &packet.ResourcePacksInfo{}); err != nil {
		t.Fatalf("write packet: %v", err)
	}

	select {
	case accepted := <-listener.incoming:
		if accepted != conn {
			t.Fatalf("accepted connection = %p, want %p", accepted, conn)
		}
	case <-time.After(time.Second):
		t.Fatal("listener did not publish passthrough connection")
	}
}

func TestListenerConnHandlerReceivesDisablePacketHandlingConnection(t *testing.T) {
	t.Parallel()

	client, server := net.Pipe()
	defer client.Close()

	handled := make(chan *Conn, 1)
	log := slog.New(internal.DiscardHandler{})
	listener := &Listener{
		cfg: ListenConfig{
			ErrorLog:              log,
			StatusProvider:        NewStatusProvider("Minecraft Server", "Gophertunnel"),
			DisablePacketHandling: true,
			ConnHandler: func(conn *Conn) error {
				handled <- conn
				return nil
			},
		},
		listener: fakeNetworkListener{addr: &net.UDPAddr{IP: net.IPv4zero, Port: 19132}},
		incoming: make(chan *Conn, 1),
		close:    make(chan struct{}),
	}
	listener.playerCount.Store(1)

	conn := newConn(server, nil, log, proto{}, -1, true)
	conn.pool = conn.proto.Packets(true)
	conn.disablePacketHandling = true
	go listener.handleConn(conn)

	if err := writePacket(client, &packet.ResourcePacksInfo{}); err != nil {
		t.Fatalf("write packet: %v", err)
	}

	select {
	case accepted := <-handled:
		if accepted != conn {
			t.Fatalf("handled connection = %p, want %p", accepted, conn)
		}
	case <-time.After(time.Second):
		t.Fatal("listener did not deliver passthrough connection to ConnHandler")
	}

	select {
	case accepted := <-listener.incoming:
		t.Fatalf("listener published connection %p to Accept despite ConnHandler", accepted)
	default:
	}
}

func TestListenerDisablePacketHandlingConsumesClientHandshake(t *testing.T) {
	t.Parallel()

	client, server := net.Pipe()
	defer client.Close()

	log := slog.New(internal.DiscardHandler{})
	listener := &Listener{
		cfg: ListenConfig{
			ErrorLog:              log,
			StatusProvider:        NewStatusProvider("Minecraft Server", "Gophertunnel"),
			DisablePacketHandling: true,
		},
		listener: fakeNetworkListener{addr: &net.UDPAddr{IP: net.IPv4zero, Port: 19132}},
		incoming: make(chan *Conn, 1),
		close:    make(chan struct{}),
	}
	listener.playerCount.Store(1)

	conn := newConn(server, nil, log, proto{}, -1, true)
	conn.pool = conn.proto.Packets(true)
	conn.disablePacketHandling = true
	conn.expect(packet.IDClientToServerHandshake)
	go listener.handleConn(conn)

	if err := writePacket(client, &packet.ClientToServerHandshake{}); err != nil {
		t.Fatalf("write packet: %v", err)
	}

	select {
	case accepted := <-listener.incoming:
		if accepted != conn {
			t.Fatalf("accepted connection = %p, want %p", accepted, conn)
		}
	case <-time.After(time.Second):
		t.Fatal("listener did not publish passthrough connection")
	}

	if err := client.SetReadDeadline(time.Now().Add(50 * time.Millisecond)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	var b [1]byte
	n, err := client.Read(b[:])
	if err == nil || n != 0 {
		t.Fatalf("listener wrote %d byte(s) while consuming client handshake; expected no local response", n)
	}
	if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
		t.Fatalf("read error = %v, want timeout", err)
	}
}

func TestListenerPongDataUsesStatusProviderSubtitle(t *testing.T) {
	t.Parallel()

	var pongData []byte
	listener := &Listener{
		cfg: ListenConfig{
			StatusProvider: NewStatusProvider("Minecraft Server", "Provider Subtitle"),
		},
		listener: fakeNetworkListener{
			addr:     &net.UDPAddr{IP: net.IPv4zero, Port: 19132},
			pongData: &pongData,
		},
	}
	listener.updatePongData()

	status := ParsePongData(pongData)
	if status.ServerName != "Minecraft Server" {
		t.Fatalf("server name = %q, want Minecraft Server", status.ServerName)
	}
	if status.ServerSubName != "Provider Subtitle" {
		t.Fatalf("server subtitle = %q, want Provider Subtitle", status.ServerSubName)
	}
}

func writePacket(w io.Writer, pk packet.Packet) error {
	buf := new(bytes.Buffer)
	header := &packet.Header{PacketID: pk.ID()}
	if err := header.Write(buf); err != nil {
		return err
	}
	pk.Marshal(proto{}.NewWriter(buf, 0))
	return packet.NewEncoder(w).Encode([][]byte{buf.Bytes()})
}

type fakeNetworkListener struct {
	addr     net.Addr
	pongData *[]byte
}

func (f fakeNetworkListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (f fakeNetworkListener) Close() error              { return nil }
func (f fakeNetworkListener) Addr() net.Addr            { return f.addr }
func (fakeNetworkListener) ID() int64                   { return 1 }
func (f fakeNetworkListener) PongData(data []byte) {
	if f.pongData != nil {
		*f.pongData = append((*f.pongData)[:0], data...)
	}
}

type listenTestNetwork struct {
	listen func(string) (NetworkListener, error)
}

func (listenTestNetwork) DialContext(context.Context, string) (net.Conn, error) {
	return nil, errors.New("not implemented")
}

func (listenTestNetwork) PingContext(context.Context, string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (n listenTestNetwork) Listen(address string) (NetworkListener, error) {
	return n.listen(address)
}
