package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PackResponseRefused = iota
	PackResponseSendPacks
	PackResponseAllPacksDownloaded
	PackResponseCompleted
)

// ResourcePackClientResponse is sent by the client in response to resource packets sent by the server. It is
// used to let the server know what action needs to be taken for the client to have all resource packs ready
// and set.
type ResourcePackClientResponse struct {
	// Response is the response type of the response. It is one of the constants found above.
	Response byte
	// PacksToDownload is a list of resource pack UUIDs combined with their version that need to be downloaded
	// (for example SomePack_1.0.0), if the Response field is PackResponseSendPacks.
	PacksToDownload []string
}

// ID ...
func (*ResourcePackClientResponse) ID() uint32 {
	return IDResourcePackClientResponse
}

func (pk *ResourcePackClientResponse) Marshal(io protocol.IO) {
	response := uint32(pk.Response)
	io.Varuint32(&response)
	if response > uint32(PackResponseCompleted) {
		io.UnknownEnumOption(response, "resource pack response")
	}
	pk.Response = byte(response)

	// A varuint32 length-prefixed string follows the response. Its contents are ignored by the client,
	// so we write the readable name of the response for parity and discard whatever is read.
	name := resourcePackResponseToString(pk.Response)
	io.String(&name)

	if pk.Response == PackResponseSendPacks {
		protocol.FuncSlice(io, &pk.PacksToDownload, io.String)
	}
}

func resourcePackResponseToString(x uint8) string {
	switch x {
	case PackResponseRefused:
		return "cancel"
	case PackResponseSendPacks:
		return "downloading"
	case PackResponseAllPacksDownloaded:
		return "downloadingfinished"
	case PackResponseCompleted:
		return "resourcepackstackfinished"
	default:
		return "unknown"
	}
}
