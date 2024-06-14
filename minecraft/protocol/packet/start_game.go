package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	SpawnBiomeTypeDefault = iota
	SpawnBiomeTypeUserDefined
)

const (
	ChatRestrictionLevelNone     = 0
	ChatRestrictionLevelDropped  = 1
	ChatRestrictionLevelDisabled = 2
)

const (
	EditorWorldTypeNotEditor = iota
	EditorWorldTypeProject
	EditorWorldTypeTestLevel
)

// StartGame is sent by the server to send information about the world the player will be spawned in. It
// contains information about the position the player spawns in, and information about the world in general
// such as its game rules.
type StartGame struct {
	// EntityUniqueID is the unique ID of the player. The unique ID is a value that remains consistent across
	// different sessions of the same world, but most servers simply fill the runtime ID of the entity out for
	// this field.
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
	// PlayerPosition is the spawn position of the player in the world. In servers this is often the same as
	// the world's spawn position found below.
	PlayerPosition mgl32.Vec3
	// Pitch is the vertical rotation of the player. Facing straight forward yields a pitch of 0. Pitch is
	// measured in degrees.
	Pitch float32
	// Yaw is the horizontal rotation of the player. Yaw is also measured in degrees.
	Yaw float32
	// WorldSeed is the seed used to generate the world. Unlike in PC edition, the seed is a 32bit integer
	// here.
	WorldSeed int64
	// SpawnBiomeType specifies if the biome that the player spawns in is user defined (through behaviour
	// packs) or builtin. See the constants above.
	SpawnBiomeType int16
	// UserDefinedBiomeName is a readable name of the biome that the player spawned in, such as 'plains'. This
	// might be a custom biome name if any custom biomes are present through behaviour packs.
	UserDefinedBiomeName string
	// Dimension is the ID of the dimension that the player spawns in. It is a value from 0-2, with 0 being
	// the overworld, 1 being the nether and 2 being the end.
	Dimension int32
	// Generator is the generator used for the world. It is a value from 0-4, with 0 being old limited worlds,
	// 1 being infinite worlds, 2 being flat worlds, 3 being nether worlds and 4 being end worlds. A value of
	// 0 will actually make the client stop rendering chunks you send beyond the world limit.
	Generator int32
	// WorldGameMode is the game mode that a player gets when it first spawns in the world. It is shown in the
	// settings and is used if the PlayerGameMode is set to 5.
	WorldGameMode int32
	// Hardcore is if the world is in hardcore mode. In hardcore mode, the player cannot respawn after dying.
	Hardcore bool
	// Difficulty is the difficulty of the world. It is a value from 0-3, with 0 being peaceful, 1 being easy,
	// 2 being normal and 3 being hard.
	Difficulty int32
	// WorldSpawn is the block on which the world spawn of the world. This coordinate has no effect on the
	// place that the client spawns, but it does have an effect on the direction that a compass points.
	WorldSpawn protocol.BlockPos
	// AchievementsDisabled defines if achievements are disabled in the world. The client crashes if this
	// value is set to true while the player's or the world's game mode is creative, and it's recommended to
	// simply always set this to false as a server.
	AchievementsDisabled bool
	// EditorWorldType is a value to dictate the type of editor mode, a special mode recently introduced adding
	// "powerful tools for editing worlds, intended for experienced creators." It is one of the constants above.
	EditorWorldType int32
	// CreatedInEditor is a value to dictate if the world was created as a project in the editor mode. The functionality
	// of this field is currently unknown.
	CreatedInEditor bool
	// ExportedFromEditor is a value to dictate if the world was exported from editor mode. The functionality of this
	// field is currently unknown.
	ExportedFromEditor bool
	// DayCycleLockTime is the time at which the day cycle was locked if the day cycle is disabled using the
	// respective game rule. The client will maintain this time as long as the day cycle is disabled.
	DayCycleLockTime int32
	// EducationEditionOffer is some Minecraft: Education Edition field that specifies what 'region' the world
	// was from, with 0 being None, 1 being RestOfWorld, and 2 being China.
	// The actual use of this field is unknown.
	EducationEditionOffer int32
	// EducationFeaturesEnabled specifies if the world has education edition features enabled, such as the
	// blocks or entities specific to education edition.
	EducationFeaturesEnabled bool
	// EducationProductID is a UUID used to identify the education edition server instance. It is generally
	// unique for education edition servers.
	EducationProductID string
	// RainLevel is the level specifying the intensity of the rain falling. When set to 0, no rain falls at
	// all.
	RainLevel float32
	// LightningLevel is the level specifying the intensity of the thunder. This may actually be set
	// independently from the RainLevel, meaning dark clouds can be produced without rain.
	LightningLevel float32
	// ConfirmedPlatformLockedContent ...
	ConfirmedPlatformLockedContent bool
	// MultiPlayerGame specifies if the world is a multi-player game. This should always be set to true for
	// servers.
	MultiPlayerGame bool
	// LANBroadcastEnabled specifies if LAN broadcast was intended to be enabled for the world.
	LANBroadcastEnabled bool
	// XBLBroadcastMode is the mode used to broadcast the joined game across XBOX Live.
	XBLBroadcastMode int32
	// PlatformBroadcastMode is the mode used to broadcast the joined game across the platform.
	PlatformBroadcastMode int32
	// CommandsEnabled specifies if commands are enabled for the player. It is recommended to always set this
	// to true on the server, as setting it to false means the player cannot, under any circumstance, use a
	// command.
	CommandsEnabled bool
	// TexturePackRequired specifies if the texture pack the world might hold is required, meaning the client
	// was forced to download it before joining.
	TexturePackRequired bool
	// GameRules defines game rules currently active with their respective values. The value of these game
	// rules may be either 'bool', 'int32' or 'float32'. Some game rules are server side only, and don't
	// necessarily need to be sent to the client.
	GameRules []protocol.GameRule
	// Experiments holds a list of experiments that are either enabled or disabled in the world that the
	// player spawns in.
	Experiments []protocol.ExperimentData
	// ExperimentsPreviouslyToggled specifies if any experiments were previously toggled in this world. It is
	// probably used for some kind of metrics.
	ExperimentsPreviouslyToggled bool
	// BonusChestEnabled specifies if the world had the bonus map setting enabled when generating it. It does
	// not have any effect client-side.
	BonusChestEnabled bool
	// StartWithMapEnabled specifies if the world has the start with map setting enabled, meaning each joining
	// player obtains a map. This should always be set to false, because the client obtains a map all on its
	// own accord if this is set to true.
	StartWithMapEnabled bool
	// PlayerPermissions is the permission level of the player. It is a value from 0-3, with 0 being visitor,
	// 1 being member, 2 being operator and 3 being custom.
	PlayerPermissions int32
	// ServerChunkTickRadius is the radius around the player in which chunks are ticked. Most servers set this
	// value to a fixed number, as it does not necessarily affect anything client-side.
	ServerChunkTickRadius int32
	// HasLockedBehaviourPack specifies if the behaviour pack of the world is locked, meaning it cannot be
	// disabled from the world. This is typically set for worlds on the marketplace that have a dedicated
	// behaviour pack.
	HasLockedBehaviourPack bool
	// HasLockedTexturePack specifies if the texture pack of the world is locked, meaning it cannot be
	// disabled from the world. This is typically set for worlds on the marketplace that have a dedicated
	// texture pack.
	HasLockedTexturePack bool
	// FromLockedWorldTemplate specifies if the world from the server was from a locked world template. For
	// servers this should always be set to false.
	FromLockedWorldTemplate bool
	// MSAGamerTagsOnly ..
	MSAGamerTagsOnly bool
	// FromWorldTemplate specifies if the world from the server was from a world template. For servers this
	// should always be set to false.
	FromWorldTemplate bool
	// WorldTemplateSettingsLocked specifies if the world was a template that locks all settings that change
	// properties above in the settings GUI. It is recommended to set this to true for servers that do not
	// allow things such as setting game rules through the GUI.
	WorldTemplateSettingsLocked bool
	// OnlySpawnV1Villagers is a hack that Mojang put in place to preserve backwards compatibility with old
	// villagers. The bool is never actually read though, so it has no functionality.
	OnlySpawnV1Villagers bool
	// PersonaDisabled is true if persona skins are disabled for the current game session.
	PersonaDisabled bool
	// CustomSkinsDisabled is true if custom skins are disabled for the current game session.
	CustomSkinsDisabled bool
	// EmoteChatMuted specifies if players will be sent a chat message when using certain emotes.
	EmoteChatMuted bool
	// BaseGameVersion is the version of the game from which Vanilla features will be used. The exact function
	// of this field isn't clear.
	BaseGameVersion string
	// LimitedWorldWidth and LimitedWorldDepth are the dimensions of the world if the world is a limited
	// world. For unlimited worlds, these may simply be left as 0.
	LimitedWorldWidth, LimitedWorldDepth int32
	// NewNether specifies if the server runs with the new nether introduced in the 1.16 update.
	NewNether bool
	// EducationSharedResourceURI is an education edition feature that transmits education resource settings to clients.
	EducationSharedResourceURI protocol.EducationSharedResourceURI
	// ForceExperimentalGameplay specifies if experimental gameplay should be force enabled. For servers this
	// should always be set to false.
	ForceExperimentalGameplay protocol.Optional[bool]
	// LevelID is a base64 encoded world ID that is used to identify the world.
	LevelID string
	// WorldName is the name of the world that the player is joining. Note that this field shows up above the
	// player list for the rest of the game session, and cannot be changed. Setting the server name to this
	// field is recommended.
	WorldName string
	// TemplateContentIdentity is a UUID specific to the premium world template that might have been used to
	// generate the world. Servers should always fill out an empty string for this.
	TemplateContentIdentity string
	// Trial specifies if the world was a trial world, meaning features are limited and there is a time limit
	// on the world.
	Trial bool
	// PlayerMovementSettings ...
	PlayerMovementSettings protocol.PlayerMovementSettings
	// Time is the total time that has elapsed since the start of the world.
	Time int64
	// EnchantmentSeed is the seed used to seed the random used to produce enchantments in the enchantment
	// table. Note that the exact correct random implementation must be used to produce the correct results
	// both client- and server-side.
	EnchantmentSeed int32
	// Blocks is a list of all custom blocks registered on the server.
	Blocks []protocol.BlockEntry
	// Items is a list of all items with their legacy IDs which are available in the game. Failing to send any
	// of the items that are in the game will crash mobile clients.
	Items []protocol.ItemEntry
	// MultiPlayerCorrelationID is a unique ID specifying the multi-player session of the player. A random
	// UUID should be filled out for this field.
	MultiPlayerCorrelationID string
	// ServerAuthoritativeInventory specifies if the server authoritative inventory system is enabled. This
	// is a new system introduced in 1.16. Backwards compatibility with the inventory transactions has to
	// some extent been preserved, but will eventually be removed.
	ServerAuthoritativeInventory bool
	// GameVersion is the version of the game the server is running. The exact function of this field isn't clear.
	GameVersion string
	// PropertyData contains properties that should be applied on the player. These properties are the same as the
	// ones that are sent in the SyncActorProperty packet.
	PropertyData map[string]any
	// ServerBlockStateChecksum is a checksum to ensure block states between the server and client match.
	// This can simply be left empty, and the client will avoid trying to verify it.
	ServerBlockStateChecksum uint64
	// ClientSideGeneration is true if the client should use the features registered in the FeatureRegistry packet to
	// generate terrain client-side to save on bandwidth.
	ClientSideGeneration bool
	// WorldTemplateID is a UUID that identifies the template that was used to generate the world. Servers that do not
	// use a world based off of a template can set this to an empty UUID.
	WorldTemplateID uuid.UUID
	// ChatRestrictionLevel specifies the level of restriction on in-game chat. It is one of the constants above.
	ChatRestrictionLevel uint8
	// DisablePlayerInteractions is true if the client should ignore other players when interacting with the world.
	DisablePlayerInteractions bool
	// UseBlockNetworkIDHashes is true if the client should use the hash of a block's name as its network ID rather than
	// its index in the expected block palette. This is useful for servers that wish to support multiple protocol versions
	// and custom blocks, but it will result in extra bytes being written for every block in a sub chunk palette.
	UseBlockNetworkIDHashes bool
	// ServerAuthoritativeSound is currently unknown as to what it does.
	ServerAuthoritativeSound bool
}

