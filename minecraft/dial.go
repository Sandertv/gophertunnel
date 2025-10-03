package minecraft

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/internal"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/oauth2"
)

// Dialer allows specifying specific settings for connection to a Minecraft server.
// The zero value of Dialer is used for the package level Dial function.
type Dialer struct {
	// ErrorLog is a log.Logger that errors that occur during packet handling of
	// servers are written to. By default, errors are not logged.
	ErrorLog *slog.Logger

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

	// XBLToken should be used in place of TokenSource if the XBL token is already known, i.e through a different
	// oauth source. This token is for with the https://multiplayer.minecraft.net relaying party.
	XBLToken *auth.XBLToken

	// PacketFunc is called whenever a packet is read from or written to the connection returned when using
	// Dialer.Dial(). It includes packets that are otherwise covered in the connection sequence, such as the
	// Login packet. The function is called with the header of the packet and its raw payload, the address
	// from which the packet originated, and the destination address.
	PacketFunc func(header packet.Header, payload []byte, src, dst net.Addr)

	// DownloadResourcePack is called individually for every texture and behaviour pack sent by the connection when
	// using Dialer.Dial(), and can be used to stop the pack from being downloaded. The function is called with the UUID
	// and version of the resource pack, the number of the current pack being downloaded, and the total amount of packs.
	// The boolean returned determines if the pack will be downloaded or not.
	DownloadResourcePack func(id uuid.UUID, version string, current, total int) bool

	// DisconnectOnUnknownPackets specifies if the connection should disconnect if packets received are not present
	// in the packet pool. If true, such packets lead to the connection being closed immediately.
	// If set to false, the packets will be returned as a packet.Unknown.
	DisconnectOnUnknownPackets bool

	// DisconnectOnInvalidPackets specifies if invalid packets (either too few bytes or too many bytes) should be
	// allowed. If true, such packets lead to the connection being closed immediately. If false,
	// packets with too many bytes will be returned while packets with too few bytes will be skipped.
	DisconnectOnInvalidPackets bool

	// Protocol is the Protocol version used to communicate with the target server. By default, this field is
	// set to the current protocol as implemented in the minecraft/protocol package. Note that packets written
	// to and read from the Conn are always any of those found in the protocol/packet package, as packets
	// are converted from and to this Protocol.
	Protocol Protocol

	// FlushRate is the rate at which packets sent are flushed. Packets are buffered for a duration up to
	// FlushRate and are compressed/encrypted together to improve compression ratios. The lower this
	// time.Duration, the lower the latency but the less efficient both network and cpu wise.
	// The default FlushRate (when set to 0) is time.Second/20. If FlushRate is set negative, packets
	// will not be flushed automatically. In this case, calling `(*Conn).Flush()` is required after any
	// calls to `(*Conn).Write()` or `(*Conn).WritePacket()` to send the packets over network.
	FlushRate time.Duration

	// EnableClientCache, if set to true, enables the client blob cache for the client. This means that the
	// server will send chunks as blobs, which may be saved by the client so that chunks don't have to be
	// transmitted every time, resulting in less network transmission.
	EnableClientCache bool

	// KeepXBLIdentityData, if set to true, enables passing XUID and title ID to the target server
	// if the authentication token is not set. This is technically not valid and some servers might kick
	// the client when an XUID is present without logging in.
	// For getting this to work with BDS, authentication should be disabled.
	KeepXBLIdentityData bool

	// EnableLegacyAuth, if set to true, will use the legacy authentication behavior
	// (pre-1.21.90) when connecting to the server. This should only be used for outdated
	// servers, as enabling it will cause compatibility issues with updated servers.
	EnableLegacyAuth bool
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
	if d.ErrorLog == nil {
		d.ErrorLog = slog.New(internal.DiscardHandler{})
	}
	d.ErrorLog = d.ErrorLog.With("src", "dialer")
	if d.Protocol == nil {
		d.Protocol = DefaultProtocol
	}
	if d.FlushRate == 0 {
		d.FlushRate = time.Second / 20
	}

	key, err := ecdsa.GenerateKey(elliptic.P384(), cryptorand.Reader)
	if err != nil {
		return nil, &net.OpError{Op: "dial", Net: "minecraft", Err: fmt.Errorf("generating ECDSA key: %w", err)}
	}
	var chainData string
	if d.TokenSource != nil || d.XBLToken != nil {
		xblToken, err := getXBLToken(ctx, d)
		if err != nil {
			return nil, &net.OpError{Op: "dial", Net: "minecraft", Err: err}
		}
		chainData, err = authChain(ctx, xblToken, key)
		if err != nil {
			return nil, &net.OpError{Op: "dial", Net: "minecraft", Err: err}
		}
		identityData, err := readChainIdentityData([]byte(chainData))
		if err != nil {
			return nil, &net.OpError{Op: "dial", Net: "minecraft", Err: err}
		}
		d.IdentityData = identityData
	}

	n, ok := networkByID(network, d.ErrorLog)
	if !ok {
		return nil, &net.OpError{Op: "dial", Net: "minecraft", Err: fmt.Errorf("dial: no network under id %v", network)}
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

	conn = newConn(netConn, key, d.ErrorLog, d.Protocol, d.FlushRate, false)
	conn.pool = conn.proto.Packets(false)
	conn.identityData = d.IdentityData
	conn.clientData = d.ClientData
	conn.packetFunc = d.PacketFunc
	conn.downloadResourcePack = d.DownloadResourcePack
	conn.cacheEnabled = d.EnableClientCache
	conn.disconnectOnInvalidPacket = d.DisconnectOnInvalidPackets
	conn.disconnectOnUnknownPacket = d.DisconnectOnUnknownPackets
	conn.maxDecompressedLen = math.MaxInt

	defaultIdentityData(&conn.identityData)
	defaultClientData(address, conn.identityData.DisplayName, &conn.clientData)

	var request []byte
	if d.TokenSource == nil {
		// We haven't logged into the user's XBL account. We create a login request with only one token
		// holding the identity data set in the Dialer after making sure we clear data from the identity data
		// that is only present when logged in.
		if !d.KeepXBLIdentityData {
			clearXBLIdentityData(&conn.identityData)
		}
		request = login.EncodeOffline(conn.identityData, conn.clientData, key, d.EnableLegacyAuth)
	} else {
		// We login as an Android device and this will show up in the 'titleId' field in the JWT chain, which
		// we can't edit. We just enforce Android data for logging in.
		setAndroidData(&conn.clientData)

		request = login.Encode(chainData, conn.clientData, key, d.EnableLegacyAuth)
		identityData, _, _, _ := login.Parse(request)
		// If we got the identity data from Minecraft auth, we need to make sure we set it in the Conn too, as
		// we are not aware of the identity data ourselves yet.
		conn.identityData = identityData
	}

	readyForLogin, connected := make(chan struct{}), make(chan struct{})
	ctx, cancel := context.WithCancelCause(ctx)
	go listenConn(conn, readyForLogin, connected, cancel)

	conn.expect(packet.IDNetworkSettings, packet.IDPlayStatus)
	if err := conn.WritePacket(&packet.RequestNetworkSettings{ClientProtocol: d.Protocol.ID()}); err != nil {
		return nil, conn.wrap(fmt.Errorf("send request network settings: %w", err), "dial")
	}
	_ = conn.Flush()

	select {
	case <-ctx.Done():
		return nil, conn.wrap(context.Cause(ctx), "dial")
	case <-conn.ctx.Done():
		return nil, conn.closeErr("dial")
	case <-readyForLogin:
		// We've received our network settings, so we can now send our login request.
		conn.expect(packet.IDServerToClientHandshake, packet.IDPlayStatus)
		if err := conn.WritePacket(&packet.Login{ConnectionRequest: request, ClientProtocol: d.Protocol.ID()}); err != nil {
			return nil, conn.wrap(fmt.Errorf("send login: %w", err), "dial")
		}
		_ = conn.Flush()

		select {
		case <-ctx.Done():
			return nil, conn.wrap(context.Cause(ctx), "dial")
		case <-conn.ctx.Done():
			return nil, conn.closeErr("dial")
		case <-connected:
			// We've connected successfully. We return the connection and no error.
			return conn, nil
		}
	}
}

// readChainIdentityData reads a login.IdentityData from the Mojang chain
// obtained through authentication.
func readChainIdentityData(chainData []byte) (login.IdentityData, error) {
	chain := struct{ Chain []string }{}
	if err := json.Unmarshal(chainData, &chain); err != nil {
		return login.IdentityData{}, fmt.Errorf("read chain: read json: %w", err)
	}
	data := chain.Chain[1]
	claims := struct {
		ExtraData login.IdentityData `json:"extraData"`
	}{}
	tok, err := jwt.ParseSigned(data, []jose.SignatureAlgorithm{jose.ES384})
	if err != nil {
		return login.IdentityData{}, fmt.Errorf("read chain: parse jwt: %w", err)
	}
	if err := tok.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return login.IdentityData{}, fmt.Errorf("read chain: read claims: %w", err)
	}
	if claims.ExtraData.Identity == "" {
		return login.IdentityData{}, fmt.Errorf("read chain: no extra data found")
	}
	return claims.ExtraData, nil
}

