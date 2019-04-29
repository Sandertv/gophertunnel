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
)

// ...
const (
	IDResourcePackDataInfo = iota + 0x52
	IDResourcePackChunkData
)
