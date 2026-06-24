package minecraft

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/internal"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

// exemptedResourcePack is a resource pack that is exempted from being downloaded. These packs may be directly
// applied by sending them in the ResourcePackStack packet.
type exemptedResourcePack struct {
	uuid    string
	version string
}

// exemptedPacks is a list of all resource packs that do not need to be downloaded, but may always be applied
// in the ResourcePackStack packet.
var exemptedPacks = []exemptedResourcePack{
	{
		uuid:    "0fba4063-dba1-4281-9b89-ff9390653530",
		version: "1.0.0",
	},
	{
		uuid:    "b41c2785-c512-4a49-af56-3a87afd47c57",
		version: "1.21.30",
	},
	{
		uuid:    "a4df0cb3-17be-4163-88d7-fcf7002b935d",
		version: "1.21.20",
	},
	{
		uuid:    "d19adffe-a2e1-4b02-8436-ca4583368c89",
		version: "1.21.10",
	},
	{
		uuid:    "85d5603d-2824-4b21-8044-34f441f4fce1",
		version: "1.21.0",
	},
	{
		uuid:    "e977cd13-0a11-4618-96fb-03dfe9c43608",
		version: "1.20.60",
	},
	{
		uuid:    "0674721c-a0aa-41a1-9ba8-1ed33ea3e7ed",
		version: "1.20.50",
	},
}

var disconnectReasons = map[int32]string{
	packet.DisconnectReasonUnknown:                                       "Unknown",
	packet.DisconnectReasonCantConnectNoInternet:                         "Please check your connection to the internet and try again.",
	packet.DisconnectReasonNoPermissions:                                 "You're not invited to play on this server.",
	packet.DisconnectReasonUnrecoverableError:                            "An unrecoverable error has occurred.",
	packet.DisconnectReasonThirdPartyBlocked:                             "Third-party server is blocked.",
	packet.DisconnectReasonThirdPartyNoInternet:                          "Please check your connection to the internet and try again.",
	packet.DisconnectReasonThirdPartyBadIP:                               "Invalid IP address.",
	packet.DisconnectReasonThirdPartyNoServerOrServerLocked:              "The server you are attempting to join may not exist or be locked.",
	packet.DisconnectReasonVersionMismatch:                               "Version mismatch",
	packet.DisconnectReasonSkinIssue:                                     "There is an issue with your skin.",
	packet.DisconnectReasonInviteSessionNotFound:                         "Unable to connect to world. The world is no longer available to join.",
	packet.DisconnectReasonEduLevelSettingsMissing:                       "This world was saved from Minecraft Education. It cannot be loaded.",
	packet.DisconnectReasonLocalServerNotFound:                           "Local server not found.",
	packet.DisconnectReasonLegacyDisconnect:                              "Disconnected by server.",
	packet.DisconnectReasonUserLeaveGameAttempted:                        "Quitting",
	packet.DisconnectReasonPlatformLockedSkinsError:                      "Platform Restricted Skin!",
	packet.DisconnectReasonRealmsWorldUnassigned:                         "This Realm has no world assigned.",
	packet.DisconnectReasonRealmsServerCantConnect:                       "Unable to connect to Realm.",
	packet.DisconnectReasonRealmsServerHidden:                            "Multiplayer Invitation",
	packet.DisconnectReasonRealmsServerDisabledBeta:                      "Realms are disabled for the beta.",
	packet.DisconnectReasonRealmsServerDisabled:                          "Realms are disabled.",
	packet.DisconnectReasonCrossPlatformDisabled:                         "Cross-Platform Play Disabled.",
	packet.DisconnectReasonCantConnect:                                   "Unable to connect to world.",
	packet.DisconnectReasonSessionNotFound:                               "Unable to connect to world. The world is no longer available to join.",
	packet.DisconnectReasonServerFull:                                    "Server Full",
	packet.DisconnectReasonInvalidPlatformSkin:                           "Invalid or corrupt skin!",
	packet.DisconnectReasonEditionVersionMismatch:                        "Unable to load world.",
	packet.DisconnectReasonEditionMismatch:                               "This world was saved from Minecraft Education. It cannot be loaded.",
	packet.DisconnectReasonLevelNewerThanExeVersion:                      "A newer version of the game has saved this world. It cannot be loaded.",
	packet.DisconnectReasonNoFailOccurred:                                "No failure occurred.",
	packet.DisconnectReasonBannedSkin:                                    "Skin Not Allowed In Multiplayer",
	packet.DisconnectReasonTimeout:                                       "Timed out",
	packet.DisconnectReasonServerNotFound:                                "Server not found.",
	packet.DisconnectReasonOutdatedServer:                                "The host is using an older version of Minecraft. Everyone should update to the latest version of Minecraft and try again.",
	packet.DisconnectReasonOutdatedClient:                                "Could not connect: Outdated client!",
	packet.DisconnectReasonMultiplayerDisabled:                           "The world has been set to single player mode.",
	packet.DisconnectReasonNoWiFi:                                        "No WiFi Connection",
	packet.DisconnectReasonNoReason:                                      "Disconnected",
	packet.DisconnectReasonDisconnected:                                  "Disconnected by Server",
	packet.DisconnectReasonInvalidPlayer:                                 "This world's multiplayer setting is set to friends only. You must be friends with the host of this world to join.",
	packet.DisconnectReasonLoggedInOtherLocation:                         "Logged in from other location",
	packet.DisconnectReasonServerIdConflict:                              "Cannot join world. The account you are signed in to is currently playing in this world on a different device.",
	packet.DisconnectReasonNotAllowed:                                    "You're not invited to play on this server.",
	packet.DisconnectReasonNotAuthenticated:                              "You need to authenticate to Microsoft services.",
	packet.DisconnectReasonInvalidTenant:                                 "Unable to connect to the world. Please check your join code and try again.",
	packet.DisconnectReasonUnknownPacket:                                 "Unknown packet",
	packet.DisconnectReasonUnexpectedPacket:                              "Unexpected packet",
	packet.DisconnectReasonInvalidCommandRequestPacket:                   "Invalid command request packet",
	packet.DisconnectReasonHostSuspended:                                 "The host has been suspended.",
	packet.DisconnectReasonLoginPacketNoRequest:                          "Login packet with no request",
	packet.DisconnectReasonLoginPacketNoCert:                             "Login packet with no certificate",
	packet.DisconnectReasonMissingClient:                                 "Missing client",
	packet.DisconnectReasonKicked:                                        "You were kicked from the game",
	packet.DisconnectReasonKickedForExploit:                              "You were kicked from the game for exploiting.",
	packet.DisconnectReasonKickedForIdle:                                 "You were kicked for being idle.",
	packet.DisconnectReasonResourcePackProblem:                           "Encountered a problem while downloading or applying resource pack.",
	packet.DisconnectReasonIncompatiblePack:                              "You are unable to join the world because you have an incompatible pack.",
	packet.DisconnectReasonOutOfStorage:                                  "Out of storage space",
	packet.DisconnectReasonInvalidLevel:                                  "Invalid Level!",
	packet.DisconnectReasonBlockMismatch:                                 "Block mismatch",
	packet.DisconnectReasonInvalidHeights:                                "Invalid heights",
	packet.DisconnectReasonInvalidWidths:                                 "Invalid widths",
	packet.DisconnectReasonShutdown:                                      "Quitting",
	packet.DisconnectReasonLoadingStateTimeout:                           "Timed out while loading",
	packet.DisconnectReasonResourcePackLoadingFailed:                     "Failed to load resource pack",
	packet.DisconnectReasonSearchingForSessionLoadingScreenFailed:        "Failed to find session",
	packet.DisconnectReasonNetherNetProtocolVersion:                      "Incompatible NetherNet protocol version",
	packet.DisconnectReasonSubsystemStatusError:                          "Subsystem status error",
	packet.DisconnectReasonEmptyAuthFromDiscovery:                        "Empty auth from discovery",
	packet.DisconnectReasonEmptyUrlFromDiscovery:                         "Empty URL from discovery",
	packet.DisconnectReasonExpiredAuthFromDiscovery:                      "Expired auth from discovery",
	packet.DisconnectReasonUnknownSignalServiceSignInFailure:             "Unknown signal service sign in failure",
	packet.DisconnectReasonXBLJoinLobbyFailure:                           "XBOX Live join lobby failure",
	packet.DisconnectReasonUnspecifiedClientInstanceDisconnection:        "Unspecified client instance disconnection",
	packet.DisconnectReasonNetherNetSessionNotFound:                      "NetherNet session not found",
	packet.DisconnectReasonNetherNetCreatePeerConnection:                 "NetherNet failed to create peer connection",
	packet.DisconnectReasonNetherNetICE:                                  "NetherNet ICE error",
	packet.DisconnectReasonNetherNetConnectRequest:                       "NetherNet connect request error",
	packet.DisconnectReasonNetherNetConnectResponse:                      "NetherNet connect response error",
	packet.DisconnectReasonNetherNetNegotiationTimeout:                   "NetherNet negotiation timed out",
	packet.DisconnectReasonNetherNetInactivityTimeout:                    "NetherNet inactivity timed out",
	packet.DisconnectReasonStaleConnectionBeingReplaced:                  "Stale connection being replaced",
	packet.DisconnectReasonBadPacket:                                     "Server sent broken packet.",
	packet.DisconnectReasonNetherNetFailedToCreateOffer:                  "NetherNet failed to create offer",
	packet.DisconnectReasonNetherNetFailedToCreateAnswer:                 "NetherNet failed to create answer",
	packet.DisconnectReasonNetherNetFailedToSetLocalDescription:          "NetherNet failed to set local description",
	packet.DisconnectReasonNetherNetFailedToSetRemoteDescription:         "NetherNet failed to set remote description",
	packet.DisconnectReasonNetherNetNegotiationTimeoutWaitingForResponse: "NetherNet negotiation timed out waiting for response",
	packet.DisconnectReasonNetherNetNegotiationTimeoutWaitingForAccept:   "NetherNet negotiation timed out waiting for accept",
	packet.DisconnectReasonNetherNetIncomingConnectionIgnored:            "NetherNet incoming connection ignored",
	packet.DisconnectReasonNetherNetSignalingParsingFailure:              "NetherNet signaling parsing failure",
	packet.DisconnectReasonNetherNetSignalingUnknownError:                "NetherNet signaling unknown error",
	packet.DisconnectReasonNetherNetSignalingUnicastDeliveryFailed:       "NetherNet signaling unicast delivery failed",
	packet.DisconnectReasonNetherNetSignalingBroadcastDeliveryFailed:     "NetherNet signaling broadcast delivery failed",
	packet.DisconnectReasonNetherNetSignalingGenericDeliveryFailed:       "NetherNet signaling generic delivery failed",
	packet.DisconnectReasonEditorMismatchEditorWorld:                     "This world is in Editor Mode. It cannot be loaded.",
	packet.DisconnectReasonEditorMismatchVanillaWorld:                    "This world is a not in Editor Mode. It cannot be loaded.",
	packet.DisconnectReasonWorldTransferNotPrimaryClient:                 "World transfer not primary client",
	packet.DisconnectReasonRequestServerShutdown:                         "Server shutdown",
	packet.DisconnectReasonClientGameSetupCancelled:                      "Game setup cancelled",
	packet.DisconnectReasonClientGameSetupFailed:                         "Game setup failed",
	packet.DisconnectReasonNetherNetSignalingSigninFailed:                "NetherNet signaling sign in failed",
	packet.DisconnectReasonSessionAccessDenied:                           "Session access denied",
	packet.DisconnectReasonServiceSigninIssue:                            "Service sign in issue",
	packet.DisconnectReasonNetherNetNoSignalingChannel:                   "NetherNet no signaling channel",
	packet.DisconnectReasonNetherNetNotLoggedIn:                          "NetherNet not logged in",
	packet.DisconnectReasonNetherNetClientSignalingError:                 "NetherNet client signaling error",
	packet.DisconnectReasonSubClientLoginDisabled:                        "Sub-client login disabled",
	packet.DisconnectReasonDeepLinkTryingToOpenDemoWorldWhileSignedIn:    "Deep link trying to open demo world while signed in",
	packet.DisconnectReasonAsyncJoinTaskDenied:                           "Async join task denied",
	packet.DisconnectReasonRealmsTimelineRequired:                        "Realms timeline required",
	packet.DisconnectReasonGuestWithoutHost:                              "Guest without host",
	packet.DisconnectReasonFailedToJoinExperience:                        "Failed to join experience",
	packet.DisconnectReasonNetherNetDataChannelClosed:                    "NetherNet data channel closed",
	packet.DisconnectReasonDiscoveryEnvironmentMismatch:                  "Discovery environment mismatch",
	packet.DisconnectReasonHostWithoutKeys:                               "The host is using offline mode.",
	packet.DisconnectReasonHostSignedOut:                                 "The host is signed out",
	packet.DisconnectReasonScriptWatchdogException:                       "The server was shut down due to an unhandled scripting watchdog exception.",
	packet.DisconnectReasonScriptMemoryLimitExceeded:                     "The server was shut down due to exceeding the scripting memory limit.",
	packet.DisconnectReasonStorageLowDuringGameplay:                      "Your device is almost out of the space that Minecraft can use to save worlds and settings on this device. Why not delete some old stuff you don't need so that you can keep saving new stuff?",
	packet.DisconnectReasonStorageFullDuringGameplay:                     "You are out of data storage space and Minecraft is unable to save your progress! Minecraft will return you to the Main Menu to clear up storage space.",
	packet.DisconnectReasonLevelStorageCorruption:                        "Something went wrong while preparing to upload your world. If this keeps happening, restarting your device may help.",
	packet.DisconnectReasonEditionMismatchVanillaToEdu:                   "The server is running an incompatible edition of Minecraft. Failed to connect.",
	packet.DisconnectReasonEditionMismatchEduToVanilla:                   "The server is not running Minecraft Education. Failed to connect.",
	packet.DisconnectReasonEditorMismatchEditorToVanilla:                 "The server is not in Editor Mode. Failed to connect.",
	packet.DisconnectReasonEditorMismatchVanillaToEditor:                 "The server is in Editor Mode. Failed to connect.",
	packet.DisconnectReasonDenyListed:                                    "You are in deny list.",
}