// listenConn listens on the connection until it is closed on another goroutine. The channel passed will
// receive a value once the connection is logged in.
func listenConn(conn *Conn, readyForLogin, connected chan struct{}, cancel context.CancelCauseFunc) {
	defer func() {
		_ = conn.Close()
	}()
	cancelContext := true
	for {
		// We finally arrived at the packet decoding loop. We constantly decode packets that arrive
		// and push them to the Conn so that they may be processed.
		packets, err := conn.dec.Decode()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				if cancelContext {
					cancel(err)
				} else {
					conn.log.Error(err.Error())
				}
			}
			return
		}
		for _, data := range packets {
			loggedInBefore, readyToLoginBefore := conn.loggedIn, conn.readyToLogin
			if err := conn.receive(data); err != nil {
				if cancelContext {
					cancel(err)
				} else {
					conn.log.Error(err.Error())
				}
				return
			}
			if !readyToLoginBefore && conn.readyToLogin {
				// This is the signal that the connection is ready to login, so we put a value in the channel so that
				// it may be detected.
				readyForLogin <- struct{}{}
			}
			if !loggedInBefore && conn.loggedIn {
				// This is the signal that the connection was considered logged in, so we put a value in the channel so
				// that it may be detected.
				cancelContext = false
				connected <- struct{}{}
			}
		}
	}
}

