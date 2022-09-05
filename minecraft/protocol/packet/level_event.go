package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

//noinspection SpellCheckingInspection
const (
	LevelEventSoundClick                   = 1000
	LevelEventSoundClickFail               = 1001
	LevelEventSoundLaunch                  = 1002
	LevelEventSoundOpenDoor                = 1003
	LevelEventSoundFizz                    = 1004
	LevelEventSoundFuse                    = 1005
	LevelEventSoundPlayRecording           = 1006
	LevelEventSoundGhastWarning            = 1007
	LevelEventSoundGhastFireball           = 1008
	LevelEventSoundBlazeFireball           = 1009
	LevelEventSoundZombieWoodenDoor        = 1010
	LevelEventSoundZombieDoorCrash         = 1012
	LevelEventSoundZombieInfected          = 1016
	LevelEventSoundZombieConverted         = 1017
	LevelEventSoundEndermanTeleport        = 1018
	LevelEventSoundAnvilBroken             = 1020
	LevelEventSoundAnvilUsed               = 1021
	LevelEventSoundAnvilLand               = 1022
	LevelEventSoundInfinityArrowPickup     = 1030
	LevelEventSoundTeleportEnderPearl      = 1032
	LevelEventSoundAddItem                 = 1040
	LevelEventSoundItemFrameBreak          = 1041
	LevelEventSoundItemFramePlace          = 1042
	LevelEventSoundItemFrameRemoveItem     = 1043
	LevelEventSoundItemFrameRotateItem     = 1044
	LevelEventSoundExperienceOrbPickup     = 1051
	LevelEventSoundTotemUsed               = 1052
	LevelEventSoundArmorStandBreak         = 1060
	LevelEventSoundArmorStandHit           = 1061
	LevelEventSoundArmorStandLand          = 1062
	LevelEventSoundArmorStandPlace         = 1063
	LevelEventSoundPointedDripstoneLand    = 1064
	LevelEventSoundDyeUsed                 = 1065
	LevelEventSoundInkSacUsed              = 1066
	LevelEventQueueCustomMusic             = 1900
	LevelEventPlayCustomMusic              = 1901
	LevelEventStopCustomMusic              = 1902
	LevelEventSetMusicVolume               = 1903
	LevelEventParticlesShoot               = 2000
	LevelEventParticlesDestroyBlock        = 2001
	LevelEventParticlesPotionSplash        = 2002
	LevelEventParticlesEyeOfEnderDeath     = 2003
	LevelEventParticlesMobBlockSpawn       = 2004
	LevelEventParticleCropGrowth           = 2005
	LevelEventParticleSoundGuardianGhost   = 2006
	LevelEventParticleDeathSmoke           = 2007
	LevelEventParticleDenyBlock            = 2008
	LevelEventParticleGenericSpawn         = 2009
	LevelEventParticlesDragonEgg           = 2010
	LevelEventParticlesCropEaten           = 2011
	LevelEventParticlesCritical            = 2012
	LevelEventParticlesTeleport            = 2013
	LevelEventParticlesCrackBlock          = 2014
	LevelEventParticlesBubble              = 2015
	LevelEventParticlesEvaporate           = 2016
	LevelEventParticlesDestroyArmorStand   = 2017
	LevelEventParticlesBreakingEgg         = 2018
	LevelEventParticleDestroyEgg           = 2019
	LevelEventParticlesEvaporateWater      = 2020
	LevelEventParticlesDestroyBlockNoSound = 2021
	LevelEventParticlesKnockbackRoar       = 2022
	LevelEventParticlesTeleportTrail       = 2023
	LevelEventParticlesPointCloud          = 2024
	LevelEventParticlesExplosion           = 2025
	LevelEventParticlesBlockExplosion      = 2026
	LevelEventParticlesVibrationSignal     = 2027
	LevelEventParticlesDripstoneDrip       = 2028
	LevelEventParticlesFizzEffect          = 2029
	LevelEventWaxOn                        = 2030
	LevelEventWaxOff                       = 2031
	LevelEventScrape                       = 2032
	LevelEventParticlesElectricSpark       = 2033
	LevelEventParticleTurtleEgg            = 2034
	LevelEventParticleSculkShriek          = 2035
	LevelEventSculkCatalystBloom           = 2036
	LevelEventSculkCharge                  = 2037
	LevelEventSculkChargePop               = 2038
	LevelEventSonicExplosion               = 2039
	LevelEventStartRaining                 = 3001
	LevelEventStartThunderstorm            = 3002
	LevelEventStopRaining                  = 3003
	LevelEventStopThunderstorm             = 3004
	LevelEventGlobalPause                  = 3005
	LevelEventSimTimeStep                  = 3006
	LevelEventSimTimeScale                 = 3007
	LevelEventActivateBlock                = 3500
	LevelEventCauldronExplode              = 3501
	LevelEventCauldronDyeArmor             = 3502
	LevelEventCauldronCleanArmor           = 3503
	LevelEventCauldronFillPotion           = 3504
	LevelEventCauldronTakePotion           = 3505
	LevelEventCauldronFillWater            = 3506
	LevelEventCauldronTakeWater            = 3507
	LevelEventCauldronAddDye               = 3508
	LevelEventCauldronCleanBanner          = 3509
	LevelEventCauldronFlush                = 3510
	LevelEventAgentSpawnEffect             = 3511
	LevelEventCauldronFillLava             = 3512
	LevelEventCauldronTakeLava             = 3513
	LevelEventCauldronFillPowderSnow       = 3514
	LevelEventCauldronTakePowderSnow       = 3515
	LevelEventStartBlockCracking           = 3600
	LevelEventStopBlockCracking            = 3601
	LevelEventUpdateBlockCracking          = 3602
	LevelEventAllPlayersSleeping           = 9800
	LevelEventSleepingPlayers              = 9801
	LevelEventJumpPrevented                = 9810
	LevelEventParticleLegacyEvent          = 0x4000
)

// LevelEvent is sent by the server to make a certain event in the level occur. It ranges from particles, to
// sounds, and other events such as starting rain and block breaking.
type LevelEvent struct {
	// EventType is the ID of the event that is being 'called'. It is one of the events found in the constants
	// above.
	EventType int32
	// Position is the position of the level event. Practically every event requires this Vec3 set for it, as
	// particles, sounds and block editing relies on it.
	Position mgl32.Vec3
	// EventData is an integer holding additional data of the event. The type of data held depends on the
	// EventType.
	EventData int32
}

// ID ...
func (*LevelEvent) ID() uint32 {
	return IDLevelEvent
}

// Marshal ...
func (pk *LevelEvent) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.EventType)
	w.Vec3(&pk.Position)
	w.Varint32(&pk.EventData)
}

// Unmarshal ...
func (pk *LevelEvent) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.EventType)
	r.Vec3(&pk.Position)
	r.Varint32(&pk.EventData)
}
