package packet

import (
	"bytes"
	"encoding/binary"
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
func (pk *ResourcePackClientResponse) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.Response)
	_ = binary.Write(buf, binary.LittleEndian, uint16(len(pk.PacksToDownload)))
	for _, pack := range pk.PacksToDownload {
		_ = protocol.WriteString(buf, pack)
	}
}

// Unmarshal ...
func (pk *ResourcePackClientResponse) Unmarshal(buf *bytes.Buffer) error {
	var length uint16
	if err := chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.Response),
		binary.Read(buf, binary.LittleEndian, &length),
	); err != nil {
		return err
	}
	for i := uint16(0); i < length; i++ {
		var pack string
		if err := protocol.String(buf, &pack); err != nil {
			return err
		}
		pk.PacksToDownload = append(pk.PacksToDownload, pack)
	}
	return nil
}
