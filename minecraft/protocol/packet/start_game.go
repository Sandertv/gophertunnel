package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/go-gl/mathgl/mgl32"
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
	WorldSeed int32
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
	GameRules map[string]interface{}
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
	// BaseGameVersion is the version of the game from which Vanilla features will be used. The exact function
	// of this field isn't clear.
	BaseGameVersion string
	// LimitedWorldWidth and LimitedWorldDepth are the dimensions of the world if the world is a limited
	// world. For unlimited worlds, these may simply be left as 0.
	LimitedWorldWidth, LimitedWorldDepth int32
	// NewNether specifies if the server runs with the new nether introduced in the 1.16 update.
	NewNether bool
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
	// ServerAuthoritativeMovement specifies if the server is authoritative over the movement of the player,
	// meaning it controls the movement of it.
	// In reality, the only thing that changes when this field is set to true is the packet sent by the player
	// when it moves. When set to true, it will send the PlayerAuthInput packet instead of the MovePlayer
	// packet.
	ServerAuthoritativeMovement bool
	// Time is the total time that has elapsed since the start of the world.
	Time int64
	// EnchantmentSeed is the seed used to seed the random used to produce enchantments in the enchantment
	// table. Note that the exact correct random implementation must be used to produce the correct results
	// both client- and server-side.
	EnchantmentSeed int32
	// Blocks is a list of all blocks registered on the server.
	Blocks []interface{}
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
}

// ID ...
func (*StartGame) ID() uint32 {
	return IDStartGame
}

// Marshal ...
func (pk *StartGame) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint64(buf, pk.EntityUniqueID)
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	_ = protocol.WriteVarint32(buf, pk.PlayerGameMode)
	_ = protocol.WriteVec3(buf, pk.PlayerPosition)
	_ = protocol.WriteFloat32(buf, pk.Pitch)
	_ = protocol.WriteFloat32(buf, pk.Yaw)
	_ = protocol.WriteVarint32(buf, pk.WorldSeed)
	_ = binary.Write(buf, binary.LittleEndian, pk.SpawnBiomeType)
	_ = protocol.WriteString(buf, pk.UserDefinedBiomeName)
	_ = protocol.WriteVarint32(buf, pk.Dimension)
	_ = protocol.WriteVarint32(buf, pk.Generator)
	_ = protocol.WriteVarint32(buf, pk.WorldGameMode)
	_ = protocol.WriteVarint32(buf, pk.Difficulty)
	_ = protocol.WriteUBlockPosition(buf, pk.WorldSpawn)
	_ = binary.Write(buf, binary.LittleEndian, pk.AchievementsDisabled)
	_ = protocol.WriteVarint32(buf, pk.DayCycleLockTime)
	_ = protocol.WriteVarint32(buf, pk.EducationEditionOffer)
	_ = binary.Write(buf, binary.LittleEndian, pk.EducationFeaturesEnabled)
	_ = protocol.WriteString(buf, pk.EducationProductID)
	_ = protocol.WriteFloat32(buf, pk.RainLevel)
	_ = protocol.WriteFloat32(buf, pk.LightningLevel)
	_ = binary.Write(buf, binary.LittleEndian, pk.ConfirmedPlatformLockedContent)
	_ = binary.Write(buf, binary.LittleEndian, pk.MultiPlayerGame)
	_ = binary.Write(buf, binary.LittleEndian, pk.LANBroadcastEnabled)
	_ = protocol.WriteVarint32(buf, pk.XBLBroadcastMode)
	_ = protocol.WriteVarint32(buf, pk.PlatformBroadcastMode)
	_ = binary.Write(buf, binary.LittleEndian, pk.CommandsEnabled)
	_ = binary.Write(buf, binary.LittleEndian, pk.TexturePackRequired)
	_ = protocol.WriteGameRules(buf, pk.GameRules)
	_ = binary.Write(buf, binary.LittleEndian, pk.BonusChestEnabled)
	_ = binary.Write(buf, binary.LittleEndian, pk.StartWithMapEnabled)
	_ = protocol.WriteVarint32(buf, pk.PlayerPermissions)
	_ = binary.Write(buf, binary.LittleEndian, pk.ServerChunkTickRadius)
	_ = binary.Write(buf, binary.LittleEndian, pk.HasLockedBehaviourPack)
	_ = binary.Write(buf, binary.LittleEndian, pk.HasLockedTexturePack)
	_ = binary.Write(buf, binary.LittleEndian, pk.FromLockedWorldTemplate)
	_ = binary.Write(buf, binary.LittleEndian, pk.MSAGamerTagsOnly)
	_ = binary.Write(buf, binary.LittleEndian, pk.FromWorldTemplate)
	_ = binary.Write(buf, binary.LittleEndian, pk.WorldTemplateSettingsLocked)
	_ = binary.Write(buf, binary.LittleEndian, pk.OnlySpawnV1Villagers)
	_ = protocol.WriteString(buf, pk.BaseGameVersion)
	_ = binary.Write(buf, binary.LittleEndian, pk.LimitedWorldWidth)
	_ = binary.Write(buf, binary.LittleEndian, pk.LimitedWorldDepth)
	_ = binary.Write(buf, binary.LittleEndian, pk.NewNether)
	_ = binary.Write(buf, binary.LittleEndian, pk.ForceExperimentalGameplay)
	_ = protocol.WriteString(buf, pk.LevelID)
	_ = protocol.WriteString(buf, pk.WorldName)
	_ = protocol.WriteString(buf, pk.TemplateContentIdentity)
	_ = binary.Write(buf, binary.LittleEndian, pk.Trial)
	_ = binary.Write(buf, binary.LittleEndian, pk.ServerAuthoritativeMovement)
	_ = binary.Write(buf, binary.LittleEndian, pk.Time)
	_ = protocol.WriteVarint32(buf, pk.EnchantmentSeed)
	_ = nbt.NewEncoder(buf).Encode(pk.Blocks)
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Items)))
	for _, item := range pk.Items {
		_ = protocol.WriteString(buf, item.Name)
		_ = binary.Write(buf, binary.LittleEndian, item.LegacyID)
	}
	_ = protocol.WriteString(buf, pk.MultiPlayerCorrelationID)
	_ = binary.Write(buf, binary.LittleEndian, pk.ServerAuthoritativeInventory)
}

