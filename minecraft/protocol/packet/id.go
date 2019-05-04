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
)

// ...
const (
	IDSetLocalPlayerAsInitialised = iota + 0x71
)
