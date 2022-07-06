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
	WorldSeed uint64
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
	// EditorWorld is a value to dictate if the world is in editor mode, a special mode recently introduced adding
	// "powerful tools for editing worlds, intended for experienced creators."
	EditorWorld bool
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
	PlayerPermissions uint8
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
	ForceExperimentalGameplay bool
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
	// WorldTemplateID is a UUID that identifies the template that was used to generate the world. Servers that do not
	// use a world based off of a template can set this to an empty UUID.
	WorldTemplateID uuid.UUID
}

// ID ...
func (*StartGame) ID() uint32 {
	return IDStartGame
}

// Marshal ...
func (pk *StartGame) Marshal(w *protocol.Writer) {
	w.Varint64(&pk.EntityUniqueID)
	w.Varuint64(&pk.EntityRuntimeID)
	w.Varint32(&pk.PlayerGameMode)
	w.Vec3(&pk.PlayerPosition)
	w.Float32(&pk.Pitch)
	w.Float32(&pk.Yaw)
	w.Uint64(&pk.WorldSeed)
	w.Int16(&pk.SpawnBiomeType)
	w.String(&pk.UserDefinedBiomeName)
	w.Varint32(&pk.Dimension)
	w.Varint32(&pk.Generator)
	w.Varint32(&pk.WorldGameMode)
	w.Varint32(&pk.Difficulty)
	w.UBlockPos(&pk.WorldSpawn)
	w.Bool(&pk.AchievementsDisabled)
	w.Bool(&pk.EditorWorld)
	w.Varint32(&pk.DayCycleLockTime)
	w.Varint32(&pk.EducationEditionOffer)
	w.Bool(&pk.EducationFeaturesEnabled)
	w.String(&pk.EducationProductID)
	w.Float32(&pk.RainLevel)
	w.Float32(&pk.LightningLevel)
	w.Bool(&pk.ConfirmedPlatformLockedContent)
	w.Bool(&pk.MultiPlayerGame)
	w.Bool(&pk.LANBroadcastEnabled)
	w.Varint32(&pk.XBLBroadcastMode)
	w.Varint32(&pk.PlatformBroadcastMode)
	w.Bool(&pk.CommandsEnabled)
	w.Bool(&pk.TexturePackRequired)
	protocol.WriteGameRules(w, &pk.GameRules)
	l := uint32(len(pk.Experiments))
	w.Uint32(&l)
	for _, experiment := range pk.Experiments {
		protocol.Experiment(w, &experiment)
	}
	w.Bool(&pk.ExperimentsPreviouslyToggled)
	w.Bool(&pk.BonusChestEnabled)
	w.Bool(&pk.StartWithMapEnabled)
	w.Uint8(&pk.PlayerPermissions)
	w.Int32(&pk.ServerChunkTickRadius)
	w.Bool(&pk.HasLockedBehaviourPack)
	w.Bool(&pk.HasLockedTexturePack)
	w.Bool(&pk.FromLockedWorldTemplate)
	w.Bool(&pk.MSAGamerTagsOnly)
	w.Bool(&pk.FromWorldTemplate)
	w.Bool(&pk.WorldTemplateSettingsLocked)
	w.Bool(&pk.OnlySpawnV1Villagers)
	w.String(&pk.BaseGameVersion)
	w.Int32(&pk.LimitedWorldWidth)
	w.Int32(&pk.LimitedWorldDepth)
	w.Bool(&pk.NewNether)
	protocol.EducationResourceURI(w, &pk.EducationSharedResourceURI)
	w.Bool(&pk.ForceExperimentalGameplay)
	if pk.ForceExperimentalGameplay {
		// This might look wrong, but is in fact correct: Mojang is writing this bool if the same bool above
		// is set to true.
		w.Bool(&pk.ForceExperimentalGameplay)
	}
	w.String(&pk.LevelID)
	w.String(&pk.WorldName)
	w.String(&pk.TemplateContentIdentity)
	w.Bool(&pk.Trial)
	protocol.PlayerMoveSettings(w, &pk.PlayerMovementSettings)
	w.Int64(&pk.Time)
	w.Varint32(&pk.EnchantmentSeed)

	l = uint32(len(pk.Blocks))
	w.Varuint32(&l)
	for i := range pk.Blocks {
		protocol.Block(w, &pk.Blocks[i])
	}

	l = uint32(len(pk.Items))
	w.Varuint32(&l)
	for i := range pk.Items {
		protocol.Item(w, &pk.Items[i])
	}
	w.String(&pk.MultiPlayerCorrelationID)
	w.Bool(&pk.ServerAuthoritativeInventory)
	w.String(&pk.GameVersion)
	w.NBT(&pk.PropertyData, nbt.NetworkLittleEndian)
	w.Uint64(&pk.ServerBlockStateChecksum)
	w.UUID(&pk.WorldTemplateID)
}