// ID ...
func (*StartGame) ID() uint32 {
	return IDStartGame
}

func (pk *StartGame) Marshal(io protocol.IO) {
	io.Varint64(&pk.EntityUniqueID)
	io.Varuint64(&pk.EntityRuntimeID)
	io.Varint32(&pk.PlayerGameMode)
	io.Vec3(&pk.PlayerPosition)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Int64(&pk.WorldSeed)
	io.Int16(&pk.SpawnBiomeType)
	io.String(&pk.UserDefinedBiomeName)
	io.Varint32(&pk.Dimension)
	io.Varint32(&pk.Generator)
	io.Varint32(&pk.WorldGameMode)
	io.Bool(&pk.Hardcore)
	io.Varint32(&pk.Difficulty)
	io.UBlockPos(&pk.WorldSpawn)
	io.Bool(&pk.AchievementsDisabled)
	io.Varint32(&pk.EditorWorldType)
	io.Bool(&pk.CreatedInEditor)
	io.Bool(&pk.ExportedFromEditor)
	io.Varint32(&pk.DayCycleLockTime)
	io.Varint32(&pk.EducationEditionOffer)
	io.Bool(&pk.EducationFeaturesEnabled)
	io.String(&pk.EducationProductID)
	io.Float32(&pk.RainLevel)
	io.Float32(&pk.LightningLevel)
	io.Bool(&pk.ConfirmedPlatformLockedContent)
	io.Bool(&pk.MultiPlayerGame)
	io.Bool(&pk.LANBroadcastEnabled)
	io.Varint32(&pk.XBLBroadcastMode)
	io.Varint32(&pk.PlatformBroadcastMode)
	io.Bool(&pk.CommandsEnabled)
	io.Bool(&pk.TexturePackRequired)
	protocol.FuncSlice(io, &pk.GameRules, io.GameRule)
	protocol.SliceUint32Length(io, &pk.Experiments)
	io.Bool(&pk.ExperimentsPreviouslyToggled)
	io.Bool(&pk.BonusChestEnabled)
	io.Bool(&pk.StartWithMapEnabled)
	io.Varint32(&pk.PlayerPermissions)
	io.Int32(&pk.ServerChunkTickRadius)
	io.Bool(&pk.HasLockedBehaviourPack)
	io.Bool(&pk.HasLockedTexturePack)
	io.Bool(&pk.FromLockedWorldTemplate)
	io.Bool(&pk.MSAGamerTagsOnly)
	io.Bool(&pk.FromWorldTemplate)
	io.Bool(&pk.WorldTemplateSettingsLocked)
	io.Bool(&pk.OnlySpawnV1Villagers)
	io.Bool(&pk.PersonaDisabled)
	io.Bool(&pk.CustomSkinsDisabled)
	io.Bool(&pk.EmoteChatMuted)
	io.String(&pk.BaseGameVersion)
	io.Int32(&pk.LimitedWorldWidth)
	io.Int32(&pk.LimitedWorldDepth)
	io.Bool(&pk.NewNether)
	protocol.Single(io, &pk.EducationSharedResourceURI)
	protocol.OptionalFunc(io, &pk.ForceExperimentalGameplay, io.Bool)
	io.Uint8(&pk.ChatRestrictionLevel)
	io.Bool(&pk.DisablePlayerInteractions)
	io.String(&pk.LevelID)
	io.String(&pk.WorldName)
	io.String(&pk.TemplateContentIdentity)
	io.Bool(&pk.Trial)
	protocol.PlayerMoveSettings(io, &pk.PlayerMovementSettings)
	io.Int64(&pk.Time)
	io.Varint32(&pk.EnchantmentSeed)
	protocol.Slice(io, &pk.Blocks)
	protocol.Slice(io, &pk.Items)
	io.String(&pk.MultiPlayerCorrelationID)
	io.Bool(&pk.ServerAuthoritativeInventory)
	io.String(&pk.GameVersion)
	io.NBT(&pk.PropertyData, nbt.NetworkLittleEndian)
	io.Uint64(&pk.ServerBlockStateChecksum)
	io.UUID(&pk.WorldTemplateID)
	io.Bool(&pk.ClientSideGeneration)
	io.Bool(&pk.UseBlockNetworkIDHashes)
	io.Bool(&pk.ServerAuthoritativeSound)
}