// Conn represents a Minecraft (Bedrock Edition) connection over a specific net.Conn transport layer. Its
// methods (Read, Write etc.) are safe to be called from multiple goroutines simultaneously, but ReadPacket
// must not be called on multiple goroutines simultaneously.
type Conn struct {
	// once is used to ensure the Conn is closed only a single time. It protects the channel below from being
	// closed multiple times.
	once       sync.Once
	ctx        context.Context
	cancelFunc context.CancelCauseFunc

	conn        net.Conn
	log         *slog.Logger
	authEnabled bool

	proto                Protocol
	acceptedProto        []Protocol
	pool                 packet.Pool
	enc                  *packet.Encoder
	dec                  *packet.Decoder
	compression          packet.Compression
	compressionSelector  func(proto Protocol) packet.Compression
	compressionThreshold int
	maxDecompressedLen   int
	readerLimits         bool

	disconnectOnUnknownPacket bool
	disconnectOnInvalidPacket bool

	identityData login.IdentityData
	clientData   login.ClientData

	gameData         GameData
	gameDataReceived atomic.Bool

	// privateKey is the private key of this end of the connection. Each connection, regardless of which side
	// the connection is on, server or client, has a unique private key generated.
	privateKey *ecdsa.PrivateKey
	// salt is a 16 byte long randomly generated byte slice which is only used if the Conn is a server sided
	// connection. It is otherwise left unused.
	salt              []byte
	disableEncryption bool
	// verifier verifies the OpenID token encapsulated in the first chain of
	// the Login packet sent from the connection. If nil, the legacy chain will
	// be instead used for authentication.
	verifier *oidc.IDTokenVerifier

	// packets is a channel of byte slices containing serialised packets that are coming in from the other
	// side of the connection.
	packets chan *packetData

	deferredPacketMu sync.Mutex
	// deferredPackets is a list of packets that were pushed back during the login sequence because they
	// were not used by the connection yet. These packets are read the first when calling to Read or
	// ReadPacket after being connected.
	deferredPackets []*packetData
	readDeadline    <-chan time.Time

	// sendMu protects bufferedSend/bufferedSendSpare.
	sendMu sync.Mutex
	// encMu serializes encoder state changes and network writes (enc.Encode).
	// Lock order (when both are needed): encMu → sendMu.
	encMu sync.Mutex
	// bufferedSend is a slice of byte slices containing packets that are 'written'. They are buffered until
	// they are sent each 20th of a second.
	bufferedSend      [][]byte
	bufferedSendSpare [][]byte
	hdr               *packet.Header

	// readyToLogin is a bool indicating if the connection is ready to login. This is used to ensure that the client
	// has received the relevant network settings before the login sequence starts.
	readyToLogin bool
	// handshakeComplete is true if the login handshake has been completed.
	handshakeComplete bool
	// loggedIn is a bool indicating if the connection was logged in. It is set to true after the entire login
	// sequence is completed.
	loggedIn bool
	// spawn is a bool channel indicating if the connection is currently waiting for its spawning in
	// the world: It is completing a sequence that will result in the spawning.
	spawn           chan struct{}
	waitingForSpawn atomic.Bool

	// expectedIDs is a slice of packet identifiers that are next expected to arrive, until the connection is
	// logged in.
	expectedIDs atomic.Value

	packMu sync.Mutex
	// resourcePacks is a slice of resource packs that the listener may hold. Each client will be asked to
	// download these resource packs upon joining.
	resourcePacks []*resource.Pack
	// texturePacksRequired specifies if clients that join must accept the texture pack in order for them to
	// be able to join the server. If they don't accept, they can only leave the server.
	texturePacksRequired bool
	// forceDisableVibrantVisuals specifies whether the connection is forced to have vibrant visuals disabled.
	forceDisableVibrantVisuals bool
	packQueue                  *resourcePackQueue
	// downloadResourcePack is an optional function passed to a Dial() call. If set, each resource pack received
	// from the server will call this function to see if it should be downloaded or not.
	downloadResourcePack func(id uuid.UUID, version string, currentPack, totalPacks int) bool
	// fetchResourcePacks is an optional function passed to a Listener. If set, the returned resource packs from the function
	// will determine which resource packs to send to the client based on its identity and client data.
	fetchResourcePacks func(identityData login.IdentityData, clientData login.ClientData, current []*resource.Pack) []*resource.Pack
	// ignoredResourcePacks is a slice of resource packs that are not being downloaded due to the downloadResourcePack
	// func returning false for the specific pack.
	ignoredResourcePacks []exemptedResourcePack

	cacheEnabled bool

	// packetFunc is an optional function passed to a Dial() call. If set, each packet read from and written
	// to this connection will call this function.
	packetFunc func(header packet.Header, payload []byte, src, dst net.Addr)

	shieldID atomic.Int32

	additional chan packet.Packet

	disablePacketHandling bool
	// disablePacketHandlingReady indicates that the connection should now forward packets directly to the caller
	// when disablePacketHandling is enabled. This becomes true after handshake completion or once post-login
	// packets start arriving on servers that skip the handshake packet.
	disablePacketHandlingReady bool
}

// newConn creates a new Minecraft connection for the net.Conn passed, reading and writing compressed
// Minecraft packets to that net.Conn.
// newConn accepts a private key which will be used to identify the connection. If a nil key is passed, the
// key is generated.
func newConn(netConn net.Conn, key *ecdsa.PrivateKey, log *slog.Logger, proto Protocol, flushRate time.Duration, limits bool) *Conn {
	disableEncryption := false
	if d, ok := netConn.(interface{ DisableEncryption() bool }); ok {
		disableEncryption = d.DisableEncryption()
	}

	conn := &Conn{
		salt:                 make([]byte, 16),
		disableEncryption:    disableEncryption,
		packets:              make(chan *packetData, 8),
		additional:           make(chan packet.Packet, 16),
		spawn:                make(chan struct{}),
		conn:                 netConn,
		privateKey:           key,
		log:                  log.With("raddr", netConn.RemoteAddr().String()),
		hdr:                  &packet.Header{},
		proto:                proto,
		readerLimits:         limits,
		compressionThreshold: 256,
	}
	conn.enc = packet.NewEncoder(netConn)
	conn.dec = packet.NewDecoder(netConn)

	if c, ok := netConn.(interface{ Context() context.Context }); ok {
		conn.ctx, conn.cancelFunc = context.WithCancelCause(c.Context())
	} else {
		conn.ctx, conn.cancelFunc = context.WithCancelCause(context.Background())
	}

	if !limits {
		// Disable the batch packet limit so that the server can send packets as often as it wants to.
		conn.dec.DisableBatchPacketLimit()
	}
	_, _ = rand.Read(conn.salt)

	conn.expectedIDs.Store([]uint32{packet.IDRequestNetworkSettings})

	if flushRate <= 0 {
		return conn
	}
	go func() {
		ticker := time.NewTicker(flushRate)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.Flush(); err != nil {
				_ = conn.close(err)
				return
			}
		}
	}()
	return conn
}

