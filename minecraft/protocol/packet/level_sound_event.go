package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	SoundEventItemUseOn = iota
	SoundEventHit
	SoundEventStep
	SoundEventFly
	SoundEventJump
	SoundEventBreak
	SoundEventPlace
	SoundEventHeavyStep
	SoundEventGallop
	SoundEventFall
	SoundEventAmbient
	SoundEventAmbientBaby
	SoundEventAmbientInWater
	SoundEventBreathe
	SoundEventDeath
	SoundEventDeathInWater
	SoundEventDeathToZombie
	SoundEventHurt
	SoundEventHurtInWater
	SoundEventMad
	SoundEventBoost
	SoundEventBow
	SoundEventSquishBig
	SoundEventSquishSmall
	SoundEventFallBig
	SoundEventFallSmall
	SoundEventSplash
	SoundEventFizz
	SoundEventFlap
	SoundEventSwim
	SoundEventDrink
	SoundEventEat
	SoundEventTakeoff
	SoundEventShake
	SoundEventPlop
	SoundEventLand
	SoundEventSaddle
	SoundEventArmor
	SoundEventMobArmorStandPlace
	SoundEventAddChest
	SoundEventThrow
	SoundEventAttack
	SoundEventAttackNodamage
	SoundEventAttackStrong
	SoundEventWarn
	SoundEventShear
	SoundEventMilk
	SoundEventThunder
	SoundEventExplode
	SoundEventFire
	SoundEventIgnite
	SoundEventFuse
	SoundEventStare
	SoundEventSpawn
	SoundEventShoot
	SoundEventBreakBlock
	SoundEventLaunch
	SoundEventBlast
	SoundEventLargeBlast
	SoundEventTwinkle
	SoundEventRemedy
	SoundEventUnfect
	SoundEventLevelup
	SoundEventBowHit
	SoundEventBulletHit
	SoundEventExtinguishFire
	SoundEventItemFizz
	SoundEventChestOpen
	SoundEventChestClosed
	SoundEventShulkerboxOpen
	SoundEventShulkerboxClosed
	SoundEventEnderchestOpen
	SoundEventEnderchestClosed
	SoundEventPowerOn
	SoundEventPowerOff
	SoundEventAttach
	SoundEventDetach
	SoundEventDeny
	SoundEventTripod
	SoundEventPop
	SoundEventDropSlot
	SoundEventNote
	SoundEventThorns
	SoundEventPistonIn
	SoundEventPistonOut
	SoundEventPortal
	SoundEventWater
	SoundEventLavaPop
	SoundEventLava
	SoundEventBurp
	SoundEventBucketFillWater
	SoundEventBucketFillLava
	SoundEventBucketEmptyWater
	SoundEventBucketEmptyLava
	SoundEventArmorEquipChain
	SoundEventArmorEquipDiamond
	SoundEventArmorEquipGeneric
	SoundEventArmorEquipGold
	SoundEventArmorEquipIron
	SoundEventArmorEquipLeather
	SoundEventArmorEquipElytra
	SoundEventRecord13
	SoundEventRecordCat
	SoundEventRecordBlocks
	SoundEventRecordChirp
	SoundEventRecordFar
	SoundEventRecordMall
	SoundEventRecordMellohi
	SoundEventRecordStal
	SoundEventRecordStrad
	SoundEventRecordWard
	SoundEventRecord11
	SoundEventRecordWait
	_
	SoundEventFlop
	SoundEventElderguardianCurse
	SoundEventMobWarning
	SoundEventMobWarningBaby
	SoundEventTeleport
	SoundEventShulkerOpen
	SoundEventShulkerClose
	SoundEventHaggle
	SoundEventHaggleYes
	SoundEventHaggleNo
	SoundEventHaggleIdle
	SoundEventChorusgrow
	SoundEventChorusdeath
	SoundEventGlass
	SoundEventPotionBrewed
	SoundEventCastSpell
	SoundEventPrepareAttack
	SoundEventPrepareSummon
	SoundEventPrepareWololo
	SoundEventFang
	SoundEventCharge
	SoundEventCameraTakePicture
	SoundEventLeashknotPlace
	SoundEventLeashknotBreak
	SoundEventGrowl
	SoundEventWhine
	SoundEventPant
	SoundEventPurr
	SoundEventPurreow
	SoundEventDeathMinVolume
	SoundEventDeathMidVolume
	_
	SoundEventImitateCaveSpider
	SoundEventImitateCreeper
	SoundEventImitateElderGuardian
	SoundEventImitateEnderDragon
	SoundEventImitateEnderman
	_
	SoundEventImitateEvocationIllager
	SoundEventImitateGhast
	SoundEventImitateHusk
	SoundEventImitateIllusionIllager
	SoundEventImitateMagmaCube
	SoundEventImitatePolarBear
	SoundEventImitateShulker
	SoundEventImitateSilverfish
	SoundEventImitateSkeleton
	SoundEventImitateSlime
	SoundEventImitateSpider
	SoundEventImitateStray
	SoundEventImitateVex
	SoundEventImitateVindicationIllager
	SoundEventImitateWitch
	SoundEventImitateWither
	SoundEventImitateWitherSkeleton
	SoundEventImitateWolf
	SoundEventImitateZombie
	SoundEventImitateZombiePigman
	SoundEventImitateZombieVillager
	SoundEventBlockEndPortalFrameFill
	SoundEventBlockEndPortalSpawn
	SoundEventRandomAnvilUse
	SoundEventBottleDragonbreath
	SoundEventPortalTravel
	SoundEventItemTridentHit
	SoundEventItemTridentReturn
	SoundEventItemTridentRiptide1
	SoundEventItemTridentRiptide2
	SoundEventItemTridentRiptide3
	SoundEventItemTridentThrow
	SoundEventItemTridentThunder
	SoundEventItemTridentHitGround
	SoundEventDefault
	SoundEventBlockFletchingTableUse
	SoundEventElemconstructOpen
	SoundEventIcebombHit
	SoundEventBalloonpop
	SoundEventLtReactionIcebomb
	SoundEventLtReactionBleach
	SoundEventLtReactionEpaste
	SoundEventLtReactionEpaste2
	_
	_
	_
	_
	SoundEventLtReactionFertilizer
	SoundEventLtReactionFireball
	SoundEventLtReactionMgsalt
	SoundEventLtReactionMiscfire
	SoundEventLtReactionFire
	SoundEventLtReactionMiscexplosion
	SoundEventLtReactionMiscmystical
	SoundEventLtReactionMiscmystical2
	SoundEventLtReactionProduct
	SoundEventSparklerUse
	SoundEventGlowstickUse
	SoundEventSparklerActive
	SoundEventConvertToDrowned
	SoundEventBucketFillFish
	SoundEventBucketEmptyFish
	SoundEventBubbleUp
	SoundEventBubbleDown
	SoundEventBubblePop
	SoundEventBubbleUpinside
	SoundEventBubbleDowninside
	SoundEventHurtBaby
	SoundEventDeathBaby
	SoundEventStepBaby
	_
	SoundEventBorn
	SoundEventBlockTurtleEggBreak
	SoundEventBlockTurtleEggCrack
	SoundEventBlockTurtleEggHatch
	_
	SoundEventBlockTurtleEggAttack
	SoundEventBeaconActivate
	SoundEventBeaconAmbient
	SoundEventBeaconDeactivate
	SoundEventBeaconPower
	SoundEventConduitActivate
	SoundEventConduitAmbient
	SoundEventConduitAttack
	SoundEventConduitDeactivate
	SoundEventConduitShort
	SoundEventSwoop
	SoundEventBlockBambooSaplingPlace
	SoundEventPresneeze
	SoundEventSneeze
	SoundEventAmbientTame
	SoundEventScared
	SoundEventBlockScaffoldingClimb
	SoundEventCrossbowLoadingStart
	SoundEventCrossbowLoadingMiddle
	SoundEventCrossbowLoadingEnd
	SoundEventCrossbowShoot
	SoundEventCrossbowQuickChargeStart
	SoundEventCrossbowQuickChargeMiddle
	SoundEventCrossbowQuickChargeEnd
	SoundEventAmbientAggressive
	SoundEventAmbientWorried
	SoundEventCantBreed
	SoundEventItemShieldBlock
	SoundEventItemBookPut
	SoundEventBlockGrindstoneUse
	SoundEventBlockBellHit
	SoundEventBlockCampfireCrackle
	SoundEventRoar
	SoundEventStun
	SoundEventBlockSweetBerryBushHurt
	SoundEventBlockSweetBerryBushPick
	SoundEventUiCartographyTableTakeResult
	SoundEventUiStonecutterTakeResult
	SoundEventBlockComposterEmpty
	SoundEventBlockComposterFill
	SoundEventBlockComposterFillSuccess
	SoundEventBlockComposterReady
	SoundEventBlockBarrelOpen
	SoundEventBlockBarrelClose
	SoundEventRaidHorn
	SoundEventBlockLoomUse
	SoundEventUndefined
)

