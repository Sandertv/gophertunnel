package minecraft

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/internal/resource"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/oauth2"
	"log"
	rand2 "math/rand"
	"net"
	"os"
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

	// Protocol is the Protocol version used to communicate with the target server. By default, this field is
	// set to the current protocol as implemented in the minecraft/protocol package. Note that packets written
	// to and read from the Conn are always any of those found in the protocol/packet package, as packets
	// are converted from and to this Protocol.
	Protocol Protocol

	// EnableClientCache, if set to true, enables the client blob cache for the client. This means that the
	// server will send chunks as blobs, which may be saved by the client so that chunks don't have to be
	// transmitted every time, resulting in less network transmission.
	EnableClientCache bool

	// KeepXBLIdentityData, if set to true, enables passing XUID and title ID to the target server
	// if the authentication token is not set. This is technically not valid and some servers might kick
	// the client when an XUID is present without logging in.
	// For getting this to work with BDS, authentication should be disabled.
	KeepXBLIdentityData bool
}

// Dial dials a Minecraft connection to the address passed over the network passed. The network is typically
// "raknet". A Conn is returned which may be used to receive packets from and send packets to.
//
// A zero value of a Dialer struct is used to initiate the connection. A custom Dialer may be used to specify
// additional behaviour.
func Dial(network, address string) (*Conn, error) {
	var d Dialer
	return d.Dial(network, address)
}

// DialTimeout dials a Minecraft connection to the address passed over the network passed. The network is
// typically "raknet". A Conn is returned which may be used to receive packets from and send packets to.
// If a connection is not established before the timeout ends, DialTimeout returns an error.
// DialTimeout uses a zero value of Dialer to initiate the connection.
func DialTimeout(network, address string, timeout time.Duration) (*Conn, error) {
	var d Dialer
	return d.DialTimeout(network, address, timeout)
}

// DialContext dials a Minecraft connection to the address passed over the network passed. The network is
// typically "raknet". A Conn is returned which may be used to receive packets from and send packets to.
// If a connection is not established before the context passed is cancelled, DialContext returns an error.
// DialContext uses a zero value of Dialer to initiate the connection.
func DialContext(ctx context.Context, network, address string) (*Conn, error) {
	var d Dialer
	return d.DialContext(ctx, network, address)
}

// Dial dials a Minecraft connection to the address passed over the network passed. The network is typically
// "raknet". A Conn is returned which may be used to receive packets from and send packets to.
func (d Dialer) Dial(network, address string) (*Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	return d.DialContext(ctx, network, address)
}

// DialTimeout dials a Minecraft connection to the address passed over the network passed. The network is
// typically "raknet". A Conn is returned which may be used to receive packets from and send packets to.
// If a connection is not established before the timeout ends, DialTimeout returns an error.
func (d Dialer) DialTimeout(network, address string, timeout time.Duration) (*Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return d.DialContext(ctx, network, address)
}