// Unmarshal ...
func (pk *StartGame) Unmarshal(buf *bytes.Buffer) error {
	if pk.GameRules == nil {
		pk.GameRules = make(map[string]interface{})
	}
	var itemCount uint32
	if err := chainErr(
		protocol.Varint64(buf, &pk.EntityUniqueID),
		protocol.Varuint64(buf, &pk.EntityRuntimeID),
		protocol.Varint32(buf, &pk.PlayerGameMode),
		protocol.Vec3(buf, &pk.PlayerPosition),
		protocol.Float32(buf, &pk.Pitch),
		protocol.Float32(buf, &pk.Yaw),
		protocol.Varint32(buf, &pk.WorldSeed),
		binary.Read(buf, binary.LittleEndian, &pk.SpawnBiomeType),
		protocol.String(buf, &pk.UserDefinedBiomeName),
		protocol.Varint32(buf, &pk.Dimension),
		protocol.Varint32(buf, &pk.Generator),
		protocol.Varint32(buf, &pk.WorldGameMode),
		protocol.Varint32(buf, &pk.Difficulty),
		protocol.UBlockPosition(buf, &pk.WorldSpawn),
		binary.Read(buf, binary.LittleEndian, &pk.AchievementsDisabled),
		protocol.Varint32(buf, &pk.DayCycleLockTime),
		protocol.Varint32(buf, &pk.EducationEditionOffer),
		binary.Read(buf, binary.LittleEndian, &pk.EducationFeaturesEnabled),
		protocol.String(buf, &pk.EducationProductID),
		protocol.Float32(buf, &pk.RainLevel),
		protocol.Float32(buf, &pk.LightningLevel),
		binary.Read(buf, binary.LittleEndian, &pk.ConfirmedPlatformLockedContent),
		binary.Read(buf, binary.LittleEndian, &pk.MultiPlayerGame),
		binary.Read(buf, binary.LittleEndian, &pk.LANBroadcastEnabled),
		protocol.Varint32(buf, &pk.XBLBroadcastMode),
		protocol.Varint32(buf, &pk.PlatformBroadcastMode),
		binary.Read(buf, binary.LittleEndian, &pk.CommandsEnabled),
		binary.Read(buf, binary.LittleEndian, &pk.TexturePackRequired),
		protocol.GameRules(buf, &pk.GameRules),
		binary.Read(buf, binary.LittleEndian, &pk.BonusChestEnabled),
		binary.Read(buf, binary.LittleEndian, &pk.StartWithMapEnabled),
		protocol.Varint32(buf, &pk.PlayerPermissions),
		binary.Read(buf, binary.LittleEndian, &pk.ServerChunkTickRadius),
		binary.Read(buf, binary.LittleEndian, &pk.HasLockedBehaviourPack),
		binary.Read(buf, binary.LittleEndian, &pk.HasLockedTexturePack),
		binary.Read(buf, binary.LittleEndian, &pk.FromLockedWorldTemplate),
		binary.Read(buf, binary.LittleEndian, &pk.MSAGamerTagsOnly),
		binary.Read(buf, binary.LittleEndian, &pk.FromWorldTemplate),
		binary.Read(buf, binary.LittleEndian, &pk.WorldTemplateSettingsLocked),
		binary.Read(buf, binary.LittleEndian, &pk.OnlySpawnV1Villagers),
		protocol.String(buf, &pk.BaseGameVersion),
		binary.Read(buf, binary.LittleEndian, &pk.LimitedWorldWidth),
		binary.Read(buf, binary.LittleEndian, &pk.LimitedWorldDepth),
		binary.Read(buf, binary.LittleEndian, &pk.NewNether),
		binary.Read(buf, binary.LittleEndian, &pk.ForceExperimentalGameplay),
		protocol.String(buf, &pk.LevelID),
		protocol.String(buf, &pk.WorldName),
		protocol.String(buf, &pk.TemplateContentIdentity),
		binary.Read(buf, binary.LittleEndian, &pk.Trial),
		binary.Read(buf, binary.LittleEndian, &pk.ServerAuthoritativeMovement),
		binary.Read(buf, binary.LittleEndian, &pk.Time),
		protocol.Varint32(buf, &pk.EnchantmentSeed),
		nbt.NewDecoder(buf).Decode(&pk.Blocks),
		protocol.Varuint32(buf, &itemCount),
	); err != nil {
		return err
	}
	pk.Items = make([]protocol.ItemEntry, itemCount)
	for i := uint32(0); i < itemCount; i++ {
		item := protocol.ItemEntry{}
		if err := chainErr(
			protocol.String(buf, &item.Name),
			binary.Read(buf, binary.LittleEndian, &item.LegacyID),
		); err != nil {
			return err
		}
		pk.Items[i] = item
	}
	return chainErr(
		protocol.String(buf, &pk.MultiPlayerCorrelationID),
		binary.Read(buf, binary.LittleEndian, &pk.ServerAuthoritativeInventory),
	)
}