// IdentityData returns the identity data of the connection. It holds the UUID, XUID and username of the
// connected client.
func (conn *Conn) IdentityData() login.IdentityData {
	return conn.identityData
}

// ClientData returns the client data the client connected with. Note that this client data may be changed
// during the session, so the data should only be used directly after connection, and should be updated after
// that by the caller.
func (conn *Conn) ClientData() login.ClientData {
	return conn.clientData
}

// SetPacketBatchFunc sets a callback called after each outbound packet batch is
// encoded. Passing nil disables the callback.
func (conn *Conn) SetPacketBatchFunc(f packet.BatchEncodeObserver) {
	conn.encMu.Lock()
	defer conn.encMu.Unlock()
	conn.enc.SetBatchEncodeObserver(f)
}

// Authenticated returns true if the connection was authenticated through XBOX Live services.
func (conn *Conn) Authenticated() bool {
	return conn.IdentityData().XUID != ""
}

// GameData returns specific game data set to the connection for the player to be initialised with. If the
// Conn is obtained using Listen, this game data may be set to the Listener. If obtained using Dial, the data
// is obtained from the server.
func (conn *Conn) GameData() GameData {
	return conn.gameData
}

// Proto returns the protocol of the connection.
func (conn *Conn) Proto() Protocol {
	return conn.proto
}

// StartGame starts the game for a client that connected to the server. StartGame should be called for a Conn
// obtained using a minecraft.Listener. The game data passed will be used to spawn the player in the world of
// the server. To spawn a Conn obtained from a call to minecraft.Dial(), use Conn.DoSpawn().
func (conn *Conn) StartGame(data GameData) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return conn.StartGameContext(ctx, data)
}

// StartGameTimeout starts the game for a client that connected to the server, returning an error if the
// connection is not yet fully connected while the timeout expires.
// StartGameTimeout should be called for a Conn obtained using a minecraft.Listener. The game data passed will
// be used to spawn the player in the world of the server. To spawn a Conn obtained from a call to
// minecraft.Dial(), use Conn.DoSpawn().
func (conn *Conn) StartGameTimeout(data GameData, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return conn.StartGameContext(ctx, data)
}

// StartGameContext starts the game for a client that connected to the server, returning an error if the
// context is closed while spawning the client.
// StartGameContext should be called for a Conn obtained using a minecraft.Listener. The game data passed will
// be used to spawn the player in the world of the server. To spawn a Conn obtained from a call to
// minecraft.Dial(), use Conn.DoSpawn().
func (conn *Conn) StartGameContext(ctx context.Context, data GameData) error {
	if err := conn.SendStartGame(data); err != nil {
		return err
	}

	select {
	case <-conn.ctx.Done():
		return conn.closeErr("start game")
	case <-ctx.Done():
		return conn.wrap(ctx.Err(), "start game")
	case <-conn.spawn:
		// Conn was spawned successfully.
		return nil
	}
}

// SendStartGame sends the packets that start a game for a client connected to a Listener without waiting for
// the client to finish spawning. Most callers should use StartGame instead, which waits until the client sends
// its spawn acknowledgement.
func (conn *Conn) SendStartGame(data GameData) error {
	if conn.gameDataReceived.Load() {
		panic("(*Conn).SendStartGame must only be called on Listener connections")
	}
	if data.WorldName == "" {
		data.WorldName = conn.gameData.WorldName
	}

	conn.gameData = data
	for _, item := range data.Items {
		if item.Name == "minecraft:shield" {
			conn.shieldID.Store(int32(item.RuntimeID))
		}
	}
	conn.waitingForSpawn.Store(true)
	return conn.startGame()
}

// DoSpawn starts the game for the client in the server. DoSpawn should be called for a Conn obtained using
// minecraft.Dial(). Use Conn.StartGame to spawn a Conn obtained using a minecraft.Listener.
// DoSpawn will start the spawning sequence using the game data found in conn.GameData(), which was sent
// earlier by the server.
// DoSpawn has a default timeout of 1 minute. DoSpawnContext or DoSpawnTimeout may be used for cancellation
// at any other times.
func (conn *Conn) DoSpawn() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return conn.DoSpawnContext(ctx)
}

// DoSpawnTimeout starts the game for the client in the server with a timeout after which an error is
// returned if the client has not yet spawned by that time. DoSpawnTimeout should be called for a Conn
// obtained using minecraft.Dial(). Use Conn.StartGame to spawn a Conn obtained using a minecraft.Listener.
// DoSpawnTimeout will start the spawning sequence using the game data found in conn.GameData(), which was
// sent earlier by the server.
func (conn *Conn) DoSpawnTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return conn.DoSpawnContext(ctx)
}

// DoSpawnContext starts the game for the client in the server, using a specific context for cancellation.
// DoSpawnContext should be called for a Conn obtained using minecraft.Dial(). Use Conn.StartGame to spawn a
// Conn obtained using a minecraft.Listener.
// DoSpawnContext will start the spawning sequence using the game data found in conn.GameData(), which was
// sent earlier by the server.
func (conn *Conn) DoSpawnContext(ctx context.Context) error {
	select {
	case <-conn.ctx.Done():
		return conn.closeErr("do spawn")
	case <-ctx.Done():
		return conn.wrap(ctx.Err(), "do spawn")
	case <-conn.spawn:
		// Conn was spawned successfully.
		return nil
	}
}

// WritePacket encodes the packet passed and writes it to the Conn. The encoded data is buffered until the
// next 20th of a second, after which the data is flushed and sent over the connection.
func (conn *Conn) WritePacket(pk packet.Packet) error {
	select {
	case <-conn.ctx.Done():
		return conn.closeErr("write packet")
	default:
	}
	conn.sendMu.Lock()
	defer conn.sendMu.Unlock()

	conn.encodePacketsTo(&conn.bufferedSend, pk)
	return nil
}

// encodePacketsTo marshals the provided packet (including header) into one or more byte slices,
// accounting for protocol conversions and invoking packetFunc callbacks. The resulting byte slices are
// appended to dst. The appended slices are copies safe to retain beyond the call.
func (conn *Conn) encodePacketsTo(dst *[][]byte, pks ...packet.Packet) {
	buf := internal.BufferPool.Get().(*bytes.Buffer)
	defer func() {
		// Reset the buffer, so we can return it to the buffer pool safely.
		buf.Reset()
		internal.BufferPool.Put(buf)
	}()

	for _, pk := range pks {
		for _, converted := range conn.proto.ConvertFromLatest(pk, conn) {
			buf.Reset()
			conn.hdr.PacketID = converted.ID()
			_ = conn.hdr.Write(buf)
			l := buf.Len()

			converted.Marshal(conn.proto.NewWriter(buf, conn.shieldID.Load()))
			if conn.packetFunc != nil {
				conn.packetFunc(*conn.hdr, buf.Bytes()[l:], conn.LocalAddr(), conn.RemoteAddr())
			}
			*dst = append(*dst, append([]byte(nil), buf.Bytes()...))
		}
	}
}

// WritePacketImmediate encodes the packets passed, queues them in the normal buffered send queue and flushes
// that queue immediately. This preserves ordering relative to packets that were already queued through
// WritePacket while still sending the data right away.
func (conn *Conn) WritePacketImmediate(pks ...packet.Packet) error {
	select {
	case <-conn.ctx.Done():
		return conn.closeErr("write immediate packet")
	default:
	}

	conn.sendMu.Lock()
	conn.encodePacketsTo(&conn.bufferedSend, pks...)
	conn.sendMu.Unlock()

	return conn.Flush()
}

// WritePacketDirect encodes the packet passed and writes it immediately to the underlying connection,
// bypassing the buffered batch that is flushed every tick.
// Use this only when packet ordering relative to already-buffered packets does not matter.
func (conn *Conn) WritePacketDirect(pks ...packet.Packet) error {
	select {
	case <-conn.ctx.Done():
		return conn.closeErr("write packet direct")
	default:
	}
	// Use a small stack-allocated buffer for the common case (usually 1 slice),
	// allowing append to spill to heap only if more capacity is needed.
	var stackBuf [4][]byte
	immediate := stackBuf[:0]

	conn.sendMu.Lock()
	conn.encodePacketsTo(&immediate, pks...)
	conn.sendMu.Unlock()

	if len(immediate) > 0 {
		conn.encMu.Lock()
		defer conn.encMu.Unlock()
		if err := conn.enc.Encode(immediate); err != nil && !errors.Is(err, net.ErrClosed) {
			// Should never happen.
			panic(fmt.Errorf("error encoding packet batch: %w", err))
		}
	}
	return nil
}

// ReadPacket reads a packet from the Conn, depending on the packet ID that is found in front of the packet
// data. If a read deadline is set, an error is returned if the deadline is reached before any packet is
// received. ReadPacket must not be called on multiple goroutines simultaneously.
//
// If the packet read was not implemented, a *packet.Unknown is returned, containing the raw payload of the
// packet read.
func (conn *Conn) ReadPacket() (pk packet.Packet, err error) {
	if len(conn.additional) > 0 {
		return <-conn.additional, nil
	}
	if data, ok := conn.takeDeferredPacket(); ok {
		pk, err := data.decode(conn)
		if err != nil {
			conn.log.Error("read packet: " + err.Error())
			return conn.ReadPacket()
		}
		if len(pk) == 0 {
			return conn.ReadPacket()
		}
		for _, additional := range pk[1:] {
			conn.additional <- additional
		}
		return pk[0], nil
	}

	select {
	case <-conn.ctx.Done():
		return nil, conn.closeErr("read packet")
	case <-conn.readDeadline:
		return nil, conn.wrap(context.DeadlineExceeded, "read packet")
	case data := <-conn.packets:
		pk, err := data.decode(conn)
		if err != nil {
			conn.log.Error("read packet: " + err.Error())
			return conn.ReadPacket()
		}
		if len(pk) == 0 {
			return conn.ReadPacket()
		}
		for _, additional := range pk[1:] {
			conn.additional <- additional
		}
		return pk[0], nil
	}
}

