package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ResourcePackStack is sent by the server to send the order in which resource packs and behaviour packs
// should be applied (and downloaded) by the client.
type ResourcePackStack struct {
	// TexturePackRequired specifies if the client must accept the texture packs the server has in order to
	// join the server. If set to true, the client gets the option to either download the resource packs and
	// join, or quit entirely. Behaviour packs never have to be downloaded.
	TexturePackRequired bool
	// BehaviourPack is a list of behaviour packs that the client needs to download before joining the server.
	// All of these behaviour packs will be applied together, and the order does not necessarily matter.
	BehaviourPacks []protocol.StackResourcePack
	// TexturePacks is a list of texture packs that the client needs to download before joining the server.
	// The order of these texture packs specifies the order that they are applied in on the client side. The
	// first in the list will be applied first.
	TexturePacks []protocol.StackResourcePack
	// Experimental specifies if the resource packs in the stack are experimental. This is internal and should
	// always be set to false.
	Experimental bool
	// BaseGameVersion is the vanilla version that the client should set its resource pack stack to.
	BaseGameVersion string
}

// ID ...
func (*ResourcePackStack) ID() uint32 {
	return IDResourcePackStack
}

// Marshal ...
func (pk *ResourcePackStack) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.TexturePackRequired)
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.BehaviourPacks)))
	for _, pack := range pk.BehaviourPacks {
		_ = protocol.WriteStackPack(buf, pack)
	}
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.TexturePacks)))
	for _, pack := range pk.TexturePacks {
		_ = protocol.WriteStackPack(buf, pack)
	}
	_ = binary.Write(buf, binary.LittleEndian, pk.Experimental)
	_ = protocol.WriteString(buf, pk.BaseGameVersion)
}

// Unmarshal ...
func (pk *ResourcePackStack) Unmarshal(buf *bytes.Buffer) error {
	var length uint32
	if err := chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.TexturePackRequired),
		protocol.Varuint32(buf, &length),
	); err != nil {
		return err
	}
	pk.BehaviourPacks = make([]protocol.StackResourcePack, length)
	for i := uint32(0); i < length; i++ {
		if err := protocol.StackPack(buf, &pk.BehaviourPacks[i]); err != nil {
			return err
		}
	}
	if err := protocol.Varuint32(buf, &length); err != nil {
		return err
	}
	pk.TexturePacks = make([]protocol.StackResourcePack, length)
	for i := uint32(0); i < length; i++ {
		if err := protocol.StackPack(buf, &pk.TexturePacks[i]); err != nil {
			return err
		}
	}
	return chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.Experimental),
		protocol.String(buf, &pk.BaseGameVersion),
	)
}
