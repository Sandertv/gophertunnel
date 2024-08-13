package protocol

const (
	EntityDataKeyFlags = iota
	EntityDataKeyStructuralIntegrity
	EntityDataKeyVariant
	EntityDataKeyColorIndex
	EntityDataKeyName
	EntityDataKeyOwner
	EntityDataKeyTarget
	EntityDataKeyAirSupply
	EntityDataKeyEffectColor
	EntityDataKeyEffectAmbience
	EntityDataKeyJumpDuration
	EntityDataKeyHurt
	EntityDataKeyHurtDirection
	EntityDataKeyRowTimeLeft
	EntityDataKeyRowTimeRight
	EntityDataKeyValue
	EntityDataKeyDisplayTileRuntimeID
	EntityDataKeyDisplayOffset
	EntityDataKeyCustomDisplay
	EntityDataKeySwell
	EntityDataKeyOldSwell
	EntityDataKeySwellDirection
	EntityDataKeyChargeAmount
	EntityDataKeyCarryBlockRuntimeID
	EntityDataKeyClientEvent
	EntityDataKeyUsingItem
	EntityDataKeyPlayerFlags
	EntityDataKeyPlayerIndex
	EntityDataKeyBedPosition
	EntityDataKeyPowerX
	EntityDataKeyPowerY
	EntityDataKeyPowerZ
	EntityDataKeyAuxPower
	EntityDataKeyFishX
	EntityDataKeyFishZ
	EntityDataKeyFishAngle
	EntityDataKeyAuxValueData
	EntityDataKeyLeashHolder
	EntityDataKeyScale
	EntityDataKeyHasNPC
	EntityDataKeyNPCData
	EntityDataKeyActions
	EntityDataKeyAirSupplyMax
	EntityDataKeyMarkVariant
	EntityDataKeyContainerType
	EntityDataKeyContainerSize
	EntityDataKeyContainerStrengthModifier
	EntityDataKeyBlockTarget
	EntityDataKeyInventory
	EntityDataKeyTargetA
	EntityDataKeyTargetB
	EntityDataKeyTargetC
	EntityDataKeyAerialAttack
	EntityDataKeyWidth
	EntityDataKeyHeight
	EntityDataKeyFuseTime
	EntityDataKeySeatOffset
	EntityDataKeySeatLockPassengerRotation
	EntityDataKeySeatLockPassengerRotationDegrees
	EntityDataKeySeatRotationOffset
	EntityDataKeySeatRotationOffsetDegrees
	EntityDataKeyDataRadius
	EntityDataKeyDataWaiting
	EntityDataKeyDataParticle
	EntityDataKeyPeekID
	EntityDataKeyAttachFace
	EntityDataKeyAttached
	EntityDataKeyAttachedPosition
	EntityDataKeyTradeTarget
	EntityDataKeyCareer
	EntityDataKeyHasCommandBlock
	EntityDataKeyCommandName
	EntityDataKeyLastCommandOutput
	EntityDataKeyTrackCommandOutput
	EntityDataKeyControllingSeatIndex
	EntityDataKeyStrength
	EntityDataKeyStrengthMax
	EntityDataKeyDataSpellCastingColor
	EntityDataKeyDataLifetimeTicks
	EntityDataKeyPoseIndex
	EntityDataKeyDataTickOffset
	EntityDataKeyAlwaysShowNameTag
	EntityDataKeyColorTwoIndex
	EntityDataKeyNameAuthor
	EntityDataKeyScore
	EntityDataKeyBalloonAnchor
	EntityDataKeyPuffedState
	EntityDataKeyBubbleTime
	EntityDataKeyAgent
	EntityDataKeySittingAmount
	EntityDataKeySittingAmountPrevious
	EntityDataKeyEatingCounter
	EntityDataKeyFlagsTwo
	EntityDataKeyLayingAmount
	EntityDataKeyLayingAmountPrevious
	EntityDataKeyDataDuration
	EntityDataKeyDataSpawnTime
	EntityDataKeyDataChangeRate
	EntityDataKeyDataChangeOnPickup
	EntityDataKeyDataPickupCount
	EntityDataKeyInteractText
	EntityDataKeyTradeTier
	EntityDataKeyMaxTradeTier
	EntityDataKeyTradeExperience
	EntityDataKeySkinID
	EntityDataKeySpawningFrames
	EntityDataKeyCommandBlockTickDelay
	EntityDataKeyCommandBlockExecuteOnFirstTick
	EntityDataKeyAmbientSoundInterval
	EntityDataKeyAmbientSoundIntervalRange
	EntityDataKeyAmbientSoundEventName
	EntityDataKeyFallDamageMultiplier
	EntityDataKeyNameRawText
	EntityDataKeyCanRideTarget
	EntityDataKeyLowTierCuredTradeDiscount
	EntityDataKeyHighTierCuredTradeDiscount
	EntityDataKeyNearbyCuredTradeDiscount
	EntityDataKeyNearbyCuredDiscountTimeStamp
	EntityDataKeyHitBox
	EntityDataKeyIsBuoyant
	EntityDataKeyFreezingEffectStrength
	EntityDataKeyBuoyancyData
	EntityDataKeyGoatHornCount
	EntityDataKeyBaseRuntimeID
	EntityDataKeyMovementSoundDistanceOffset
	EntityDataKeyHeartbeatIntervalTicks
	EntityDataKeyHeartbeatSoundEvent
	EntityDataKeyPlayerLastDeathPosition
	EntityDataKeyPlayerLastDeathDimension
	EntityDataKeyPlayerHasDied
	EntityDataKeyCollisionBox
	EntityDataKeyVisibleMobEffects
)