func getXBLToken(ctx context.Context, dialer Dialer) (*auth.XBLToken, error) {
	if dialer.XBLToken != nil {
		return dialer.XBLToken, nil
	}

	liveToken, err := dialer.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("request Live Connect token: %w", err)
	}

	xblToken, err := auth.RequestXBLToken(ctx, liveToken, "https://multiplayer.minecraft.net/")
	if err != nil {
		return nil, fmt.Errorf("request XBOX Live token: %w", err)
	}

	return xblToken, nil
}

// authChain requests the Minecraft auth JWT chain using the credentials passed. If successful, an encoded
// chain ready to be put in a login request is returned.
func authChain(ctx context.Context, xblToken *auth.XBLToken, key *ecdsa.PrivateKey) (string, error) {
	// Obtain the raw chain data using the XBL token.
	chain, err := auth.RequestMinecraftChain(ctx, xblToken, key)
	if err != nil {
		return "", fmt.Errorf("request Minecraft auth chain: %w", err)
	}
	return chain, nil
}

//go:embed skin_resource_patch.json
var skinResourcePatch []byte

//go:embed skin_geometry.json
var skinGeometry []byte

// defaultClientData edits the ClientData passed to have defaults set to all fields that were left unchanged.
func defaultClientData(address, username string, d *login.ClientData) {
	d.ServerAddress = address
	d.ThirdPartyName = username
	if d.DeviceOS == 0 {
		d.DeviceOS = protocol.DeviceAndroid
	}
	if d.DefaultInputMode == 0 {
		d.DefaultInputMode = packet.InputModeTouch
	}
	if d.CurrentInputMode == 0 {
		d.CurrentInputMode = packet.InputModeTouch
	}
	if d.GameVersion == "" {
		d.GameVersion = protocol.CurrentVersion
	}
	if d.ClientRandomID == 0 {
		d.ClientRandomID = rand.Int63()
	}
	if d.DeviceID == "" {
		d.DeviceID = uuid.New().String()
	}
	if d.LanguageCode == "" {
		d.LanguageCode = "en_GB"
	}
	if d.PlayFabID == "" {
		id := make([]byte, 8)
		_, _ = cryptorand.Read(id)
		d.PlayFabID = hex.EncodeToString(id)
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
		d.SkinResourcePatch = base64.StdEncoding.EncodeToString(skinResourcePatch)
	}
	if d.SkinGeometry == "" {
		d.SkinGeometry = base64.StdEncoding.EncodeToString(skinGeometry)
	}
	if d.SkinGeometryVersion == "" {
		d.SkinGeometryVersion = base64.StdEncoding.EncodeToString([]byte("0.0.0"))
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
		// Vanilla (realms, in particular) will sometimes send port 19132 when you ping a port that isn't 19132 already,
		// but we should ignore that.
		if err != nil || port == 19132 {
			return address
		}
		// Remove the port from the address.
		addressParts := strings.Split(address, ":")
		address = strings.Join(strings.Split(address, ":")[:len(addressParts)-1], ":")
		return address + ":" + portStr
	}
	return address
}
