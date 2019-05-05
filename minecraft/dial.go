package minecraft

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/opt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/device"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"log"
	rand2 "math/rand"
	"net"
	"time"
)

// Dial dials a Minecraft connection to the address passed over the network passed. The network must be "tcp",
// "tcp4", "tcp6", "unix", "unixpacket" or "raknet". A Conn is returned which may be used to receive packets
// from and send packets to.
// A list of optional options that may be passed may be found in the minecraft/opt package. It includes
// options such as the credentials to login to XBOX Live.
func Dial(network string, address string, opts ...opt.Opt) (conn *Conn, err error) {
	var netConn net.Conn
	switch network {
	case "raknet":
		// If the network is specifically 'raknet', we use the raknet library to dial a RakNet connection.
		netConn, err = raknet.Dial(address)
	default:
		// If not set to 'raknet', we fall back to the default net.Dial method to find a proper connection for
		// the network passed.
		netConn, err = net.Dial(network, address)
	}
	if err != nil {
		return nil, err
	}
	key, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	options := opt.Map(opts)
	var chainData string
	if v, ok := options["credentials"]; ok {
		chainData, err = authChain(v.(opt.Creds), key)
		if err != nil {
			return nil, err
		}
	}
	conn = newConn(netConn, key)
	conn.clientData = defaultClientData(address)
	if v, ok := options["client_data"]; ok {
		conn.clientData = v.(login.ClientData)
	}
	conn.expect(packet.IDServerToClientHandshake, packet.IDPlayStatus)

	go listenConn(conn)

	request := login.Encode(chainData, conn.clientData, key)
	if err := conn.WritePacket(&packet.Login{ConnectionRequest: request, ClientProtocol: protocol.CurrentProtocol}); err != nil {
		return nil, err
	}
	select {
	case <-conn.connected:
		// We've connected successfully. We return the connection and no error.
		return conn, nil
	case <-conn.close:
		// The connection was closed before we even were fully 'connected', so we return an error.
		conn.close <- true
		return nil, fmt.Errorf("connection timeout")
	}
}

// listenConn listens on the connection until it is closed on another goroutine.
func listenConn(conn *Conn) {
	defer func() {
		_ = conn.Close()
	}()
	for {
		// We finally arrived at the packet decoding loop. We constantly decode packets that arrive
		// and push them to the Conn so that they may be processed.
		packets, err := conn.decoder.Decode()
		if err != nil {
			if !raknet.ErrConnectionClosed(err) {
				log.Printf("error reading from client connection: %v\n", err)
			}
			return
		}
		for _, data := range packets {
			if err := conn.handleIncoming(data); err != nil {
				log.Printf("%v", err)
				return
			}
		}
	}
}

// authChain requests the Minecraft auth JWT chain using the credentials passed. If successful, an encoded
// chain ready to be put in a login request is returned.
func authChain(credentials opt.Creds, key *ecdsa.PrivateKey) (string, error) {
	// Obtain the Live token, and using that the XSTS token.
	liveToken, err := auth.RequestLiveToken(credentials.Login, credentials.Password)
	if err != nil {
		return "", fmt.Errorf("error obtaining Live token: %v", err)
	}
	xsts, err := auth.RequestXSTSToken(liveToken)
	if err != nil {
		return "", fmt.Errorf("error obtaining XSTS token: %v", err)
	}

	// Obtain the raw chain data using the
	chain, err := auth.RequestMinecraftChain(xsts, key)
	if err != nil {
		return "", fmt.Errorf("error obtaining Minecraft auth chain: %v", err)
	}
	return chain, nil
}

// defaultClientData returns a valid, mostly filled out ClientData struct using the connection address
// passed, which is sent by default, if no other client data is set.
func defaultClientData(address string) login.ClientData {
	rand2.Seed(time.Now().Unix())
	return login.ClientData{
		ClientRandomID:   rand2.Int63(),
		DeviceOS:         device.Win10,
		GameVersion:      protocol.CurrentVersion,
		DeviceID:         uuid.Must(uuid.NewRandom()).String(),
		LanguageCode:     "en_UK",
		ThirdPartyName:   "Steve",
		SelfSignedID:     uuid.Must(uuid.NewRandom()).String(),
		SkinGeometryName: "geometry.humanoid",
		ServerAddress:    address,
		SkinID:           uuid.Must(uuid.NewRandom()).String(),
		SkinData:         base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0}, 32*64*4)),
	}
}
