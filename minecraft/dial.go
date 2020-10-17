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
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	rand2 "math/rand"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
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

	// TokenSource is the source for Microsoft Live Connect tokens. If set to a non-nil oauth2.TokenSource,
	// this field is used to obtain tokens which in turn are used to authenticate to XBOX Live.
	// The minecraft/auth package provides an oauth2.TokenSource implementation (auth.tokenSource) to use
	// device auth to login.
	// If TokenSource is nil, the connection will not use authentication.
	TokenSource oauth2.TokenSource

	// PacketFunc is called whenever a packet is read from or written to the connection returned when using
	// Dialer.Dial(). It includes packets that are otherwise covered in the connection sequence, such as the
	// Login packet. The function is called with the header of the packet and its raw payload, the address
	// from which the packet originated, and the destination address.
	PacketFunc func(header packet.Header, payload []byte, src, dst net.Addr)

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
	if dialer.TokenSource != nil {
		chainData, err = authChain(dialer.TokenSource, key)
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
		dialer := raknet.Dialer{ErrorLog: log.New(ioutil.Discard, "", 0)}
		var pong []byte
		pong, err = dialer.Ping(address)
		if err != nil {
			err = fmt.Errorf("raknet ping: %w", err)
			break
		}
		netConn, err = dialer.Dial(addressWithPongPort(pong, address))
		if err != nil {
			err = fmt.Errorf("raknet: %w", err)
		}
	default:
		// If not set to 'raknet', we fall back to the default net.Dial method to find a proper connection for
		// the network passed.
		netConn, err = net.Dial(network, address)
	}
	if err != nil {
		return nil, err
	}
	conn = newConn(netConn, key, dialer.ErrorLog)
	conn.identityData = dialer.IdentityData
	conn.clientData = dialer.ClientData
	conn.packetFunc = dialer.PacketFunc
	conn.cacheEnabled = dialer.EnableClientCache

	// Disable the batch packet limit so that the server can send packets as often as it wants to.
	conn.dec.DisableBatchPacketLimit()

	defaultClientData(address, conn.identityData.DisplayName, &conn.clientData)
	defaultIdentityData(&conn.identityData)

	var request []byte
	if dialer.TokenSource == nil {
		// We haven't logged into the user's XBL account. We create a login request with only one token
		// holding the identity data set in the Dialer.
		request = login.EncodeOffline(conn.identityData, conn.clientData, key)
	} else {
		// We login as an Android device and this will show up in the 'titleId' field in the JWT chain, which
		// we can't edit. We just enforce Android data for logging in.
		setAndroidData(&conn.clientData)

		request = login.Encode(chainData, conn.clientData, key)
		identityData, _, _ := login.Decode(request)
		// If we got the identity data from Minecraft auth, we need to make sure we set it in the Conn too, as
		// we are not aware of the identity data ourselves yet.
		conn.identityData = identityData
	}
	c := make(chan struct{})
	go listenConn(conn, dialer.ErrorLog, c)

	conn.expect(packet.IDServerToClientHandshake, packet.IDPlayStatus)
	if err := conn.WritePacket(&packet.Login{ConnectionRequest: request, ClientProtocol: protocol.CurrentProtocol}); err != nil {
		return nil, err
	}
	select {
	case <-conn.close:
		// The connection was closed before we even were fully 'connected', so we return an error.
		if conn.disconnectMessage.Load() != "" {
			return nil, fmt.Errorf("disconnected while connecting: %v", conn.disconnectMessage.Load())
		}
		return nil, fmt.Errorf("connection timeout")
	case <-c:
		// We've connected successfully. We return the connection and no error.
		return conn, nil
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
		packets, err := conn.dec.Decode()
		if err != nil {
			if !raknet.ErrConnectionClosed(err) {
				logger.Printf("error reading from dialer connection: %v\n", err)
			}
			return
		}
		for _, data := range packets {
			loggedInBefore := conn.loggedIn
			if err := conn.receive(data); err != nil {
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
func authChain(src oauth2.TokenSource, key *ecdsa.PrivateKey) (string, error) {
	// Obtain the Live token, and using that the XSTS token.
	liveToken, err := src.Token()
	if err != nil {
		return "", fmt.Errorf("error obtaining Live Connect token: %v", err)
	}
	xsts, err := auth.RequestXBLToken(liveToken, "https://multiplayer.minecraft.net/")
	if err != nil {
		return "", fmt.Errorf("error obtaining XBOX Live token: %v", err)
	}

	// Obtain the raw chain data using the
	chain, err := auth.RequestMinecraftChain(xsts, key)
	if err != nil {
		return "", fmt.Errorf("error obtaining Minecraft auth chain: %v", err)
	}
	return chain, nil
}

// defaultClientData edits the ClientData passed to have defaults set to all fields that were left unchanged.
func defaultClientData(address, username string, d *login.ClientData) {
	rand2.Seed(time.Now().Unix())

	d.ServerAddress = address
	if d.DeviceOS == 0 {
		d.DeviceOS = protocol.DeviceAndroid
	}
	if d.GameVersion == "" {
		d.GameVersion = protocol.CurrentVersion
	}
	if d.ClientRandomID == 0 {
		d.ClientRandomID = rand2.Int63()
	}
	if d.DeviceID == "" {
		d.DeviceID = uuid.New().String()
	}
	if d.LanguageCode == "" {
		d.LanguageCode = "en_GB"
	}
	if d.ThirdPartyName == "" {
		d.ThirdPartyName = username
	}
	if d.AnimatedImageData == nil {
		d.AnimatedImageData = make([]login.SkinAnimation, 0)
	}
	if d.PersonaPieces == nil {
		d.PersonaPieces = make([]login.PersonaPiece, 0)
	}
	if d.PieceTintColours == nil {
		d.PieceTintColours = make([]login.PersonaPieceTintColour, 0)
	}
	if d.SelfSignedID == "" {
		d.SelfSignedID = uuid.New().String()
	}
	if d.SkinID == "" {
		d.SkinID = uuid.New().String()
	}
	if d.SkinData == "" {
		d.SkinData = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0, 0, 0, 255}, 32*64))
		d.SkinImageHeight = 32
		d.SkinImageWidth = 64
	}
	if d.SkinResourcePatch == "" {
		p, _ := json.Marshal(map[string]interface{}{
			"geometry": map[string]interface{}{
				"default": "Standard_Custom",
			},
		})
		d.SkinResourcePatch = base64.StdEncoding.EncodeToString(p)
	}
}

