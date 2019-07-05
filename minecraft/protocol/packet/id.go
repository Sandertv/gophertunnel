package packet

const (
	IDLogin = iota + 0x01
	IDPlayStatus
	IDServerToClientHandshake
	IDClientToServerHandshake
	IDDisconnect
	IDResourcePacksInfo
	IDResourcePackStack
	IDResourcePackClientResponse
	IDText
	IDSetTime
	IDStartGame
	IDAddPlayer
	IDAddEntity
	IDRemoveEntity
	IDAddItemEntity
	_
	IDTakeItemEntity
	IDMoveEntityAbsolute
	IDMovePlayer
	IDRiderJump
	IDUpdateBlock
	IDAddPainting
	IDExplode
	_
	IDLevelEvent
	IDBlockEvent
	IDEntityEvent
	IDMobEffect
	IDUpdateAttributes
	IDInventoryTransaction
	IDMobEquipment
	IDMobArmourEquipment
	IDInteract
	IDBlockPickRequest
	IDEntityPickRequest
	IDPlayerAction
	IDEntityFall
	IDHurtArmour
	IDSetEntityData
	IDSetEntityMotion
	IDSetEntityLink
	IDSetHealth
	IDSetSpawnPosition
	IDAnimate
	IDRespawn
	IDContainerOpen
	IDContainerClose
	IDPlayerHotBar
	IDInventoryContent
	IDInventorySlot
	IDContainerSetData
	IDCraftingData
	IDCraftingEvent
	IDGUIDataPickItem
	IDAdventureSettings
	IDBlockEntityData
	IDPlayerInput
	IDFullChunkData
	IDSetCommandsEnabled
	IDSetDifficulty
	IDChangeDimension
	IDSetPlayerGameType
	IDPlayerList
	IDSimpleEvent
	IDEvent
	IDSpawnExperienceOrb
	IDClientBoundMapItemData
	IDMapInfoRequest
	IDRequestChunkRadius
	IDChunkRadiusUpdated
	IDItemFrameDropItem
	IDGameRulesChanged
	IDCamera
	IDBossEvent
	IDShowCredits
	IDAvailableCommands
	IDCommandRequest
	IDCommandBlockUpdate
	IDCommandOutput
	IDUpdateTrade
	IDUpdateEquip
	IDResourcePackDataInfo
	IDResourcePackChunkData
	IDResourcePackChunkRequest
	IDTransfer
	IDPlaySound
)

// ...
const (
	IDPlayerSkin = iota + 0x5d
)

// ...
const (
	IDModalFormRequest = iota + 0x64
	IDModalFormResponse
	IDServerSettingsRequest
	IDServerSettingsResponse
)

// ...
const (
	IDSetLocalPlayerAsInitialised = iota + 0x71
)
