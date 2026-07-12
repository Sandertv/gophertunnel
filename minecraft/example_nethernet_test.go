package minecraft

import (
	"errors"
	"sync"

	"github.com/df-mc/go-nethernet/endpoint"
)

// ExampleNetherNet demonstrates a NetherNet<->NetherNet MITM proxy implementation.
func ExampleNetherNet() {
	// Start listening on HTTP/TLS for incoming connection requests.
	server, err := endpoint.ServeTLS(":19132", "/path/to/cert-file", "/path/to/key-file")
	if err != nil {
		panic(err)
	}
	defer server.Close()

	// Create a NetherNet network configuration using the HTTP server for
	// WebRTC Signaling (exchanging connection details with remote peers).
	n := NetherNet{Signaling: server}

	// Start listening on the Minecraft game protocol layer using the NetherNet network.
	// The address can be any string because the HTTP/TLS server is responsible for loading the
	// TLS certificate and key, and the current interface does not expose a way to configure them.
	//
	// More complex signaling implementations, such as the JSON-RPC/WebSocket protocol
	// used by peer-to-peer worlds, require authentication and connection cleanup. Those
	// concerns are therefore better handled outside of ListenNetwork.
	var cfg ListenConfig
	l, err := cfg.ListenNetwork(n, server.NetworkID())
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		go handleConn(conn.(*Conn), l)
	}
}

// handleConn handles an incoming connection by connecting to the upstream server over NetherNet.
// The upstream server can be hosted using a Bedrock Dedicated Server with 'transport=nethernet'
// set in server.properties.
func handleConn(conn *Conn, listener *Listener) {
	client := endpoint.NewClient()
	serverConn, err := Dialer{
		ClientData: conn.ClientData(),
	}.DialContextNetwork(conn.Context(), NetherNet{Signaling: client}, "http://localhost:19133")
	if err != nil {
		panic(err)
	}
	var g sync.WaitGroup
	g.Add(2)
	go func() {
		if err := conn.StartGame(serverConn.GameData()); err != nil {
			panic(err)
		}
		g.Done()
	}()
	go func() {
		if err := serverConn.DoSpawn(); err != nil {
			panic(err)
		}
		g.Done()
	}()
	g.Wait()

	go func() {
		defer listener.Disconnect(conn, "connection lost")
		defer serverConn.Close()
		for {
			pk, err := conn.ReadPacket()
			if err != nil {
				return
			}
			if err := serverConn.WritePacket(pk); err != nil {
				var disc DisconnectError
				if ok := errors.As(err, &disc); ok {
					_ = listener.Disconnect(conn, disc.Error())
				}
				return
			}
		}
	}()
	go func() {
		defer serverConn.Close()
		defer listener.Disconnect(conn, "connection lost")
		for {
			pk, err := serverConn.ReadPacket()
			if err != nil {
				var disc DisconnectError
				if ok := errors.As(err, &disc); ok {
					_ = listener.Disconnect(conn, disc.Error())
				}
				return
			}
			if err := conn.WritePacket(pk); err != nil {
				return
			}
		}
	}()
}