const (
	EntityDataFlagOnFire = iota
	EntityDataFlagSneaking
	EntityDataFlagRiding
	EntityDataFlagSprinting
	EntityDataFlagUsingItem
	EntityDataFlagInvisible
	EntityDataFlagTempted
	EntityDataFlagInLove
	EntityDataFlagSaddled
	EntityDataFlagPowered
	EntityDataFlagIgnited
	EntityDataFlagBaby
	EntityDataFlagConverting
	EntityDataFlagCritical
	EntityDataFlagShowName
	EntityDataFlagAlwaysShowName
	EntityDataFlagNoAI
	EntityDataFlagSilent
	EntityDataFlagWallClimbing
	EntityDataFlagClimb
	EntityDataFlagSwim
	EntityDataFlagFly
	EntityDataFlagWalk
	EntityDataFlagResting
	EntityDataFlagSitting
	EntityDataFlagAngry
	EntityDataFlagInterested
	EntityDataFlagCharged
	EntityDataFlagTamed
	EntityDataFlagOrphaned
	EntityDataFlagLeashed
	EntityDataFlagSheared
	EntityDataFlagGliding
	EntityDataFlagElder
	EntityDataFlagMoving
	EntityDataFlagBreathing
	EntityDataFlagChested
	EntityDataFlagStackable
	EntityDataFlagShowBottom
	EntityDataFlagStanding
	EntityDataFlagShaking
	EntityDataFlagIdling
	EntityDataFlagCasting
	EntityDataFlagCharging
	EntityDataFlagKeyboardControlled
	EntityDataFlagPowerJump
	EntityDataFlagDash
	EntityDataFlagLingering
	EntityDataFlagHasCollision
	EntityDataFlagHasGravity
	EntityDataFlagFireImmune
	EntityDataFlagDancing
	EntityDataFlagEnchanted
	EntityDataFlagReturnTrident
	EntityDataFlagContainerPrivate
	EntityDataFlagTransforming
	EntityDataFlagDamageNearbyMobs
	EntityDataFlagSwimming
	EntityDataFlagBribed
	EntityDataFlagPregnant
	EntityDataFlagLayingEgg
	EntityDataFlagPassengerCanPick
	EntityDataFlagTransitionSitting
	EntityDataFlagEating
	EntityDataFlagLayingDown
	EntityDataFlagSneezing
	EntityDataFlagTrusting
	EntityDataFlagRolling
	EntityDataFlagScared
	EntityDataFlagInScaffolding
	EntityDataFlagOverScaffolding
	EntityDataFlagDescendThroughBlock
	EntityDataFlagBlocking
	EntityDataFlagTransitionBlocking
	EntityDataFlagBlockedUsingShield
	EntityDataFlagBlockedUsingDamagedShield
	EntityDataFlagSleeping
	EntityDataFlagWantsToWake
	EntityDataFlagTradeInterest
	EntityDataFlagDoorBreaker
	EntityDataFlagBreakingObstruction
	EntityDataFlagDoorOpener
	EntityDataFlagCaptain
	EntityDataFlagStunned
	EntityDataFlagRoaring
	EntityDataFlagDelayedAttack
	EntityDataFlagAvoidingMobs
	EntityDataFlagAvoidingBlock
	EntityDataFlagFacingTargetToRangeAttack
	EntityDataFlagHiddenWhenInvisible
	EntityDataFlagInUI
	EntityDataFlagStalking
	EntityDataFlagEmoting
	EntityDataFlagCelebrating
	EntityDataFlagAdmiring
	EntityDataFlagCelebratingSpecial
	EntityDataFlagOutOfControl
	EntityDataFlagRamAttack
	EntityDataFlagPlayingDead
	EntityDataFlagInAscendingBlock
	EntityDataFlagOverDescendingBlock
	EntityDataFlagCroaking
	EntityDataFlagDigestMob
	EntityDataFlagJumpGoal
	EntityDataFlagEmerging
	EntityDataFlagSniffing
	EntityDataFlagDigging
	EntityDataFlagSonicBoom
	EntityDataFlagHasDashTimeout
	EntityDataFlagPushTowardsClosestSpace
	EntityDataFlagScenting
	EntityDataFlagRising
	EntityDataFlagFeelingHappy
	EntityDataFlagSearching
	EntityDataFlagCrawling
	EntityDataFlagTimerFlag1
	EntityDataFlagTimerFlag2
	EntityDataFlagTimerFlag3
)