// ResourcePacks returns a slice of all resource packs the connection holds. For a Conn obtained using a
// Listener, this holds all resource packs set to the Listener. For a Conn obtained using Dial, the resource
// packs include all packs sent by the server connected to.
func (conn *Conn) ResourcePacks() []*resource.Pack {
	return conn.resourcePacks
}

// Write writes a slice of serialised packet data to the Conn. The data is buffered until the next 20th of a
// tick, after which it is flushed to the connection. Write returns the amount of bytes written n.
func (conn *Conn) Write(b []byte) (n int, err error) {
	conn.sendMu.Lock()
	defer conn.sendMu.Unlock()

	conn.bufferedSend = append(conn.bufferedSend, b)
	return len(b), nil
}

// ReadBytes reads a packet from the connection without decoding it directly.
// For direct reading, consider using ReadPacket() which decodes the packet.
func (conn *Conn) ReadBytes() ([]byte, error) {
	if data, ok := conn.takeDeferredPacket(); ok {
		return data.full, nil
	}
	select {
	case <-conn.ctx.Done():
		return nil, conn.closeErr("read")
	case <-conn.readDeadline:
		return nil, conn.wrap(context.DeadlineExceeded, "read")
	case data := <-conn.packets:
		return data.full, nil
	}
}

// Read reads a packet from the connection into the byte slice passed, provided the byte slice is big enough
// to carry the full packet.
// It is recommended to use ReadPacket() and ReadBytes() rather than Read() in cases where reading is done directly.
func (conn *Conn) Read(b []byte) (n int, err error) {
	if data, ok := conn.takeDeferredPacket(); ok {
		if len(b) < len(data.full) {
			return 0, conn.wrap(errBufferTooSmall, "read")
		}
		return copy(b, data.full), nil
	}
	select {
	case <-conn.ctx.Done():
		return 0, conn.closeErr("read")
	case <-conn.readDeadline:
		return 0, conn.wrap(context.DeadlineExceeded, "read")
	case data := <-conn.packets:
		if len(b) < len(data.full) {
			return 0, conn.wrap(errBufferTooSmall, "read")
		}
		return copy(b, data.full), nil
	}
}

// Flush flushes the packets currently buffered by the connections to the underlying net.Conn, so that they
// are directly sent.
func (conn *Conn) Flush() error {
	select {
	case <-conn.ctx.Done():
		return conn.closeErr("flush")
	default:
	}

	conn.encMu.Lock()
	defer conn.encMu.Unlock()

	conn.sendMu.Lock()
	if len(conn.bufferedSend) == 0 {
		conn.sendMu.Unlock()
		return nil
	}

	// Detach the current buffer and swap in the spare so writers can keep appending while we encode,
	// without reallocating bufferedSend.
	toSend := conn.bufferedSend
	conn.bufferedSend = conn.bufferedSendSpare[:0]
	conn.bufferedSendSpare = nil
	conn.sendMu.Unlock()

	if err := conn.enc.Encode(toSend); err != nil && !errors.Is(err, net.ErrClosed) {
		// Should never happen.
		panic(fmt.Errorf("error encoding packet batch: %w", err))
	}

	// Clear out toSend so that re-using the slice after resetting its length to 0 doesn't keep references
	// to packet payloads alive, causing an 'invisible' memory leak.
	for i := range toSend {
		toSend[i] = nil
	}

	conn.sendMu.Lock()
	conn.bufferedSendSpare = toSend[:0]
	conn.sendMu.Unlock()
	return nil
}

// Close closes the Conn and its underlying connection. Before closing, it also calls Flush() so that any
// packets currently pending are sent out.
func (conn *Conn) Close() error {
	return conn.close(net.ErrClosed)
}

// LocalAddr returns the local address of the underlying connection.
func (conn *Conn) LocalAddr() net.Addr {
	return conn.conn.LocalAddr()
}

