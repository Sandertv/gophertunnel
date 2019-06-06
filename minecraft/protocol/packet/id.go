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
)

// ...
const (
	IDRequestChunkRadius = iota + 0x45
	IDChunkRadiusUpdated
)

// ...
const (
	IDResourcePackDataInfo = iota + 0x52
	IDResourcePackChunkData
	IDResourcePackChunkRequest
	IDTransfer
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
