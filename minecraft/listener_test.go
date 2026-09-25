package minecraft

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// A peer that connects but never logs in must be closed once LoginTimeout passes.
func TestListenerLoginTimeoutClosesSilentConnection(t *testing.T) {
	network := newPipeTestNetwork()
	listener, err := ListenConfig{AuthenticationDisabled: true, LoginTimeout: 50 * time.Millisecond}.ListenNetwork(network, "")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	client := network.connect()
	defer client.Close()
	if !waitClosed(client, time.Second) {
		t.Fatal("silent connection was not closed after the login timeout")
	}
	waitForZero(t, "player count", listener.PlayerCount)
	waitForZero(t, "pending logins", func() int { return int(listener.pendingLogins.Load()) })
}

// A client that logs in before LoginTimeout must stay connected after it passes.
func TestListenerLoginTimeoutKeepsLoggedInConnection(t *testing.T) {
	listener, err := ListenConfig{AuthenticationDisabled: true, LoginTimeout: 2 * time.Second}.Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	accepted := make(chan *Conn, 1)
	go func() {
		c, err := listener.Accept()
		if err != nil {
			return
		}
		conn := c.(*Conn)
		accepted <- conn
		_ = conn.StartGame(GameData{})
	}()
	client, err := Dialer{}.DialTimeout("raknet", listener.Addr().String(), 5*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()
	conn := <-accepted
	if n := listener.pendingLogins.Load(); n != 0 {
		t.Fatalf("pending logins = %d after login, want 0", n)
	}
	time.Sleep(3 * time.Second)
	select {
	case <-conn.Context().Done():
		t.Fatalf("logged-in connection was closed: %v", context.Cause(conn.Context()))
	default:
	}
}

// Connections beyond MaximumPendingLogins must be refused, and a connection that leaves frees its slot.
func TestListenerMaximumPendingLoginsRefusesExcessConnections(t *testing.T) {
	network := newPipeTestNetwork()
	listener, err := ListenConfig{AuthenticationDisabled: true, MaximumPendingLogins: 1}.ListenNetwork(network, "")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	first := network.connect()
	waitForCount(t, "pending logins", 1, func() int { return int(listener.pendingLogins.Load()) })
	refused := network.connect()
	defer refused.Close()
	if !waitClosed(refused, time.Second) {
		t.Fatal("connection beyond MaximumPendingLogins was not closed")
	}

	_ = first.Close()
	waitForZero(t, "pending logins", func() int { return int(listener.pendingLogins.Load()) })
	admitted := network.connect()
	defer admitted.Close()
	waitForCount(t, "pending logins", 1, func() int { return int(listener.pendingLogins.Load()) })
}

// waitClosed reports whether the peer of client closed within timeout.
func waitClosed(client net.Conn, timeout time.Duration) bool {
	_ = client.SetReadDeadline(time.Now().Add(timeout))
	_, err := io.Copy(io.Discard, client)
	return err == nil
}

func waitForZero(t *testing.T, what string, count func() int) {
	t.Helper()
	waitForCount(t, what, 0, count)
}

func waitForCount(t *testing.T, what string, want int, count func() int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for count() != want {
		if time.Now().After(deadline) {
			t.Fatalf("%s = %d, want %d", what, count(), want)
		}
		time.Sleep(time.Millisecond)
	}
}

// pipeTestNetwork connects clients to its own listener over in-memory pipes.
type pipeTestNetwork struct {
	conns  chan net.Conn
	closed chan struct{}
	once   sync.Once
}

func newPipeTestNetwork() *pipeTestNetwork {
	return &pipeTestNetwork{conns: make(chan net.Conn), closed: make(chan struct{})}
}

// connect hands a new connection to the listener and returns the client side.
func (n *pipeTestNetwork) connect() net.Conn {
	client, server := net.Pipe()
	n.conns <- server
	return client
}

func (n *pipeTestNetwork) DialContext(ctx context.Context, _ string) (net.Conn, error) {
	client, server := net.Pipe()
	select {
	case n.conns <- server:
		return client, nil
	case <-ctx.Done():
		_ = client.Close()
		_ = server.Close()
		return nil, ctx.Err()
	}
}

func (n *pipeTestNetwork) PingContext(context.Context, string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (n *pipeTestNetwork) Listen(string) (NetworkListener, error) { return n, nil }

func (n *pipeTestNetwork) Accept() (net.Conn, error) {
	select {
	case c := <-n.conns:
		return c, nil
	case <-n.closed:
		return nil, net.ErrClosed
	}
}

func (n *pipeTestNetwork) Close() error {
	n.once.Do(func() { close(n.closed) })
	return nil
}

func (n *pipeTestNetwork) Addr() net.Addr  { return &net.UDPAddr{IP: net.IPv4zero, Port: 19132} }
func (n *pipeTestNetwork) ID() int64       { return 1 }
func (n *pipeTestNetwork) PongData([]byte) {}