// DialContext dials a Minecraft connection to the address passed over the network passed. The network is
// typically "raknet". A Conn is returned which may be used to receive packets from and send packets to.
// If a connection is not established before the context passed is cancelled, DialContext returns an error.
func (d Dialer) DialContext(ctx context.Context, network, address string) (conn *Conn, err error) {
	key, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	var chainData string
	if d.TokenSource != nil {
		chainData, err = authChain(ctx, d.TokenSource, key)
		if err != nil {
			return nil, &net.OpError{Op: "dial", Net: "minecraft", Err: err}
		}
	}
	if d.ErrorLog == nil {
		d.ErrorLog = log.New(os.Stderr, "", log.LstdFlags)
	}
	if d.Protocol == nil {
		d.Protocol = proto{}
	}

	n, ok := networkByID(network)
	if !ok {
		return nil, fmt.Errorf("listen: no network under id: %v", network)
	}

	var pong []byte
	var netConn net.Conn
	if pong, err = n.PingContext(ctx, address); err == nil {
		netConn, err = n.DialContext(ctx, addressWithPongPort(pong, address))
	} else {
		netConn, err = n.DialContext(ctx, address)
	}
	if err != nil {
		return nil, err
	}

	conn = newConn(netConn, key, d.ErrorLog)
	conn.proto = d.Protocol
	conn.pool = conn.proto.Packets()
	conn.identityData = d.IdentityData
	conn.clientData = d.ClientData
	conn.packetFunc = d.PacketFunc
	conn.cacheEnabled = d.EnableClientCache

	// Disable the batch packet limit so that the server can send packets as often as it wants to.
	conn.dec.DisableBatchPacketLimit()

	defaultClientData(address, conn.identityData.DisplayName, &conn.clientData)
	defaultIdentityData(&conn.identityData)

	var request []byte
	if d.TokenSource == nil {
		// We haven't logged into the user's XBL account. We create a login request with only one token
		// holding the identity data set in the Dialer after making sure we clear data from the identity data
		// that is only present when logged in.
		if !d.KeepXBLIdentityData {
			clearXBLIdentityData(&conn.identityData)
		}
		request = login.EncodeOffline(conn.identityData, conn.clientData, key)
	} else {
		// We login as an Android device and this will show up in the 'titleId' field in the JWT chain, which
		// we can't edit. We just enforce Android data for logging in.
		setAndroidData(&conn.clientData)

		request = login.Encode(chainData, conn.clientData, key)
		identityData, _, _, _ := login.Parse(request)
		// If we got the identity data from Minecraft auth, we need to make sure we set it in the Conn too, as
		// we are not aware of the identity data ourselves yet.
		conn.identityData = identityData
	}
	c := make(chan struct{})
	go listenConn(conn, d.ErrorLog, c)

	conn.expect(packet.IDServerToClientHandshake, packet.IDPlayStatus)
	if err := conn.WritePacket(&packet.Login{ConnectionRequest: request, ClientProtocol: d.Protocol.ID()}); err != nil {
		return nil, err
	}
	_ = conn.Flush()
	select {
	case <-conn.close:
		return nil, conn.closeErr("dial")
	case <-ctx.Done():
		return nil, conn.wrap(ctx.Err(), "dial")
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
func authChain(ctx context.Context, src oauth2.TokenSource, key *ecdsa.PrivateKey) (string, error) {
	// Obtain the Live token, and using that the XSTS token.
	liveToken, err := src.Token()
	if err != nil {
		return "", fmt.Errorf("error obtaining Live Connect token: %v", err)
	}
	xsts, err := auth.RequestXBLToken(ctx, liveToken, "https://multiplayer.minecraft.net/")
	if err != nil {
		return "", fmt.Errorf("error obtaining XBOX Live token: %v", err)
	}

	// Obtain the raw chain data using the
	chain, err := auth.RequestMinecraftChain(ctx, xsts, key)
	if err != nil {
		return "", fmt.Errorf("error obtaining Minecraft auth chain: %v", err)
	}
	return chain, nil
}

// defaultClientData edits the ClientData passed to have defaults set to all fields that were left unchanged.
func defaultClientData(address, username string, d *login.ClientData) {
	rand2.Seed(time.Now().Unix())

	d.ServerAddress = address
	d.ThirdPartyName = username
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
		d.SkinResourcePatch = base64.StdEncoding.EncodeToString([]byte(resource.DefaultSkinResourcePatch))
	}
	if d.SkinGeometry == "" {
		d.SkinGeometry = base64.StdEncoding.EncodeToString([]byte(resource.DefaultSkinGeometry))
	}
}

// setAndroidData ensures the login.ClientData passed matches settings you would see on an Android device.
func setAndroidData(data *login.ClientData) {
	data.DeviceOS = protocol.DeviceAndroid
	data.GameVersion = protocol.CurrentVersion
}

// clearXBLIdentityData clears data from the login.IdentityData that is only set when a player is logged into
// XBOX Live.
func clearXBLIdentityData(data *login.IdentityData) {
	data.XUID = ""
	data.TitleID = ""
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

// splitPong splits the pong data passed by ;, taking into account escaping these.
func splitPong(s string) []string {
	var runes []rune
	var tokens []string
	inEscape := false
	for _, r := range s {
		switch {
		case r == '\\':
			inEscape = true
		case r == ';':
			tokens = append(tokens, string(runes))
			runes = runes[:0]
		case inEscape:
			inEscape = false
			fallthrough
		default:
			runes = append(runes, r)
		}
	}
	return append(tokens, string(runes))
}

// addressWithPongPort parses the redirect IPv4 port from the pong and returns the address passed with the port
// found if present, or the original address if not.
func addressWithPongPort(pong []byte, address string) string {
	frag := splitPong(string(pong))
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