// Unmarshal ...
func (pk *StartGame) Unmarshal(r *protocol.Reader) {
	var blockCount, itemCount uint32
	r.Varint64(&pk.EntityUniqueID)
	r.Varuint64(&pk.EntityRuntimeID)
	r.Varint32(&pk.PlayerGameMode)
	r.Vec3(&pk.PlayerPosition)
	r.Float32(&pk.Pitch)
	r.Float32(&pk.Yaw)
	r.Uint64(&pk.WorldSeed)
	r.Int16(&pk.SpawnBiomeType)
	r.String(&pk.UserDefinedBiomeName)
	r.Varint32(&pk.Dimension)
	r.Varint32(&pk.Generator)
	r.Varint32(&pk.WorldGameMode)
	r.Varint32(&pk.Difficulty)
	r.UBlockPos(&pk.WorldSpawn)
	r.Bool(&pk.AchievementsDisabled)
	r.Bool(&pk.EditorWorld)
	r.Varint32(&pk.DayCycleLockTime)
	r.Varint32(&pk.EducationEditionOffer)
	r.Bool(&pk.EducationFeaturesEnabled)
	r.String(&pk.EducationProductID)
	r.Float32(&pk.RainLevel)
	r.Float32(&pk.LightningLevel)
	r.Bool(&pk.ConfirmedPlatformLockedContent)
	r.Bool(&pk.MultiPlayerGame)
	r.Bool(&pk.LANBroadcastEnabled)
	r.Varint32(&pk.XBLBroadcastMode)
	r.Varint32(&pk.PlatformBroadcastMode)
	r.Bool(&pk.CommandsEnabled)
	r.Bool(&pk.TexturePackRequired)
	protocol.GameRules(r, &pk.GameRules)
	var l uint32
	r.Uint32(&l)
	pk.Experiments = make([]protocol.ExperimentData, l)
	for i := uint32(0); i < l; i++ {
		protocol.Experiment(r, &pk.Experiments[i])
	}
	r.Bool(&pk.ExperimentsPreviouslyToggled)
	r.Bool(&pk.BonusChestEnabled)
	r.Bool(&pk.StartWithMapEnabled)
	r.Uint8(&pk.PlayerPermissions)
	r.Int32(&pk.ServerChunkTickRadius)
	r.Bool(&pk.HasLockedBehaviourPack)
	r.Bool(&pk.HasLockedTexturePack)
	r.Bool(&pk.FromLockedWorldTemplate)
	r.Bool(&pk.MSAGamerTagsOnly)
	r.Bool(&pk.FromWorldTemplate)
	r.Bool(&pk.WorldTemplateSettingsLocked)
	r.Bool(&pk.OnlySpawnV1Villagers)
	r.String(&pk.BaseGameVersion)
	r.Int32(&pk.LimitedWorldWidth)
	r.Int32(&pk.LimitedWorldDepth)
	r.Bool(&pk.NewNether)
	protocol.EducationResourceURI(r, &pk.EducationSharedResourceURI)
	r.Bool(&pk.ForceExperimentalGameplay)
	if pk.ForceExperimentalGameplay {
		// This might look wrong, but is in fact correct: Mojang is writing this bool if the same bool above
		// is set to true.
		r.Bool(&pk.ForceExperimentalGameplay)
	}
	r.String(&pk.LevelID)
	r.String(&pk.WorldName)
	r.String(&pk.TemplateContentIdentity)
	r.Bool(&pk.Trial)
	protocol.PlayerMoveSettings(r, &pk.PlayerMovementSettings)
	r.Int64(&pk.Time)
	r.Varint32(&pk.EnchantmentSeed)

	r.Varuint32(&blockCount)
	pk.Blocks = make([]protocol.BlockEntry, blockCount)
	for i := uint32(0); i < blockCount; i++ {
		protocol.Block(r, &pk.Blocks[i])
	}

	r.Varuint32(&itemCount)
	pk.Items = make([]protocol.ItemEntry, itemCount)
	for i := uint32(0); i < itemCount; i++ {
		protocol.Item(r, &pk.Items[i])
	}
	r.String(&pk.MultiPlayerCorrelationID)
	r.Bool(&pk.ServerAuthoritativeInventory)
	r.String(&pk.GameVersion)
	r.NBT(&pk.PropertyData, nbt.NetworkLittleEndian)
	r.Uint64(&pk.ServerBlockStateChecksum)
	r.UUID(&pk.WorldTemplateID)
}
