package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PackResponseRefused = iota + 1
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

// Marshal ...
func (pk *ResourcePackClientResponse) Marshal(w *protocol.Writer) {
	w.Uint8(&pk.Response)
	l := uint16(len(pk.PacksToDownload))
	w.Uint16(&l)
	for _, pack := range pk.PacksToDownload {
		w.String(&pack)
	}
}

// Unmarshal ...
func (pk *ResourcePackClientResponse) Unmarshal(r *protocol.Reader) {
	var length uint16
	r.Uint8(&pk.Response)
	r.Uint16(&length)

	pk.PacksToDownload = make([]string, length)
	for i := uint16(0); i < length; i++ {
		r.String(&pk.PacksToDownload[i])
	}
}