// LevelSoundEvent is sent by the server to make any kind of built-in sound heard to a player. It is sent to,
// for example, play a stepping sound or a shear sound. The packet is also sent by the client, in which case
// it could be forwarded by the server to the other players online. If possible, the packets from the client
// should be ignored however, and the server should play them on its own accord.
type LevelSoundEvent struct {
	// SoundType is the type of the sound to play. It is one of the constants above. Some of the sound types
	// require additional data, which is set in the EventData field.
	SoundType uint32
	// Position is the position of the sound event. The player will be able to hear the direction of the sound
	// based on what position is sent here.
	Position mgl32.Vec3
	// ExtraData is a packed integer that some sound types use to provide extra data. An example of this is
	// the note sound, which is composed of a pitch and an instrument type.
	ExtraData int32
	// EntityType is the string entity type of the entity that emitted the sound, for example
	// 'minecraft:skeleton'. Some sound types use this entity type for additional data.
	EntityType string
	// BabyMob specifies if the sound should be that of a baby mob. It is most notably used for parrot
	// imitations, which will change based on if this field is set to true or not.
	BabyMob bool
	// DisableRelativeVolume specifies if the sound should be played relatively or not. If set to true, the
	// sound will have full volume, regardless of where the Position is, whereas if set to false, the sound's
	// volume will be based on the distance to Position.
	DisableRelativeVolume bool
}

// ID ...
func (*LevelSoundEvent) ID() uint32 {
	return IDLevelSoundEvent
}

// Marshal ...
func (pk *LevelSoundEvent) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint32(buf, pk.SoundType)
	_ = protocol.WriteVec3(buf, pk.Position)
	_ = protocol.WriteVarint32(buf, pk.ExtraData)
	_ = protocol.WriteString(buf, pk.EntityType)
	_ = binary.Write(buf, binary.LittleEndian, pk.BabyMob)
	_ = binary.Write(buf, binary.LittleEndian, pk.DisableRelativeVolume)
}

// Unmarshal ...
func (pk *LevelSoundEvent) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.Varuint32(buf, &pk.SoundType),
		protocol.Vec3(buf, &pk.Position),
		protocol.Varint32(buf, &pk.ExtraData),
		protocol.String(buf, &pk.EntityType),
		binary.Read(buf, binary.LittleEndian, &pk.BabyMob),
		binary.Read(buf, binary.LittleEndian, &pk.DisableRelativeVolume),
	)
}
