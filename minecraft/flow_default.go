package minecraft

import (
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

var DefaultLoginFlow = LoginFlowHandler{
	HandleClientCacheStatus:           handleClientCacheStatus,
	HandleResourcePackClientResponse:  handleResourcePackClientResponse,
	HandleResourcePackChunkRequest:    handleResourcePackChunkRequest,
	HandleRequestChunkRadius:          handleRequestChunkRadius,
	HandleSetLocalPlayerAsInitialised: handleSetLocalPlayerAsInitialised,
	HandleResourcePacksInfo:           handleResourcePacksInfo,
	HandleResourcePackDataInfo:        handleResourcePackDataInfo,
	HandleResourcePackChunkData:       handleResourcePackChunkData,
	HandleResourcePackStack:           handleResourcePackStack,
	HandleDimensionData:               handleDimensionData,
	HandleStartGame:                   handleStartGame,
	HandleItemRegistry:                handleItemRegistry,
	HandleChunkRadiusUpdated:          handleChunkRadiusUpdated,
}

// handleRequestNetworkSettings handles an incoming RequestNetworkSettings packet. It returns an error if the protocol
// version is not supported, otherwise sending back a NetworkSettings packet.
func handleRequestNetworkSettings(conn *Conn, pk *packet.RequestNetworkSettings) error {
	found := false
	for _, pro := range conn.AcceptedProtocols() {
		if pro.ID() == pk.ClientProtocol {
			conn.SetProto(pro)
			conn.SetPacketPool(pro.Packets(true))
			found = true
			break
		}
	}
	if !found {
		status := packet.PlayStatusLoginFailedClient
		if pk.ClientProtocol > protocol.CurrentProtocol {
			// The server is outdated in this case, so we have to change the status we send.
			status = packet.PlayStatusLoginFailedServer
		}
		_ = conn.WritePacket(&packet.PlayStatus{Status: status})
		return fmt.Errorf("incompatible protocol version: expected %v, got %v", protocol.CurrentProtocol, pk.ClientProtocol)
	}

	comp := conn.Compression(conn.Proto())

	if err := conn.WritePacket(&packet.NetworkSettings{
		CompressionThreshold: uint16(conn.CompressionThreshold()),
		CompressionAlgorithm: comp.EncodeCompression(),
	}); err != nil {
		return fmt.Errorf("send NetworkSettings: %w", err)
	}
	_ = conn.Flush()
	conn.EnableCompression(comp, conn.CompressionThreshold(), conn.MaxDecompressedLen())
	return nil
}

// handleNetworkSettings handles an incoming NetworkSettings packet, enabling compression for future packets.
func handleNetworkSettings(conn *Conn, pk *packet.NetworkSettings) error {
	alg, ok := packet.CompressionByID(pk.CompressionAlgorithm)
	if !ok {
		return fmt.Errorf("unknown compression algorithm %v", pk.CompressionAlgorithm)
	}
	conn.EnableCompression(alg, int(pk.CompressionThreshold), conn.MaxDecompressedLen())
	if err := conn.WritePacket(&packet.Login{ConnectionRequest: conn.loginRequest, ClientProtocol: conn.Proto().ID()}); err != nil {
		return fmt.Errorf("send Login: %w", err)
	}
	return nil
}

// handleLogin handles an incoming login packet. It verifies and decodes the login request found in the packet
// and returns an error if it couldn't be done successfully.
func handleLogin(conn *Conn, pk *packet.Login) error {
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
	if pkc, ok := conn.conn.(publicKeyConn); ok {
		if pub := pkc.PublicKey(); pub != nil && !authResult.PublicKey.Equal(pub) {
			_ = conn.WritePacket(&packet.Disconnect{Reason: packet.DisconnectReasonNotAuthenticated})
			return fmt.Errorf("identity public key mismatch: %s != %s", login.MarshalPublicKey(authResult.PublicKey), login.MarshalPublicKey(pub))
		}
	}
	if conn.allow != nil {
		if reason, ok := conn.allow(conn.RemoteAddr(), conn.identityData, conn.clientData); !ok {
			_ = conn.WritePacket(&packet.Disconnect{Reason: packet.DisconnectReasonKicked, Message: reason})
			return conn.Close()
		}
	}
	if conn.disableEncryption {
		// Encryption is disabled (e.g. it is handled by the transport), so we skip the handshake and continue
		// directly with the next step of the login flow.
		if err := sendLoginSuccessAndPacks(conn); err != nil {
			return err
		}
		return nil
	}
	if err := conn.EnableEncryption(authResult.PublicKey); err != nil {
		return fmt.Errorf("enable encryption: %w", err)
	}
	return nil
}

// handleClientToServerHandshake handles an incoming ClientToServerHandshake packet.
func handleClientToServerHandshake(conn *Conn, _ *packet.ClientToServerHandshake) error {
	if err := sendLoginSuccessAndPacks(conn); err != nil {
		return err
	}
	return nil
}

func sendLoginSuccessAndPacks(conn *Conn) error {
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
			Size:        uint64(pack.Len()),
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
func handleServerToClientHandshake(conn *Conn, pk *packet.ServerToClientHandshake) error {
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

	keyBytes, err := conn.encryptionKey(salt, pub)
	if err != nil {
		return fmt.Errorf("derive encryption key: %w", err)
	}

	conn.encMu.Lock()
	conn.enc.EnableEncryption(keyBytes)
	conn.encMu.Unlock()
	conn.dec.EnableEncryption(keyBytes)

	_ = conn.WritePacket(&packet.ClientToServerHandshake{})
	return nil
}

// handleClientCacheStatus handles a ClientCacheStatus packet sent by the client. It specifies if the client
// has support for the client blob cache.
func handleClientCacheStatus(conn *Conn, pk *packet.ClientCacheStatus) error {
	conn.SetCacheEnabled(pk.Enabled)
	return nil
}

// handleResourcePacksInfo handles a ResourcePacksInfo packet sent by the server. The client responds by
// sending the packs it needs downloaded.
func handleResourcePacksInfo(conn *Conn, pk *packet.ResourcePacksInfo) error {
	return conn.HandleResourcePacksInfo(pk)
}

// handleResourcePackStack handles a ResourcePackStack packet sent by the server. The stack defines the order
// that resource packs are applied in.
func handleResourcePackStack(conn *Conn, pk *packet.ResourcePackStack) error {
	return conn.HandleResourcePackStack(pk)
}

// handleResourcePackClientResponse handles an incoming resource pack client response packet. The packet is
// handled differently depending on the response.
func handleResourcePackClientResponse(conn *Conn, pk *packet.ResourcePackClientResponse) error {
	return conn.HandleResourcePackClientResponse(pk)
}

// handleResourcePackDataInfo handles a resource pack data info packet, which initiates the downloading of the
// pack by the client.
func handleResourcePackDataInfo(conn *Conn, pk *packet.ResourcePackDataInfo) error {
	return conn.BeginPackDataInfo(pk)
}

// handleResourcePackChunkData handles a resource pack chunk data packet, which holds a fragment of a resource
// pack that is being downloaded.
func handleResourcePackChunkData(conn *Conn, pk *packet.ResourcePackChunkData) error {
	return conn.HandlePackChunkData(pk)
}

// handleResourcePackChunkRequest handles a resource pack chunk request, which requests a part of the resource
// pack to be downloaded.
func handleResourcePackChunkRequest(conn *Conn, pk *packet.ResourcePackChunkRequest) error {
	return conn.HandlePackChunkRequest(pk)
}

func handleDimensionData(conn *Conn, pk *packet.DimensionData) error {
	gameData := conn.GameData()
	gameData.Dimensions = pk.Definitions
	conn.SetGameData(gameData)
	return nil
}

// handleStartGame handles an incoming StartGame packet. It is the signal that the player has been added to a
// world, and it obtains most of its dedicated properties.
func handleStartGame(conn *Conn, pk *packet.StartGame) error {
	conn.SetGameData(GameData{
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
		DayCycleLockTime:             pk.DayCycleLockTime,
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
		Dimensions:                   conn.GameData().Dimensions,
	})
	return nil
}

func handleItemRegistry(conn *Conn, pk *packet.ItemRegistry) error {
	gameData := conn.GameData()
	gameData.Items = pk.Items
	conn.SetGameData(gameData)
	for _, item := range pk.Items {
		if item.Name == "minecraft:shield" {
			conn.SetShieldID(int32(item.RuntimeID))
		}
	}

	_ = conn.WritePacket(&packet.RequestChunkRadius{ChunkRadius: 16, MaxChunkRadius: 16})
	return nil
}

// handleRequestChunkRadius handles an incoming RequestChunkRadius packet. It sets the initial chunk radius of
// the connection, and spawns the player.
func handleRequestChunkRadius(conn *Conn, pk *packet.RequestChunkRadius) error {
	if pk.ChunkRadius < 1 {
		return fmt.Errorf("expected chunk radius of at least 1, got %v", pk.ChunkRadius)
	}
	radius := pk.ChunkRadius
	if r := conn.GameData().ChunkRadius; r != 0 {
		radius = r
	}
	_ = conn.WritePacket(&packet.ChunkRadiusUpdated{ChunkRadius: radius})
	gameData := conn.GameData()
	gameData.ChunkRadius = pk.ChunkRadius
	conn.SetGameData(gameData)
	_ = conn.WritePacket(&packet.PlayStatus{Status: packet.PlayStatusPlayerSpawn})
	_ = conn.WritePacket(&packet.CreativeContent{})
	return nil
}

// handleChunkRadiusUpdated handles an incoming ChunkRadiusUpdated packet, which updates the initial chunk
// radius of the connection.
func handleChunkRadiusUpdated(conn *Conn, pk *packet.ChunkRadiusUpdated) error {
	if pk.ChunkRadius < 1 {
		return fmt.Errorf("expected chunk radius of at least 1, got %v", pk.ChunkRadius)
	}
	gameData := conn.GameData()
	gameData.ChunkRadius = pk.ChunkRadius
	conn.SetGameData(gameData)
	conn.GameDataReceived()
	return nil
}

func handleSetLocalPlayerAsInitialised(conn *Conn, pk *packet.SetLocalPlayerAsInitialised) error {
	if pk.EntityRuntimeID != conn.GameData().EntityRuntimeID {
		return fmt.Errorf("entity runtime ID mismatch: expected %v (from StartGame), got %v", conn.GameData().EntityRuntimeID, pk.EntityRuntimeID)
	}
	conn.Spawned()
	return nil
}

func handlePlayStatus(conn *Conn, pk *packet.PlayStatus) error {
	switch pk.Status {
	case packet.PlayStatusLoginSuccess:
		if err := conn.WritePacket(&packet.ClientCacheStatus{Enabled: conn.ClientCacheEnabled()}); err != nil {
			return fmt.Errorf("send ClientCacheStatus: %w", err)
		}
	case packet.PlayStatusLoginFailedClient:
		_ = conn.close(conn.closeErr("client outdated"))
		return fmt.Errorf("client outdated")
	case packet.PlayStatusLoginFailedServer:
		_ = conn.close(conn.closeErr("server outdated"))
		return fmt.Errorf("server outdated")
	case packet.PlayStatusPlayerSpawn:
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
	switch pk.Status {
	case packet.PlayStatusLoginSuccess:
		// The next packet we expect is the ResourcePacksInfo packet.
		conn.expect(packet.IDResourcePacksInfo)
		return conn.Flush()
	case packet.PlayStatusPlayerSpawn:
		// We've spawned and can send the last packet in the spawn sequence.
		conn.waitingForSpawn.Store(true)
		conn.tryFinaliseClientConn()
	}
	return nil
}
