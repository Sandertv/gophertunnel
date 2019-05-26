package packet

import (
	"bytes"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	EventSoundClick     = 1000
	EventSoundClickFail = 1001
	EventSoundShoot     = 1002
	EventSoundDoor      = 1003
	EventSoundFizz      = 1004
	EventSoundIgnite    = 1005

	EventSoundGhast      = 1007
	EventSoundGhastShoot = 1008
	EventSoundBlazeShoot = 1009
	EventSoundDoorBump   = 1010

	EventSoundDoorCrash = 1012

	EventSoundEndermanTeleport = 1018

	EventSoundAnvilBreak = 1020
	EventSoundAnvilUse   = 1021
	EventSoundAnvilFall  = 1022

	EventSoundPop = 1030

	EventSoundPortal = 1032

	EventSoundItemFrameAddItem    = 1040
	EventSoundItemFrameBreak      = 1041
	EventSoundItemFramePlace      = 1042
	EventSoundItemFrameRemoveItem = 1043
	EventSoundItemFrameRotateItem = 1044

	EventSoundCamera = 1050
	EventSoundOrb    = 1051
	EventSoundTotem  = 1052

	EventSoundArmourStandBreak = 1060
	EventSoundArmourStandHit   = 1061
	EventSoundArmourStandFall  = 1062
	EventSoundArmourStandPlace = 1063

	EventParticleShoot      = 2000
	EventParticleDestroy    = 2001
	EventParticleSplash     = 2002
	EventParticleEyeDespawn = 2003
	EventParticleSpawn      = 2004

	EventGuardianCurse = 2006

	EventParticleBlockForceField   = 2008
	EventParticleProjectileHit     = 2009
	EventParticleDragonEggTeleport = 2010

	EventParticleEndermanTeleport = 2013
	EventParticlePunchBlock       = 2014

	EventStartRain    = 3001
	EventStartThunder = 3002
	EventStopRain     = 3003
	EventStopThunder  = 3004
	EventPauseGame    = 3005

	EventRedstoneTrigger     = 3500
	EventCauldronExplode     = 3501
	EventCauldronDyeArmour   = 3502
	EventCauldronCleanArmour = 3503
	EventCauldronFillPotion  = 3504
	EventCauldronTakePotion  = 3505
	EventCauldronFillWater   = 3506
	EventCauldronTakeWater   = 3507
	EventCauldronAddDye      = 3508
	EventCauldronCleanBanner = 3509

	EventBlockStartBreak = 3600
	EventBlockStopBreak  = 3601

	EventSetData = 4000

	EventPlayersSleeping = 9800

	EventAddParticleMask = 0x4000
)

// LevelEvent is sent by the server to make a certain event in the level occur. It ranges from particles, to
// sounds, and other events such as starting rain and block breaking.
type LevelEvent struct {
	// EventID is the ID of the event that is being 'called'. It is one of the events found in the constants
	// above.
	EventID int32
	// Position is the position of the level event. Practically every event requires this Vec3 set for it, as
	// particles, sounds and block editing relies on it.
	Position mgl32.Vec3
	// EventData is an integer holding additional data of the event. The type of data held depends on the
	// EventID.
	EventData int32
}

// ID ...
func (*LevelEvent) ID() uint32 {
	return IDLevelEvent
}

// Marshal ...
func (pk *LevelEvent) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint32(buf, pk.EventID)
	_ = protocol.WriteVec3(buf, pk.Position)
	_ = protocol.WriteVarint32(buf, pk.EventData)
}

// Unmarshal ...
func (pk *LevelEvent) Unmarshal(buf *bytes.Buffer) error {
	return ChainErr(
		protocol.Varint32(buf, &pk.EventID),
		protocol.Vec3(buf, &pk.Position),
		protocol.Varint32(buf, &pk.EventData),
	)
}