// setAndroidData ensures the login.ClientData passed matches settings you would see on an Android device.
func setAndroidData(data *login.ClientData) {
	data.DeviceOS = protocol.DeviceAndroid
	data.GameVersion = protocol.CurrentVersion
}

// defaultIdentityData edits the IdentityData passed to have defaults set to all fields that were left
// unchanged.
func defaultIdentityData(data *login.IdentityData) {
	if data.Identity == "" {
		data.Identity = uuid.New().String()
	}
	if data.DisplayName == "" {
		data.DisplayName = "Steve"
	}
}

// regex is used to split strings by semicolons, except semicolons that are escaped. Note that this regex will
// not work properly with the case 'Test\\;`, where it would be expected that ';' is not escaped. Client-side,
// however, it is still escaped, so gophertunnel mimics this behaviour.
// (see https://github.com/Sandertv/gophertunnel/commit/d0c9c4c99cd02e441290efe4ee3568a39f7233f9#commitcomment-43046604)
var regex = regexp.MustCompile(`[^\\];`)

// addressWithPongPort parses the redirect IPv4 port from the pong and returns the address passed with the port
// found if present, or the original address if not.
func addressWithPongPort(pong []byte, address string) string {
	indices := regex.FindAllStringIndex(string(pong), -1)
	frag := make([]string, len(indices)+1)

	first := 0
	for i, index := range indices {
		frag[i] = string(pong[first : index[1]-1])
		first = index[1]
	}
	if len(frag) > 10 {
		portStr := frag[10]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return address
		}
		// Remove the port from the address.
		addressParts := strings.Split(address, ":")
		address = strings.Join(strings.Split(address, ":")[:len(addressParts)-1], ":")
		return address + ":" + strconv.Itoa(port)
	}
	return address
}