// RemoteAddr returns the remote address of the underlying connection.
func (conn *Conn) RemoteAddr() net.Addr {
	return conn.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadline of the connection. It is equivalent to calling SetReadDeadline
// and SetWriteDeadline at the same time.
func (conn *Conn) SetDeadline(t time.Time) error {
	return conn.SetReadDeadline(t)
}

// SetReadDeadline sets the read deadline of the Conn to the time passed. The time must be after time.Now().
// Passing an empty time.Time to the method (time.Time{}) results in the read deadline being cleared.
func (conn *Conn) SetReadDeadline(t time.Time) error {
	if t.Equal(time.Time{}) {
		conn.readDeadline = make(chan time.Time)
	} else if t.Before(time.Now()) {
		panic(fmt.Errorf("error setting read deadline: time passed is before time.Now()"))
	} else {
		conn.readDeadline = time.After(time.Until(t))
	}
	return nil
}

// SetWriteDeadline is a stub function to implement net.Conn. It has no functionality.
func (conn *Conn) SetWriteDeadline(time.Time) error {
	return nil
}

// Latency returns a rolling average of latency between the sending and the receiving end of the connection.
// The latency returned is updated continuously and is half the round trip time (RTT).
func (conn *Conn) Latency() time.Duration {
	if c, ok := conn.conn.(interface {
		Latency() time.Duration
	}); ok {
		return c.Latency()
	}
	panic(fmt.Sprintf("connection type %T has no Latency() time.Duration method", conn.conn))
}

// ClientCacheEnabled checks if the connection has the client blob cache enabled. If true, the server may send
// blobs to the client to reduce network transmission, but if false, the client does not support it, and the
// server must send chunks as usual.
func (conn *Conn) ClientCacheEnabled() bool {
	return conn.cacheEnabled
}

// ChunkRadius returns the initial chunk radius of the connection. For connections obtained through a
// Listener, this is the radius that the client requested. For connections obtained through a Dialer, this
// is the radius that the server approved upon.
func (conn *Conn) ChunkRadius() int {
	return int(conn.gameData.ChunkRadius)
}

// SetGameData manually sets the game data for this connection. This is useful when DisablePacketHandling
// is enabled and you want to populate the internal state without automatic packet handling.
// This allows GameData() to return meaningful data even when packet handlers aren't running.
func (conn *Conn) SetGameData(data GameData) {
	conn.gameData = data
	// When setting gameData with Items, also update shieldID if present
	for _, item := range data.Items {
		if item.Name == "minecraft:shield" {
			conn.shieldID.Store(int32(item.RuntimeID))
			break
		}
	}
}

// Context returns the connection's context. The context is canceled when the connection is closed,
// allowing for cancellation of operations that are tied to the lifecycle of the connection.
func (conn *Conn) Context() context.Context {
	return conn.ctx
}

// Disconnect disconnects the connection by first sending a disconnect packet with the message passed, and
// closing the connection after. If the message passed is empty, the client will be immediately sent to the
// server list instead of a disconnect screen.
func (conn *Conn) Disconnect(message string) error {
	return conn.DisconnectPacket(packet.Disconnect{
		HideDisconnectionScreen: message == "",
		Message:                 message,
	})
}

// DisconnectPacket disconnects the connection by first sending pk, and closing
// the connection after.
func (conn *Conn) DisconnectPacket(pk packet.Disconnect) error {
	_ = conn.WritePacketImmediate(&pk)
	return conn.close(conn.closeErr(conn.disconnectPacketMessage(&pk)))
}

func (conn *Conn) disconnectPacketMessage(pk *packet.Disconnect) string {
	if pk.Message != "" {
		return pk.Message
	}
	if reason, ok := disconnectReasons[pk.Reason]; ok {
		return reason
	}
	conn.log.Debug("unknown disconnect reason", "reason", pk.Reason)
	return fmt.Sprintf("Unknown disconnect reason: %d", pk.Reason)
}

// takeDeferredPacket locks the deferred packets lock and takes the next packet from the list of deferred
// packets. If none was found, it returns false, and if one was found, the data and true is returned.
func (conn *Conn) takeDeferredPacket() (*packetData, bool) {
	conn.deferredPacketMu.Lock()
	defer conn.deferredPacketMu.Unlock()

	if len(conn.deferredPackets) == 0 {
		return nil, false
	}
	data := conn.deferredPackets[0]
	// Explicitly clear out the packet at offset 0. When we slice it to remove the first element, that element
	// will not be garbage collectable, because the array it's in is still referenced by the slice. Doing this
	// makes sure garbage collecting the packet is possible.
	conn.deferredPackets[0] = nil
	conn.deferredPackets = conn.deferredPackets[1:]
	return data, true
}

// deferPacket defers a packet so that it is obtained in the next ReadPacket call
func (conn *Conn) deferPacket(pk *packetData) {
	conn.deferredPacketMu.Lock()
	conn.deferredPackets = append(conn.deferredPackets, pk)
	conn.deferredPacketMu.Unlock()
}

// receive receives an incoming serialised packet from the underlying connection. If the connection is not yet
// logged in, the packet is immediately handled.
func (conn *Conn) receive(data []byte) error {
	pkData, err := parseData(data, conn)
	if err != nil {
		return err
	}
	if pkData.h.PacketID == packet.IDDisconnect {
		// We always handle disconnect packets and close the connection if one comes in.
		pks, err := pkData.decode(conn)
		if err != nil {
			return err
		}
		disconnectPacket := pks[0].(*packet.Disconnect)
		disconnectMessage := conn.disconnectPacketMessage(disconnectPacket)
		_ = conn.close(conn.wrap(&DisconnectPacketError{
			Reason:                  disconnectPacket.Reason,
			HideDisconnectionScreen: disconnectPacket.HideDisconnectionScreen,
			Message:                 disconnectPacket.Message,
			FilteredMessage:         disconnectPacket.FilteredMessage,
			DisplayMessage:          disconnectMessage,
		}, "receive"))
		return nil
	}
	if conn.disablePacketHandling {
		if conn.handshakeComplete || conn.loggedIn {
			conn.disablePacketHandlingReady = true
		} else if !conn.disablePacketHandlingReady {
			switch pkData.h.PacketID {
			case packet.IDResourcePacksInfo, packet.IDStartGame, packet.IDPlayStatus:
				// Servers that skip the handshake packet should still switch to passthrough mode once post-login
				// packets start coming in.
				conn.disablePacketHandlingReady = true
			}
		}
		if conn.disablePacketHandlingReady {
			if pkData.h.PacketID == packet.IDClientToServerHandshake {
				return nil // don't forward it
			}
			select {
			case <-conn.ctx.Done():
			case conn.packets <- pkData:
			}
			return nil
		}
	}
	if conn.loggedIn && !conn.waitingForSpawn.Load() {
		select {
		case <-conn.ctx.Done():
		case previous := <-conn.packets:
			// There was already a packet in this channel, so take it out and defer it so that it is read
			// next.
			conn.deferPacket(previous)
		default:
		}
		select {
		case <-conn.ctx.Done():
		case conn.packets <- pkData:
		}
		return nil
	}
	return conn.handle(pkData)
}

// handle tries to handle the incoming packetData.
func (conn *Conn) handle(pkData *packetData) error {
	for _, id := range conn.expectedIDs.Load().([]uint32) {
		if id == pkData.h.PacketID {
			// If the packet was expected, so we handle it right now.
			pks, err := pkData.decode(conn)
			if err != nil {
				return err
			}
			return conn.handleMultiple(pks)
		}
	}
	// This is not the packet we expected next in the login sequence. We push it back so that it may
	// be handled by the user.
	conn.deferPacket(pkData)
	return nil
}

// handleMultiple handles multiple packets and returns an error if at least one of those packets could not be handled
// successfully.
func (conn *Conn) handleMultiple(pks []packet.Packet) error {
	var err error
	for _, pk := range pks {
		if e := conn.handlePacket(pk); e != nil {
			err = fmt.Errorf("handle %T: %w", pk, e)
		}
	}
	return err
}

// handlePacket handles an incoming packet. It returns an error if any of the data found in the packet was not
// valid or if handling failed for any other reason.
func (conn *Conn) handlePacket(pk packet.Packet) error {
	defer func() {
		_ = conn.Flush()
	}()
	switch pk := pk.(type) {
	// Internal packets destined for the server.
	case *packet.RequestNetworkSettings:
		return conn.handleRequestNetworkSettings(pk)
	case *packet.Login:
		return conn.handleLogin(pk)
	case *packet.ClientToServerHandshake:
		return conn.handleClientToServerHandshake()
	case *packet.ClientCacheStatus:
		return conn.handleClientCacheStatus(pk)
	case *packet.ResourcePackClientResponse:
		return conn.handleResourcePackClientResponse(pk)
	case *packet.ResourcePackChunkRequest:
		return conn.handleResourcePackChunkRequest(pk)
	case *packet.RequestChunkRadius:
		return conn.handleRequestChunkRadius(pk)
	case *packet.SetLocalPlayerAsInitialised:
		return conn.handleSetLocalPlayerAsInitialised(pk)

	// Internal packets destined for the client.
	case *packet.NetworkSettings:
		return conn.handleNetworkSettings(pk)
	case *packet.ServerToClientHandshake:
		return conn.handleServerToClientHandshake(pk)
	case *packet.PlayStatus:
		return conn.handlePlayStatus(pk)
	case *packet.ResourcePacksInfo:
		return conn.handleResourcePacksInfo(pk)
	case *packet.ResourcePackDataInfo:
		return conn.handleResourcePackDataInfo(pk)
	case *packet.ResourcePackChunkData:
		return conn.handleResourcePackChunkData(pk)
	case *packet.ResourcePackStack:
		return conn.handleResourcePackStack(pk)
	case *packet.StartGame:
		return conn.handleStartGame(pk)
	case *packet.ItemRegistry:
		return conn.handleItemRegistry(pk)
	case *packet.ChunkRadiusUpdated:
		return conn.handleChunkRadiusUpdated(pk)
	case *packet.DimensionData:
		return conn.handleDimensionData(pk)
	}
	return nil
}

// handleRequestNetworkSettings handles an incoming RequestNetworkSettings packet. It returns an error if the protocol
// version is not supported, otherwise sending back a NetworkSettings packet.
func (conn *Conn) handleRequestNetworkSettings(pk *packet.RequestNetworkSettings) error {
	found := false
	for _, pro := range conn.acceptedProto {
		if pro.ID() == pk.ClientProtocol {
			conn.proto = pro
			conn.pool = pro.Packets(true)
			found = true
			break
		}
	}

	// Allow newer clients to connect. Most protocol updates are still playable for the most part.
	// if pk.ClientProtocol > protocol.CurrentProtocol {
	// 	found = true
	// } else {
	// 	found = true // just allow all
	// }

	if !found {
		status := packet.PlayStatusLoginFailedClient
		// Dead code because of the newly added check above
		if pk.ClientProtocol > protocol.CurrentProtocol {
			// The server is outdated in this case, so we have to change the status we send.
			status = packet.PlayStatusLoginFailedServer
		}
		_ = conn.WritePacket(&packet.PlayStatus{Status: status})
		return fmt.Errorf("incompatible protocol version: expected %v, got %v", protocol.CurrentProtocol, pk.ClientProtocol)
	}

	if conn.compressionSelector != nil {
		if c := conn.compressionSelector(conn.proto); c != nil {
			conn.compression = c
		}
	}

	conn.expect(packet.IDLogin)
	if err := conn.WritePacket(&packet.NetworkSettings{
		CompressionThreshold: uint16(conn.compressionThreshold),
		CompressionAlgorithm: conn.compression.EncodeCompression(),
	}); err != nil {
		return fmt.Errorf("send NetworkSettings: %w", err)
	}
	_ = conn.Flush()
	conn.encMu.Lock()
	conn.enc.EnableCompression(conn.compression, conn.compressionThreshold)
	conn.encMu.Unlock()
	conn.dec.EnableCompression(conn.compression, conn.maxDecompressedLen)
	return nil
}

// handleNetworkSettings handles an incoming NetworkSettings packet, enabling compression for future packets.
func (conn *Conn) handleNetworkSettings(pk *packet.NetworkSettings) error {
	alg, ok := packet.CompressionByID(pk.CompressionAlgorithm)
	if !ok {
		conn.log.Warn("unknown compression algorithm", "algorithm", pk.CompressionAlgorithm)
	}
	conn.encMu.Lock()
	conn.enc.EnableCompression(alg, int(pk.CompressionThreshold))
	conn.encMu.Unlock()
	conn.dec.EnableCompression(alg, conn.maxDecompressedLen)
	conn.readyToLogin = true
	return nil
}

// handleLogin handles an incoming login packet. It verifies and decodes the login request found in the packet
// and returns an error if it couldn't be done successfully.
func (conn *Conn) handleLogin(pk *packet.Login) error {
	// The next expected packet is a response from the client to the handshake.
	conn.expect(packet.IDClientToServerHandshake)
	var (
		err        error
		authResult login.AuthResult
	)
	conn.identityData, conn.clientData, authResult, err = login.Parse(pk.ConnectionRequest, conn.verifier)
	if err != nil {
		return fmt.Errorf("parse login request: %w", err)
	}

	// Make sure the player is logged in with XBOX Live when necessary.
	if !authResult.XBOXLiveAuthenticated && conn.authEnabled {
		_ = conn.WritePacket(&packet.Disconnect{Message: text.Colourf("<red>You must be logged in with XBOX Live to join.</red>")})
		return fmt.Errorf("client was not authenticated to XBOX Live")
	}
	if err := conn.enableEncryption(authResult.PublicKey); err != nil {
		return fmt.Errorf("enable encryption: %w", err)
	}
	return nil
}

// handleClientToServerHandshake handles an incoming ClientToServerHandshake packet.
func (conn *Conn) handleClientToServerHandshake() error {
	conn.handshakeComplete = true
	if conn.disablePacketHandling {
		conn.disablePacketHandlingReady = true
		return nil
	}
	// The next expected packet is a resource pack client response.
	conn.expect(packet.IDResourcePackClientResponse, packet.IDClientCacheStatus)
	if err := conn.WritePacket(&packet.PlayStatus{Status: packet.PlayStatusLoginSuccess}); err != nil {
		return fmt.Errorf("send PlayStatus (Status=LoginSuccess): %w", err)
	}

	if conn.fetchResourcePacks != nil {
		conn.resourcePacks = conn.fetchResourcePacks(conn.identityData, conn.clientData, slices.Clone(conn.resourcePacks))
	}
	pk := &packet.ResourcePacksInfo{TexturePackRequired: conn.texturePacksRequired, ForceDisableVibrantVisuals: conn.forceDisableVibrantVisuals}
	for _, pack := range conn.resourcePacks {
		texturePack := protocol.TexturePackInfo{
			UUID:        pack.UUID(),
			Version:     pack.Version(),
			Size:        uint64(pack.Size()),
			DownloadURL: pack.DownloadURL(),
		}
		if pack.Encrypted() {
			texturePack.ContentKey = pack.ContentKey()
			texturePack.ContentIdentity = pack.Manifest().Header.UUID.String()
		}
		pk.TexturePacks = append(pk.TexturePacks, texturePack)
	}
	// Finally we send the packet after the play status.
	if err := conn.WritePacket(pk); err != nil {
		return fmt.Errorf("send ResourcePacksInfo: %w", err)
	}
	return nil
}

// saltClaims holds the claims for the salt sent by the server in the ServerToClientHandshake packet.
type saltClaims struct {
	Salt string `json:"salt"`
}

// handleServerToClientHandshake handles an incoming ServerToClientHandshake packet. It initialises encryption
// on the client side of the connection, using the hash and the public key from the server exposed in the
// packet.
func (conn *Conn) handleServerToClientHandshake(pk *packet.ServerToClientHandshake) error {
	tok, err := jwt.ParseSigned(string(pk.JWT), []jose.SignatureAlgorithm{jose.ES384})
	if err != nil {
		return fmt.Errorf("parse server token: %w", err)
	}
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	raw, _ := tok.Headers[0].ExtraHeaders["x5u"]
	kStr, _ := raw.(string)

	pub := new(ecdsa.PublicKey)
	if err := login.ParsePublicKey(kStr, pub); err != nil {
		return fmt.Errorf("parse server public key: %w", err)
	}

	var c saltClaims
	if err := tok.Claims(pub, &c); err != nil {
		return fmt.Errorf("verify claims: %w", err)
	}
	c.Salt = strings.TrimRight(c.Salt, "=")
	salt, err := base64.RawStdEncoding.DecodeString(c.Salt)
	if err != nil {
		return fmt.Errorf("decode ServerToClientHandshake salt: %w", err)
	}

	if !conn.disableEncryption {
		keyBytes, err := conn.encryptionKey(salt, pub)
		if err != nil {
			return fmt.Errorf("derive encryption key: %w", err)
		}

		// Finally we enable encryption for the enc and dec using the secret pubKey bytes we produced.
		conn.encMu.Lock()
		conn.enc.EnableEncryption(keyBytes)
		conn.encMu.Unlock()
		conn.dec.EnableEncryption(keyBytes)
	}

	// We write a ClientToServerHandshake packet (which has no payload) as a response.
	_ = conn.WritePacket(&packet.ClientToServerHandshake{})
	conn.handshakeComplete = true
	return nil
}

// handleClientCacheStatus handles a ClientCacheStatus packet sent by the client. It specifies if the client
// has support for the client blob cache.
func (conn *Conn) handleClientCacheStatus(pk *packet.ClientCacheStatus) error {
	conn.cacheEnabled = pk.Enabled
	return nil
}

// handleResourcePacksInfo handles a ResourcePacksInfo packet sent by the server. The client responds by
// sending the packs it needs downloaded.
func (conn *Conn) handleResourcePacksInfo(pk *packet.ResourcePacksInfo) error {
	// First create a new resource pack queue with the information in the packet so we can download them
	// properly later.
	totalPacks := len(pk.TexturePacks)
	conn.packQueue = &resourcePackQueue{
		packAmount:       totalPacks,
		downloadingPacks: make(map[string]downloadingPack),
		awaitingPacks:    make(map[string]*downloadingPack),
	}
	packsToDownload := make([]string, 0, totalPacks)

	for index, pack := range pk.TexturePacks {
		id := pack.UUID.String()
		if _, ok := conn.packQueue.downloadingPacks[id]; ok {
			conn.log.Warn("handle ResourcePacksInfo: duplicate texture pack", "UUID", pack.UUID)
			conn.packQueue.packAmount--
			continue
		}
		if conn.downloadResourcePack != nil && !conn.downloadResourcePack(uuid.MustParse(id), pack.Version, index, totalPacks) {
			conn.ignoredResourcePacks = append(conn.ignoredResourcePacks, exemptedResourcePack{
				uuid:    id,
				version: pack.Version,
			})
			conn.packQueue.packAmount--
			continue
		}

		// Try to use the Download URL if set
		if pack.DownloadURL != "" {
			newPack, err := resource.ReadURLContextLimit(conn.ctx, pack.DownloadURL, pack.Size)
			if err != nil {
				conn.log.Warn("handle ResourcePacksInfo: failed to download pack from URL", "UUID", pack.UUID, "download_url", pack.DownloadURL, "err", err)
			} else if newPack.UUID() != pack.UUID || newPack.Version() != pack.Version {
				conn.log.Warn("handle ResourcePacksInfo: downloaded pack from URL did not match advertised pack", "UUID", pack.UUID, "version", pack.Version, "downloaded_UUID", newPack.UUID(), "downloaded_version", newPack.Version(), "download_url", pack.DownloadURL)
			} else {
				conn.resourcePacks = append(conn.resourcePacks, newPack.WithContentKey(pack.ContentKey))
				conn.packQueue.packAmount--
				continue
			}
		}

		// This UUID_Version is a hack Mojang put in place.
		packsToDownload = append(packsToDownload, id+"_"+pack.Version)
		conn.packQueue.downloadingPacks[id] = downloadingPack{
			size:       pack.Size,
			buf:        bytes.NewBuffer(make([]byte, 0, pack.Size)),
			newFrag:    make(chan []byte),
			contentKey: pack.ContentKey,
		}
	}

	if len(packsToDownload) != 0 {
		conn.expect(packet.IDResourcePackDataInfo, packet.IDResourcePackChunkData, packet.IDStartGame)
		_ = conn.WritePacket(&packet.ResourcePackClientResponse{
			Response:        packet.PackResponseSendPacks,
			PacksToDownload: packsToDownload,
		})
		return nil
	}
	conn.expect(packet.IDResourcePackStack, packet.IDStartGame)

	_ = conn.WritePacket(&packet.ResourcePackClientResponse{Response: packet.PackResponseAllPacksDownloaded})
	return nil
}

// handleResourcePackStack handles a ResourcePackStack packet sent by the server. The stack defines the order
// that resource packs are applied in.
func (conn *Conn) handleResourcePackStack(pk *packet.ResourcePackStack) error {
	// We currently don't apply resource packs in any way, so instead we just check if all resource packs in
	// the stacks are also downloaded.
	for _, pack := range pk.TexturePacks {
		if !conn.hasPack(pack.UUID, pack.Version, false) {
			return fmt.Errorf("texture pack (UUID=%v, version=%v) not downloaded", pack.UUID, pack.Version)
		}
	}
	conn.expect(packet.IDDimensionData, packet.IDStartGame)
	_ = conn.WritePacket(&packet.ResourcePackClientResponse{Response: packet.PackResponseCompleted})
	return nil
}

// hasPack checks if the connection has a resource pack downloaded with the UUID and version passed, provided
// the pack either has or does not have behaviours in it.
func (conn *Conn) hasPack(uuid string, version string, hasBehaviours bool) bool {
	for _, exempted := range exemptedPacks {
		if exempted.uuid == uuid && exempted.version == version {
			// The server may send this resource pack on the stack without sending it in the info, as the client
			// always has it downloaded.
			return true
		}
	}
	conn.packMu.Lock()
	defer conn.packMu.Unlock()

	for _, ignored := range conn.ignoredResourcePacks {
		if ignored.uuid == uuid && ignored.version == version {
			return true
		}
	}
	for _, pack := range conn.resourcePacks {
		if pack.UUID().String() == uuid && pack.Version() == version && pack.HasBehaviours() == hasBehaviours {
			return true
		}
	}
	return false
}

const (
	// packChunkSize is the size of a single chunk of data from a resource pack: 128 KiB.
	packChunkSize = 1024 * 128
	// resourcePackChunkSendDelay spaces ResourcePackChunkData packets so slow clients are not flooded while
	// downloading packs. Clients after 1.26.30 may fail resource pack downloads when pack chunks are sent
	// too aggressively.
	resourcePackChunkSendDelay = 200 * time.Millisecond
)

// handleResourcePackClientResponse handles an incoming resource pack client response packet. The packet is
// handled differently depending on the response.
func (conn *Conn) handleResourcePackClientResponse(pk *packet.ResourcePackClientResponse) error {
	switch pk.Response {
	case packet.PackResponseRefused:
		// Even though this response is never sent, we handle it appropriately in case it is changed to work
		// correctly again.
		return conn.close(conn.closeErr("resource pack refused"))
	case packet.PackResponseSendPacks:
		packs := pk.PacksToDownload
		conn.packQueue = &resourcePackQueue{packs: conn.resourcePacks}
		if err := conn.packQueue.Request(packs); err != nil {
			return fmt.Errorf("lookup resource packs by UUID: %w", err)
		}
		// Proceed with the first resource pack download. We run all downloads in sequence rather than in
		// parallel, as it's less prone to packet loss.
		if err := conn.nextResourcePackDownload(); err != nil {
			return err
		}
	case packet.PackResponseAllPacksDownloaded:
		pk := &packet.ResourcePackStack{BaseGameVersion: protocol.CurrentVersion, Experiments: []protocol.ExperimentData{{Name: "cameras", Enabled: true}}}
		for _, pack := range conn.resourcePacks {
			resourcePack := protocol.StackResourcePack{UUID: pack.UUID().String(), Version: pack.Version()}
			pk.TexturePacks = append(pk.TexturePacks, resourcePack)
		}
		for _, exempted := range exemptedPacks {
			pk.TexturePacks = append(pk.TexturePacks, protocol.StackResourcePack{
				UUID:    exempted.uuid,
				Version: exempted.version,
			})
		}
		if err := conn.WritePacket(pk); err != nil {
			return fmt.Errorf("send ResourcePackStack: %w", err)
		}
	case packet.PackResponseCompleted:
		conn.loggedIn = true
	default:
		return fmt.Errorf("unknown ResourcePackClientResponse response type %v", pk.Response)
	}
	return nil
}

// startGame sends a StartGame packet using the game data of the connection.
func (conn *Conn) startGame() error {
	data := conn.gameData
	if len(data.Dimensions) > 0 {
		if err := conn.WritePacket(&packet.DimensionData{Definitions: data.Dimensions}); err != nil {
			return err
		}
	}
	if err := conn.WritePacket(&packet.JigsawStructureData{
		StructureData: map[string]any{
			"processors":     make([]map[string]any, 0),
			"template_pools": make([]map[string]any, 0),
			"jigsaws":        make([]map[string]any, 0),
			"structure_sets": make([]map[string]any, 0),
		},
	}); err != nil {
		return err
	}
	if err := conn.WritePacket(&packet.VoxelShapes{}); err != nil {
		return err
	}
	if err := conn.WritePacket(&packet.StartGame{
		Difficulty:                   data.Difficulty,
		EntityUniqueID:               data.EntityUniqueID,
		EntityRuntimeID:              data.EntityRuntimeID,
		PlayerGameMode:               data.PlayerGameMode,
		PlayerPosition:               data.PlayerPosition,
		Pitch:                        data.Pitch,
		Yaw:                          data.Yaw,
		WorldSeed:                    data.WorldSeed,
		Dimension:                    data.Dimension,
		WorldSpawn:                   data.WorldSpawn,
		EditorWorldType:              data.EditorWorldType,
		CreatedInEditor:              data.CreatedInEditor,
		ExportedFromEditor:           data.ExportedFromEditor,
		PersonaDisabled:              data.PersonaDisabled,
		CustomSkinsDisabled:          data.CustomSkinsDisabled,
		EmoteChatMuted:               data.EmoteChatMuted,
		GameRules:                    data.GameRules,
		Time:                         data.Time,
		Blocks:                       data.CustomBlocks,
		AchievementsDisabled:         true,
		Generator:                    1,
		EducationFeaturesEnabled:     true,
		MultiPlayerGame:              true,
		MultiPlayerCorrelationID:     uuid.Must(uuid.NewRandom()).String(),
		CommandsEnabled:              true,
		WorldName:                    data.WorldName,
		LANBroadcastEnabled:          true,
		PlayerMovementSettings:       data.PlayerMovementSettings,
		WorldGameMode:                data.WorldGameMode,
		Hardcore:                     data.Hardcore,
		XBLBroadcastMode:             data.XBLBroadcastMode,
		ServerAuthoritativeInventory: data.ServerAuthoritativeInventory,
		PlayerPermissions:            data.PlayerPermissions,
		Experiments:                  data.Experiments,
		ClientSideGeneration:         data.ClientSideGeneration,
		ChatRestrictionLevel:         data.ChatRestrictionLevel,
		DisablePlayerInteractions:    data.DisablePlayerInteractions,
		BaseGameVersion:              data.BaseGameVersion,
		GameVersion:                  protocol.CurrentVersion,
		UseBlockNetworkIDHashes:      data.UseBlockNetworkIDHashes,
		PropertyData:                 data.PropertyData,
	}); err != nil {
		return err
	}
	if err := conn.WritePacket(&packet.ItemRegistry{Items: data.Items}); err != nil {
		return err
	}
	if err := conn.Flush(); err != nil {
		return err
	}
	conn.expect(packet.IDRequestChunkRadius, packet.IDSetLocalPlayerAsInitialised)
	return nil
}

// nextResourcePackDownload moves to the next resource pack to download and sends a resource pack data info
// packet with information about it.
func (conn *Conn) nextResourcePackDownload() error {
	pk, ok := conn.packQueue.NextPack()
	if !ok {
		return fmt.Errorf("no resource packs to download")
	}
	if err := conn.WritePacket(pk); err != nil {
		return fmt.Errorf("send ResourcePackDataInfo: %w", err)
	}
	// Set the next expected packet to ResourcePackChunkRequest packets.
	conn.expect(packet.IDResourcePackChunkRequest)
	return nil
}

// handleResourcePackDataInfo handles a resource pack data info packet, which initiates the downloading of the
// pack by the client.
func (conn *Conn) handleResourcePackDataInfo(pk *packet.ResourcePackDataInfo) error {
	id := strings.Split(pk.UUID, "_")[0]

	pack, ok := conn.packQueue.downloadingPacks[id]
	if !ok {
		// We either already downloaded the pack or we got sent an invalid UUID, that did not match any pack
		// sent in the ResourcePacksInfo packet.
		return fmt.Errorf("handle ResourcePackDataInfo: unknown pack (UUID=%v)", id)
	}
	if pack.size != pk.Size {
		// Size mismatch: The ResourcePacksInfo packet had a size for the pack that did not match with the
		// size sent here.
		conn.log.Warn("handle ResourcePackDataInfo: pack had a different size in ResourcePacksInfo than in ResourcePackDataInfo", "UUID", id, "packs_info_size", pack.size, "data_info_size", pk.Size)
		pack.size = pk.Size
	}

	// Remove the resource pack from the downloading packs and add it to the awaiting packets.
	delete(conn.packQueue.downloadingPacks, id)
	conn.packQueue.awaitingPacks[id] = &pack

	pack.chunkSize = pk.DataChunkSize

	// The client calculates the chunk count by itself: You could in theory send a chunk count of 0 even
	// though there's data, and the client will still download normally.
	chunkCount := int32(pk.Size / uint64(pk.DataChunkSize))
	if pk.Size%uint64(pk.DataChunkSize) != 0 {
		chunkCount++
	}

	idCopy := pk.UUID
	go func() {
		for i := int32(0); i < chunkCount; i++ {
			_ = conn.WritePacket(&packet.ResourcePackChunkRequest{
				UUID:       idCopy,
				ChunkIndex: i,
			})
			select {
			case <-conn.ctx.Done():
				return
			case frag := <-pack.newFrag:
				// Write the fragment to the full buffer of the downloading resource pack.
				_, _ = pack.buf.Write(frag)
			}
		}
		conn.packMu.Lock()
		defer conn.packMu.Unlock()

		if pack.buf.Len() != int(pack.size) {
			conn.log.Error(fmt.Sprintf("download resource pack: incorrect resource pack size: expected %v, got %v", pack.size, pack.buf.Len()), "UUID", id)
			return
		}
		// First parse the resource pack from the total byte buffer we obtained.
		newPack, err := resource.Read(pack.buf)
		if err != nil {
			conn.log.Error("download resource pack: invalid full resource pack data: "+err.Error(), "UUID", id)
			return
		}
		conn.packQueue.packAmount--
		// Finally we add the resource to the resource packs slice.
		conn.resourcePacks = append(conn.resourcePacks, newPack.WithContentKey(pack.contentKey))
		if conn.packQueue.packAmount == 0 {
			conn.expect(packet.IDResourcePackStack)
			_ = conn.WritePacket(&packet.ResourcePackClientResponse{Response: packet.PackResponseAllPacksDownloaded})
		}
	}()
	return nil
}

// handleResourcePackChunkData handles a resource pack chunk data packet, which holds a fragment of a resource
// pack that is being downloaded.
func (conn *Conn) handleResourcePackChunkData(pk *packet.ResourcePackChunkData) error {
	pk.UUID = strings.Split(pk.UUID, "_")[0]
	pack, ok := conn.packQueue.awaitingPacks[pk.UUID]
	if !ok {
		// We haven't received a ResourcePackDataInfo packet from the server, so we can't use this data to
		// download a resource pack.
		return fmt.Errorf("chunk data for resource pack that was not being downloaded")
	}
	lastData := pack.buf.Len()+int(pack.chunkSize) >= int(pack.size)
	if !lastData && uint32(len(pk.Data)) != pack.chunkSize {
		// The chunk data didn't have the full size and wasn't the last data to be sent for the resource pack,
		// meaning we got too little data.
		return fmt.Errorf("expected chunk size %v, got %v", pack.chunkSize, len(pk.Data))
	}
	if pk.ChunkIndex != pack.expectedIndex {
		return fmt.Errorf("expected chunk index %v, got %v", pack.expectedIndex, pk.ChunkIndex)
	}
	pack.expectedIndex++
	pack.newFrag <- pk.Data
	return nil
}

// handleResourcePackChunkRequest handles a resource pack chunk request, which requests a part of the resource
// pack to be downloaded.
func (conn *Conn) handleResourcePackChunkRequest(pk *packet.ResourcePackChunkRequest) error {
	current := conn.packQueue.currentPack
	uuid, _, _ := strings.Cut(pk.UUID, "_")
	if current.UUID().String() != uuid {
		return fmt.Errorf("expected pack UUID %v, but got %v", current.UUID(), pk.UUID)
	}
	if conn.packQueue.currentOffset != uint64(pk.ChunkIndex)*packChunkSize {
		return fmt.Errorf("expected pack UUID %v, but got %v", conn.packQueue.currentOffset/packChunkSize, pk.ChunkIndex)
	}
	response := &packet.ResourcePackChunkData{
		UUID:       pk.UUID,
		ChunkIndex: uint32(pk.ChunkIndex),
		DataOffset: conn.packQueue.currentOffset,
		Data:       make([]byte, packChunkSize),
	}
	conn.packQueue.currentOffset += packChunkSize
	// We read the data directly into the response's data.
	if n, err := current.ReadAt(response.Data, int64(response.DataOffset)); err != nil {
		// If we hit an EOF, we don't need to return an error, as we've simply reached the end of the content
		// AKA the last chunk.
		if err != io.EOF {
			return fmt.Errorf("read resource pack chunk: %w", err)
		}
		response.Data = response.Data[:n]
	}
	if err := conn.WritePacket(response); err != nil {
		return fmt.Errorf("send ResourcePackChunkData: %w", err)
	}

	lastChunk := response.DataOffset+uint64(len(response.Data)) >= uint64(current.Size())
	if lastChunk {
		if !conn.packQueue.AllDownloaded() {
			_ = conn.nextResourcePackDownload()
		} else {
			conn.expect(packet.IDResourcePackClientResponse)
		}
	}
	if err := waitResourcePackChunkSendDelay(conn.ctx); err != nil {
		return err
	}

	return nil
}

// waitResourcePackChunkSendDelay waits before processing the next resource pack chunk request.
func waitResourcePackChunkSendDelay(ctx context.Context) error {
	timer := time.NewTimer(resourcePackChunkSendDelay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (conn *Conn) handleDimensionData(pk *packet.DimensionData) error {
	conn.gameData.Dimensions = pk.Definitions
	return nil
}

var hiveRegex = regexp.MustCompile(`.*\.hivebedrock\.network.*`)

// handleStartGame handles an incoming StartGame packet. It is the signal that the player has been added to a
// world, and it obtains most of its dedicated properties.
func (conn *Conn) handleStartGame(pk *packet.StartGame) error {
	if hiveRegex.MatchString(conn.clientData.ServerAddress) {
		pk.BaseGameVersion = "1.17.0" // temp fix for hive
	}

	// We store dimensions in the conn through handleDimensionData, so we need to
	// restore it after building GameData from the StartGame packet.
	dimensions := conn.gameData.Dimensions
	conn.gameData = GameDataFromStartGame(pk)
	conn.gameData.Dimensions = dimensions

	_ = conn.WritePacket(&packet.ServerBoundLoadingScreen{Type: packet.LoadingScreenTypeStart})
	_ = conn.WritePacket(&packet.RequestChunkRadius{ChunkRadius: 16, MaxChunkRadius: 16})
	conn.expect(packet.IDItemRegistry, packet.IDResourcePackStack)
	return nil
}

func GameDataFromStartGame(pk *packet.StartGame) GameData {
	return GameData{
		Difficulty:                   pk.Difficulty,
		WorldName:                    pk.WorldName,
		WorldSeed:                    pk.WorldSeed,
		EntityUniqueID:               pk.EntityUniqueID,
		EntityRuntimeID:              pk.EntityRuntimeID,
		PlayerGameMode:               pk.PlayerGameMode,
		BaseGameVersion:              pk.BaseGameVersion,
		PlayerPosition:               pk.PlayerPosition,
		Pitch:                        pk.Pitch,
		Yaw:                          pk.Yaw,
		Dimension:                    pk.Dimension,
		WorldSpawn:                   pk.WorldSpawn,
		EditorWorldType:              pk.EditorWorldType,
		CreatedInEditor:              pk.CreatedInEditor,
		ExportedFromEditor:           pk.ExportedFromEditor,
		PersonaDisabled:              pk.PersonaDisabled,
		CustomSkinsDisabled:          pk.CustomSkinsDisabled,
		EmoteChatMuted:               pk.EmoteChatMuted,
		GameRules:                    pk.GameRules,
		Time:                         pk.Time,
		ServerBlockStateChecksum:     pk.ServerBlockStateChecksum,
		CustomBlocks:                 pk.Blocks,
		PlayerMovementSettings:       pk.PlayerMovementSettings,
		WorldGameMode:                pk.WorldGameMode,
		Hardcore:                     pk.Hardcore,
		XBLBroadcastMode:             pk.XBLBroadcastMode,
		ServerAuthoritativeInventory: pk.ServerAuthoritativeInventory,
		PlayerPermissions:            pk.PlayerPermissions,
		ChatRestrictionLevel:         pk.ChatRestrictionLevel,
		DisablePlayerInteractions:    pk.DisablePlayerInteractions,
		ClientSideGeneration:         pk.ClientSideGeneration,
		Experiments:                  pk.Experiments,
		UseBlockNetworkIDHashes:      pk.UseBlockNetworkIDHashes,
		PropertyData:                 pk.PropertyData,
	}
}

// handleItemRegistry handles an incoming ItemRegistry packet. It contains the item definitions that the client
// should use, including the shield ID which is necessary for reading and writing items in the future.
func (conn *Conn) handleItemRegistry(pk *packet.ItemRegistry) error {
	conn.gameData.Items = pk.Items
	for _, item := range pk.Items {
		if item.Name == "minecraft:shield" {
			conn.shieldID.Store(int32(item.RuntimeID))
		}
	}

	// _ = conn.WritePacket(&packet.RequestChunkRadius{ChunkRadius: 16, MaxChunkRadius: 16})
	conn.expect(packet.IDChunkRadiusUpdated, packet.IDPlayStatus)
	return nil
}

// handleRequestChunkRadius handles an incoming RequestChunkRadius packet. It sets the initial chunk radius
// of the connection, and spawns the player.
func (conn *Conn) handleRequestChunkRadius(pk *packet.RequestChunkRadius) error {
	if pk.ChunkRadius < 1 {
		return fmt.Errorf("expected chunk radius of at least 1, got %v", pk.ChunkRadius)
	}
	conn.expect(packet.IDSetLocalPlayerAsInitialised)
	radius := pk.ChunkRadius
	if r := conn.gameData.ChunkRadius; r != 0 {
		radius = r
	}
	_ = conn.WritePacket(&packet.ChunkRadiusUpdated{ChunkRadius: radius})
	conn.gameData.ChunkRadius = pk.ChunkRadius
	_ = conn.WritePacket(&packet.PlayStatus{Status: packet.PlayStatusPlayerSpawn})
	_ = conn.WritePacket(&packet.CreativeContent{})
	return nil
}

// handleChunkRadiusUpdated handles an incoming ChunkRadiusUpdated packet, which updates the initial chunk
// radius of the connection.
func (conn *Conn) handleChunkRadiusUpdated(pk *packet.ChunkRadiusUpdated) error {
	if pk.ChunkRadius < 1 {
		return fmt.Errorf("expected chunk radius of at least 1, got %v", pk.ChunkRadius)
	}
	conn.expect(packet.IDPlayStatus)

	conn.gameData.ChunkRadius = pk.ChunkRadius
	conn.gameDataReceived.Store(true)

	conn.tryFinaliseClientConn()
	return nil
}

// handleSetLocalPlayerAsInitialised handles an incoming SetLocalPlayerAsInitialised packet. It is the final
// packet in the spawning sequence and it marks the point where a server sided connection is considered
// logged in.
func (conn *Conn) handleSetLocalPlayerAsInitialised(pk *packet.SetLocalPlayerAsInitialised) error {
	if pk.EntityRuntimeID != conn.gameData.EntityRuntimeID {
		return fmt.Errorf("entity runtime ID mismatch: expected %v (from StartGame), got %v", conn.gameData.EntityRuntimeID, pk.EntityRuntimeID)
	}
	if conn.waitingForSpawn.CompareAndSwap(true, false) {
		close(conn.spawn)
	}
	return nil
}

// handlePlayStatus handles an incoming PlayStatus packet. It reacts differently depending on the status
// found in the packet.
func (conn *Conn) handlePlayStatus(pk *packet.PlayStatus) error {
	switch pk.Status {
	case packet.PlayStatusLoginSuccess:
		if err := conn.WritePacket(&packet.ClientCacheStatus{Enabled: conn.cacheEnabled}); err != nil {
			return fmt.Errorf("send ClientCacheStatus: %w", err)
		}
		// The next packet we expect is the ResourcePacksInfo packet.
		conn.expect(packet.IDResourcePacksInfo)
		return conn.Flush()
	case packet.PlayStatusLoginFailedClient:
		_ = conn.close(conn.closeErr("client outdated"))
		return fmt.Errorf("client outdated")
	case packet.PlayStatusLoginFailedServer:
		_ = conn.close(conn.closeErr("server outdated"))
		return fmt.Errorf("server outdated")
	case packet.PlayStatusPlayerSpawn:
		// We've spawned and can send the last packet in the spawn sequence.
		conn.waitingForSpawn.Store(true)
		conn.tryFinaliseClientConn()
		return nil
	case packet.PlayStatusLoginFailedInvalidTenant:
		_ = conn.close(conn.closeErr("invalid edu edition game owner"))
		return fmt.Errorf("invalid edu edition game owner")
	case packet.PlayStatusLoginFailedVanillaEdu:
		_ = conn.close(conn.closeErr("cannot join an edu edition game on vanilla"))
		return fmt.Errorf("cannot join an edu edition game on vanilla")
	case packet.PlayStatusLoginFailedEduVanilla:
		_ = conn.close(conn.closeErr("cannot join a vanilla game on edu edition"))
		return fmt.Errorf("cannot join a vanilla game on edu edition")
	case packet.PlayStatusLoginFailedServerFull:
		_ = conn.close(conn.closeErr("server full"))
		return fmt.Errorf("server full")
	case packet.PlayStatusLoginFailedEditorVanilla:
		_ = conn.close(conn.closeErr("cannot join a vanilla game on editor"))
		return fmt.Errorf("cannot join a vanilla game on editor")
	case packet.PlayStatusLoginFailedVanillaEditor:
		_ = conn.close(conn.closeErr("cannot join an editor game on vanilla"))
		return fmt.Errorf("cannot join an editor game on vanilla")
	default:
		return fmt.Errorf("unknown play status %v", pk.Status)
	}
}

// tryFinaliseClientConn attempts to finalise the client connection by sending
// the SetLocalPlayerAsInitialised packet when if the ChunkRadiusUpdated and
// PlayStatus packets have been sent.
func (conn *Conn) tryFinaliseClientConn() {
	if conn.waitingForSpawn.Load() && conn.gameDataReceived.Load() {
		conn.waitingForSpawn.Store(false)
		conn.gameDataReceived.Store(false)

		close(conn.spawn)
		conn.loggedIn = true
		_ = conn.WritePacket(&packet.ServerBoundLoadingScreen{Type: packet.LoadingScreenTypeEnd})
		_ = conn.WritePacket(&packet.SetLocalPlayerAsInitialised{EntityRuntimeID: conn.gameData.EntityRuntimeID})
	}
}

// enableEncryption enables encryption on the server side over the connection. It sends an unencrypted
// handshake packet to the client and enables encryption after that.
func (conn *Conn) enableEncryption(clientPublicKey *ecdsa.PublicKey) error {
	signer, _ := jose.NewSigner(jose.SigningKey{Key: conn.privateKey, Algorithm: jose.ES384}, &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]any{"x5u": login.MarshalPublicKey(&conn.privateKey.PublicKey)},
	})
	// We produce an encoded JWT using the header and payload above, then we send the JWT in a ServerToClient-
	// Handshake packet so that the client can initialise encryption.
	serverJWT, err := jwt.Signed(signer).Claims(saltClaims{Salt: base64.RawStdEncoding.EncodeToString(conn.salt)}).Serialize()
	if err != nil {
		return fmt.Errorf("compact serialise server JWT: %w", err)
	}
	if err := conn.WritePacket(&packet.ServerToClientHandshake{JWT: []byte(serverJWT)}); err != nil {
		return fmt.Errorf("send ServerToClientHandshake: %w", err)
	}
	// Flush immediately as we'll enable encryption after this.
	_ = conn.Flush()

	if !conn.disableEncryption {
		keyBytes, err := conn.encryptionKey(conn.salt, clientPublicKey)
		if err != nil {
			return fmt.Errorf("derive encryption key: %w", err)
		}

		// Finally we enable encryption for the encoder and decoder using the secret key bytes we produced.
		conn.encMu.Lock()
		conn.enc.EnableEncryption(keyBytes)
		conn.encMu.Unlock()
		conn.dec.EnableEncryption(keyBytes)
	}

	return nil
}

// encryptionKey computes the encryption key for the connection using the salt
// and the remote connection's public key. It derives the shared secret through
// ECDH key exchange, then produces a 32-byte key by hashing the salt with the
// shared secret.
func (conn *Conn) encryptionKey(salt []byte, pub *ecdsa.PublicKey) ([32]byte, error) {
	privateKey, err := conn.privateKey.ECDH()
	if err != nil {
		return [32]byte{}, fmt.Errorf("convert private key to ECDH: %w", err)
	}
	publicKey, err := pub.ECDH()
	if err != nil {
		return [32]byte{}, fmt.Errorf("convert public key to ECDH: %w", err)
	}
	sharedSecret, err := privateKey.ECDH(publicKey)
	if err != nil {
		return [32]byte{}, fmt.Errorf("compute shared secret: %w", err)
	}
	return sha256.Sum256(append(salt, sharedSecret...)), nil
}

// expect sets the packet IDs that are next expected to arrive.
func (conn *Conn) expect(packetIDs ...uint32) {
	conn.expectedIDs.Store(packetIDs)
}

func (conn *Conn) close(cause error) error {
	var err error
	conn.once.Do(func() {
		err = conn.Flush()
		conn.cancelFunc(cause)
		_ = conn.conn.Close()
	})
	return err
}

// closeErr returns an adequate connection closed error for the op passed. If the connection was closed
// through a Disconnect packet, the message is contained.
func (conn *Conn) closeErr(op string) error {
	select {
	case <-conn.ctx.Done():
		return conn.wrap(context.Cause(conn.ctx), op)
	default:
		return conn.wrap(net.ErrClosed, op)
	}
}
