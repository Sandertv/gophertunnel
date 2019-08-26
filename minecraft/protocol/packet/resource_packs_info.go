package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ResourcePacksInfo is sent by the server to inform the client on what resource packs the server has. It
// sends a list of the resource packs it has and basic information on them like the version and description.
type ResourcePacksInfo struct {
	// TexturePackRequired specifies if the client must accept the texture packs the server has in order to
	// join the server. If set to true, the client gets the option to either download the resource packs and
	// join, or quit entirely. Behaviour packs never have to be downloaded.
	TexturePackRequired bool
	// HasScripts specifies if any of the resource packs contain scripts in them. If set to true, only clients
	// that support scripts will be able to download them.
	HasScripts bool
	// BehaviourPack is a list of behaviour packs that the client needs to download before joining the server.
	// All of these behaviour packs will be applied together.
	BehaviourPacks []protocol.ResourcePackInfo
	// TexturePacks is a list of texture packs that the client needs to download before joining the server.
	// The order of these texture packs is not relevant in this packet. It is however important in the
	// ResourcePackStack packet.
	TexturePacks []protocol.ResourcePackInfo
}

// ID ...
func (*ResourcePacksInfo) ID() uint32 {
	return IDResourcePacksInfo
}

// Marshal ...
func (pk *ResourcePacksInfo) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.TexturePackRequired)
	_ = binary.Write(buf, binary.LittleEndian, pk.HasScripts)
	_ = binary.Write(buf, binary.LittleEndian, int16(len(pk.BehaviourPacks)))
	for _, pack := range pk.BehaviourPacks {
		_ = protocol.WritePackInfo(buf, pack)
	}
	_ = binary.Write(buf, binary.LittleEndian, int16(len(pk.TexturePacks)))
	for _, pack := range pk.TexturePacks {
		_ = protocol.WritePackInfo(buf, pack)
	}
}

// Unmarshal ...
func (pk *ResourcePacksInfo) Unmarshal(buf *bytes.Buffer) error {
	var length int16
	if err := chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.TexturePackRequired),
		binary.Read(buf, binary.LittleEndian, &pk.HasScripts),
		binary.Read(buf, binary.LittleEndian, &length),
	); err != nil {
		return err
	}
	pk.BehaviourPacks = make([]protocol.ResourcePackInfo, length)
	for i := int16(0); i < length; i++ {
		if err := protocol.PackInfo(buf, &pk.BehaviourPacks[i]); err != nil {
			return err
		}
	}
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return err
	}
	pk.TexturePacks = make([]protocol.ResourcePackInfo, length)
	for i := int16(0); i < length; i++ {
		if err := protocol.PackInfo(buf, &pk.TexturePacks[i]); err != nil {
			return err
		}
	}
	return nil
}
