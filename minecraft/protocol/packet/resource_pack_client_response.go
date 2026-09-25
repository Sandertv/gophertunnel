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
	Response uint32
	// PacksToDownload is a list of resource pack UUIDs combined with their version that need to be downloaded
	// (for example SomePack_1.0.0), if the Response field is PackResponseSendPacks.
	PacksToDownload []string
}

// ID ...
func (*ResourcePackClientResponse) ID() uint32 {
	return IDResourcePackClientResponse
}

func (pk *ResourcePackClientResponse) Marshal(io protocol.IO) {
	io.Varuint32(&pk.Response)
	name, ok := resourcePackResponseToString(pk.Response)
	if !ok {
		io.UnknownEnumOption(pk.Response, "resource pack response")
		return
	}
	io.String(&name)

	if pk.Response == PackResponseSendPacks {
		protocol.FuncSlice(io, &pk.PacksToDownload, io.String)
	}
}

func resourcePackResponseToString(x uint32) (string, bool) {
	names := [...]string{"cancel", "downloading", "downloadingfinished", "resourcepackstackfinished"}
	if x >= uint32(len(names)) {
		return "", false
	}
	return names[x], true
}