const (
	EntityDataTypeByte uint32 = iota
	EntityDataTypeInt16
	EntityDataTypeInt32
	EntityDataTypeFloat32
	EntityDataTypeString
	EntityDataTypeCompoundTag
	EntityDataTypeBlockPos
	EntityDataTypeInt64
	EntityDataTypeVec3
)

// EntityMetadata represents a map that holds metadata associated with an entity. The data held in the map depends on
// the entity and varies on a per-entity basis.
type EntityMetadata map[uint32]any

// NewEntityMetadata initializes and returns a new entity metadata map.
func NewEntityMetadata() EntityMetadata {
	return map[uint32]any{
		EntityDataKeyFlags:       int64(0),
		EntityDataKeyFlagsTwo:    int64(0),
		EntityDataKeyPlayerFlags: byte(0),
	}
}

// SetFlag sets a flag with a given index and value within the entity metadata map.
func (m EntityMetadata) SetFlag(key uint32, index uint8) {
	v := m[key]
	switch key {
	case EntityDataKeyPlayerFlags:
		m[key] = v.(byte) ^ (1 << index)
	default:
		m[key] = v.(int64) ^ (1 << int64(index))
	}
}

// Flag returns true if the flag with the index passed is set within the entity metadata.
func (m EntityMetadata) Flag(key uint32, index uint8) bool {
	v := m[key]
	switch key {
	case EntityDataKeyPlayerFlags:
		return v.(byte)&(1<<index) != 0
	default:
		return v.(int64)&(1<<int64(index)) != 0
	}
}
