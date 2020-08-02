package minecraft

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"io/ioutil"
	"log"
	rand2 "math/rand"
	"net"
	"os"
	"time"
)

// Dialer allows specifying specific settings for connection to a Minecraft server.
// The zero value of Dialer is used for the package level Dial function.
type Dialer struct {
	// ErrorLog is a log.Logger that errors that occur during packet handling of servers are written to. By
	// default, ErrorLog is set to one equal to the global logger.
	ErrorLog *log.Logger

	// ClientData is the client data used to login to the server with. It includes fields such as the skin,
	// locale and UUIDs unique to the client. If empty, a default is sent produced using defaultClientData().
	ClientData login.ClientData
	// IdentityData is the identity data used to login to the server with. It includes the username, UUID and
	// XUID of the player.
	// The IdentityData object is obtained using Minecraft auth if Email and Password are set. If not, the
	// object provided here is used, or a default one if left empty.
	IdentityData login.IdentityData

	// Email is the email used to login to the XBOX Live account. If empty, no attempt will be made to login,
	// and an unauthenticated login request will be sent.
	Email string
	// Password is the password used to login to the XBOX Live account. If Email is non-empty, a login attempt
	// will be made using this password.
	Password string

	// PacketFunc is called whenever a packet is read from or written to the connection returned when using
	// Dialer.Dial(). It includes packets that are otherwise covered in the connection sequence, such as the
	// Login packet. The function is called with the header of the packet and its raw payload, the address
	// from which the packet originated, and the destination address.
	PacketFunc func(header packet.Header, payload []byte, src, dst net.Addr)

	// SendPacketViolations makes the Dialer send PacketViolationWarnings to servers it connects to when it
	// receives packets it cannot decode properly. Additionally, it will log PacketViolationWarnings coming
	// from the server.
	SendPacketViolations bool

	// EnableClientCache, if set to true, enables the client blob cache for the client. This means that the
	// server will send chunks as blobs, which may be saved by the client so that chunks don't have to be
	// transmitted every time, resulting in less network transmission.
	EnableClientCache bool
}

// Dial dials a Minecraft connection to the address passed over the network passed. The network is typically
// "raknet". A Conn is returned which may be used to receive packets from and send packets to.
//
// A zero value of a Dialer struct is used to initiate the connection. A custom Dialer may be used to specify
// additional behaviour.
func Dial(network string, address string) (conn *Conn, err error) {
	return Dialer{}.Dial(network, address)
}

// Dial dials a Minecraft connection to the address passed over the network passed. The network is typically
// "raknet". A Conn is returned which may be used to receive packets from and send packets to.
// Specific fields in the Dialer specify additional behaviour during the connection, such as authenticating
// to XBOX Live and custom client data.
func (dialer Dialer) Dial(network string, address string) (conn *Conn, err error) {
	key, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	var chainData string
	if dialer.Email != "" {
		chainData, err = authChain(dialer.Email, dialer.Password, key)
		if err != nil {
			return nil, err
		}
	}
	if dialer.ErrorLog == nil {
		dialer.ErrorLog = log.New(os.Stderr, "", log.LstdFlags)
	}
	var netConn net.Conn

	switch network {
	case "raknet":
		// If the network is specifically 'raknet', we use the raknet library to dial a RakNet connection.
		netConn, err = raknet.Dialer{ErrorLog: log.New(ioutil.Discard, "", 0)}.Dial(address)
	default:
		// If not set to 'raknet', we fall back to the default net.Dial method to find a proper connection for
		// the network passed.
		netConn, err = net.Dial(network, address)
	}
	if err != nil {
		return nil, err
	}
	conn = newConn(netConn, key, dialer.ErrorLog)
	conn.clientData = defaultClientData(address)
	conn.identityData = defaultIdentityData()
	conn.packetFunc = dialer.PacketFunc
	conn.cacheEnabled = dialer.EnableClientCache
	conn.sendPacketViolations = dialer.SendPacketViolations
	// Disable the batch packet limit so that the server can send packets as often as it wants to.
	conn.decoder.DisableBatchPacketLimit()

	if dialer.ClientData.SkinID != "" {
		// If a custom client data struct was set, we change the default.
		conn.clientData = dialer.ClientData
	}
	var emptyIdentityData login.IdentityData
	if dialer.IdentityData != emptyIdentityData {
		// If a custom identity data object was set, we change the default.
		conn.identityData = dialer.IdentityData
	}
	conn.expect(packet.IDServerToClientHandshake, packet.IDPlayStatus)

	c := make(chan struct{})
	go listenConn(conn, dialer.ErrorLog, c)

	if conn.clientData.AnimatedImageData == nil {
		conn.clientData.AnimatedImageData = make([]login.SkinAnimation, 0)
	}
	if conn.clientData.PersonaPieces == nil {
		conn.clientData.PersonaPieces = make([]login.PersonaPiece, 0)
	}
	if conn.clientData.PieceTintColours == nil {
		conn.clientData.PieceTintColours = make([]login.PersonaPieceTintColour, 0)
	}

	var request []byte
	if dialer.Email == "" {
		// We haven't logged into the user's XBL account. We create a login request with only one token
		// holding the identity data set in the Dialer.
		request = login.EncodeOffline(conn.identityData, conn.clientData, key)

	} else {
		request = login.Encode(chainData, conn.clientData, key)
		identityData, _, _ := login.Decode(request)
		// If we got the identity data from Minecraft auth, we need to make sure we set it in the Conn too, as
		// we are not aware of the identity data ourselves yet.
		conn.identityData = identityData
	}
	if err := conn.WritePacket(&packet.Login{ConnectionRequest: request, ClientProtocol: protocol.CurrentProtocol}); err != nil {
		return nil, err
	}
	select {
	case <-c:
		// We've connected successfully. We return the connection and no error.
		return conn, nil
	case <-conn.closeCtx.Done():
		// The connection was closed before we even were fully 'connected', so we return an error.
		if conn.disconnectMessage.Load() != "" {
			return nil, fmt.Errorf("disconnected while connecting: %v", conn.disconnectMessage.Load())
		}
		return nil, fmt.Errorf("connection timeout")
	}
}

