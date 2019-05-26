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
)

// ...
const (
	IDFullChunkData = iota + 0x3a
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
