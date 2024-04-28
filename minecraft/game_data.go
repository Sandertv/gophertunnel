package minecraft

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// GameData is a loose wrapper around a part of the data found in the StartGame packet. It holds data sent
// specifically at the start of the game, such as the position of the player, the game mode, etc.
type GameData struct {
	// WorldName is the name of the world that the player spawns in. This name will be displayed at the top of
	// the player list when opening the in-game menu. It may contain colour codes and does not have to be an
	// actual world name, but instead, can be the server name.
	// If WorldName is left empty, the name of the Listener will be used to show above the player list
	// in-game.
	WorldName string
	// WorldSeed is the seed used to generate the world. Unlike in PC edition, the seed is a 32bit integer
	// here.
	WorldSeed int64
	// Difficulty is the difficulty of the world that the player spawns in. A difficulty of 0, peaceful, means
	// the player will automatically regenerate health and hunger.
	Difficulty int32
	// EntityUniqueID is the unique ID of the player. The unique ID is unique for the entire world and is
	// often used in packets. Most servers send an EntityUniqueID equal to the EntityRuntimeID.
	EntityUniqueID int64
	// EntityRuntimeID is the runtime ID of the player. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// PlayerGameMode is the game mode the player currently has. It is a value from 0-4, with 0 being
	// survival mode, 1 being creative mode, 2 being adventure mode, 3 being survival spectator and 4 being
	// creative spectator.
	// This field may be set to 5 to make the client fall back to the game mode set in the WorldGameMode
	// field.
	PlayerGameMode int32
	// PersonaDisabled is true if persona skins are disabled for the current game session.
	PersonaDisabled bool
	// CustomSkinsDisabled is true if custom skins are disabled for the current game session.
	CustomSkinsDisabled bool
	// BaseGameVersion is the version of the game from which Vanilla features will be used. The exact function
	// of this field isn't clear.
	BaseGameVersion string
	// PlayerPosition is the spawn position of the player in the world. In servers this is often the same as
	// the world's spawn position found below.
	PlayerPosition mgl32.Vec3
	// Pitch is the vertical rotation of the player. Facing straight forward yields a pitch of 0. Pitch is
	// measured in degrees.
	Pitch float32
	// Yaw is the horizontal rotation of the player. Yaw is also measured in degrees.
	Yaw float32
	// Dimension is the ID of the dimension that the player spawns in. It is a value from 0-2, with 0 being
	// the overworld, 1 being the nether and 2 being the end.
	Dimension int32
	// WorldSpawn is the block on which the world spawn of the world. This coordinate has no effect on the
	// place that the client spawns, but it does have an effect on the direction that a compass points.
	WorldSpawn protocol.BlockPos
	// EditorWorldType is a value to dictate the type of editor mode, a special mode recently introduced adding
	// "powerful tools for editing worlds, intended for experienced creators."
	EditorWorldType int32
	// CreatedInEditor is a value to dictate if the world was created as a project in the editor mode. The functionality
	// of this field is currently unknown.
	CreatedInEditor bool
	// ExportedFromEditor is a value to dictate if the world was exported from editor mode. The functionality of this
	// field is currently unknown.
	ExportedFromEditor bool
	// WorldGameMode is the game mode that a player gets when it first spawns in the world. It is shown in the
	// settings and is used if the PlayerGameMode is set to 5.
	WorldGameMode int32
	// Hardcore is if the world is in hardcore mode. In hardcore mode, the player cannot respawn after dying.
	Hardcore bool
	// GameRules defines game rules currently active with their respective values. The value of these game
	// rules may be either 'bool', 'int32' or 'float32'. Some game rules are server side only, and don't
	// necessarily need to be sent to the client.
	GameRules []protocol.GameRule
	// Time is the total time that has elapsed since the start of the world.
	Time int64
	// ServerBlockStateChecksum is a checksum to ensure block states between the server and client match.
	// This can simply be left empty, and the client will avoid trying to verify it.
	ServerBlockStateChecksum uint64
	// CustomBlocks is a list of custom blocks added to the game by the server. These blocks all have a name
	// and block properties.
	CustomBlocks []protocol.BlockEntry
	// Items is a list of all items existing in the game, including custom items registered by the server.
	Items []protocol.ItemEntry
	// PlayerMovementSettings specify the different server authoritative movement settings that it has
	// enabled.
	PlayerMovementSettings protocol.PlayerMovementSettings
	// ServerAuthoritativeInventory specifies if the server authoritative inventory system is enabled. This
	// is a new system introduced in 1.16. Backwards compatibility with the inventory transactions has to
	// some extent been preserved, but will eventually be removed.
	ServerAuthoritativeInventory bool
	// Experiments is a list of experiments enabled on the server side. These experiments are used to enable
	// disable experimental features.
	Experiments []protocol.ExperimentData
	// PlayerPermissions is the permission level of the player. It is a value from 0-3, with 0 being visitor,
	// 1 being member, 2 being operator and 3 being custom.
	PlayerPermissions int32
	// ChunkRadius is the initial chunk radius that the connection gets. This can be changed later on using a
	// packet.ChunkRadiusUpdated.
	ChunkRadius int32
	// ClientSideGeneration is true if the client should use the features registered in the FeatureRegistry packet to
	// generate terrain client-side to save on bandwidth.
	ClientSideGeneration bool
	// ChatRestrictionLevel specifies the level of restriction on in-game chat. It is one of the constants above.
	ChatRestrictionLevel uint8
	// DisablePlayerInteractions is true if the client should ignore other players when interacting with the world.
	DisablePlayerInteractions bool
	// UseBlockNetworkIDHashes is true if the client should use the hash of a block's name as its network ID rather than
	// its index in the expected block palette. This is useful for servers that wish to support multiple protocol versions
	// and custom blocks, but it will result in extra bytes being written for every block in a sub chunk palette.
	UseBlockNetworkIDHashes bool
}