// listenConn listens on the connection until it is closed on another goroutine. The channel passed will
// receive a value once the connection is logged in.
func listenConn(conn *Conn, logger *log.Logger, c chan struct{}) {
	defer func() {
		_ = conn.Close()
	}()
	for {
		// We finally arrived at the packet decoding loop. We constantly decode packets that arrive
		// and push them to the Conn so that they may be processed.
		packets, err := conn.decoder.Decode()
		if err != nil {
			if !raknet.ErrConnectionClosed(err) {
				logger.Printf("error reading from client connection: %v\n", err)
			}
			return
		}
		for _, data := range packets {
			loggedInBefore := conn.loggedIn
			if err := conn.handleIncoming(data); err != nil {
				logger.Printf("error: %v", err)
				return
			}
			if !loggedInBefore && conn.loggedIn {
				// This is the signal that the connection was considered logged in, so we put a value in the
				// channel so that it may be detected.
				c <- struct{}{}
			}
		}
	}
}

// authChain requests the Minecraft auth JWT chain using the credentials passed. If successful, an encoded
// chain ready to be put in a login request is returned.
func authChain(email, password string, key *ecdsa.PrivateKey) (string, error) {
	// Obtain the Live token, and using that the XSTS token.
	liveToken, err := auth.RequestLiveToken(email, password)
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
	p, _ := json.Marshal(map[string]interface{}{
		"geometry": map[string]interface{}{
			"default": "Standard_Custom",
		},
	})
	return login.ClientData{
		ClientRandomID:    rand2.Int63(),
		DeviceOS:          protocol.DeviceWin10,
		GameVersion:       protocol.CurrentVersion,
		DeviceID:          uuid.Must(uuid.NewRandom()).String(),
		LanguageCode:      "en_GB",
		ThirdPartyName:    "Steve",
		SelfSignedID:      uuid.Must(uuid.NewRandom()).String(),
		ServerAddress:     address,
		SkinID:            uuid.Must(uuid.NewRandom()).String(),
		SkinData:          base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0, 0, 0, 255}, 32*64)),
		SkinResourcePatch: base64.StdEncoding.EncodeToString(p),
		SkinImageWidth:    64,
		SkinImageHeight:   32,
	}
}

// defaultIdentityData returns a valid default identity data object which may be used to fill out if the
// client is not authenticated and if no identity data was provided.
func defaultIdentityData() login.IdentityData {
	return login.IdentityData{
		Identity:    uuid.New().String(),
		DisplayName: "Steve",
	}
}
