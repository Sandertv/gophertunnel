package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ResourcePacksInfo is sent by the server to inform the client on what resource packs the server has. It
// sends a list of the resource packs it has and basic information on them like the version and description.
type ResourcePacksInfo struct {
	// MustAccept specifies if the client must accept the resource packs the server has in order to join the
	// server. If set to true, the client gets the option to either download the resource packs and join, or
	// quit entirely.
	MustAccept bool
	// HasScripts specifies if any of the resource packs contain scripts in them. If set to true, only clients
	// that support scripts will be able to download them.
	HasScripts bool
	// BehaviourPack is a list of behaviour packs that the client needs to download before joining the server.
	// All of these behaviour packs will be applied together.
	BehaviourPacks []ResourcePack
	// TexturePacks is a list of texture packs that the client needs to download before joining the server.
	// The order of these texture packs specifies which texture pack is applied first, with the first in the
	// list being the first to be applied.
	TexturePacks []ResourcePack
}

// ID ...
func (*ResourcePacksInfo) ID() uint32 {
	return IDResourcePacksInfo
}

// Marshal ...
func (pk *ResourcePacksInfo) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.MustAccept)
	_ = binary.Write(buf, binary.LittleEndian, pk.HasScripts)
	_ = binary.Write(buf, binary.LittleEndian, int16(len(pk.BehaviourPacks)))
	for _, pack := range pk.BehaviourPacks {
		pack.Marshal(buf)
	}
	_ = binary.Write(buf, binary.LittleEndian, int16(len(pk.TexturePacks)))
	for _, pack := range pk.TexturePacks {
		pack.Marshal(buf)
	}
}

// Unmarshal ...
func (pk *ResourcePacksInfo) Unmarshal(buf *bytes.Buffer) error {
	if err := binary.Read(buf, binary.LittleEndian, &pk.MustAccept); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &pk.HasScripts); err != nil {
		return err
	}
	var length int16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return err
	}
	for i := int16(0); i < length; i++ {
		pack := &ResourcePack{}
		if err := pack.Unmarshal(buf); err != nil {
			return err
		}
		pk.BehaviourPacks = append(pk.BehaviourPacks, *pack)
	}
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return err
	}
	for i := int16(0); i < length; i++ {
		pack := &ResourcePack{}
		if err := pack.Unmarshal(buf); err != nil {
			return err
		}
		pk.TexturePacks = append(pk.TexturePacks, *pack)
	}
	return nil
}

// ResourcePack represents a resource pack sent over network. It holds information about the resource pack
// such as its name, description and version.
type ResourcePack struct {
	// UUID is the UUID of the resource pack. Each resource pack downloaded must have a different UUID in
	// order for the client to be able to handle them properly.
	UUID string
	// Version is the version of the resource pack. The client will cache resource packs sent by the server as
	// long as they carry the same version. Sending a resource pack with a different version than previously
	// will force the client to re-download it.
	Version string
	// Size is the total size in bytes that the resource pack occupies. This is the size of the compressed
	// archive (zip) of the resource pack.
	Size int64
	// ContentKey is the key used to decrypt the resource pack if it is encrypted. This is generally the case
	// for marketplace resource packs.
	ContentKey string
	// SubPackName ...
	SubPackName string
	// ContentIdentity ...
	ContentIdentity string
	// HasScripts specifies if the resource packs has any scripts in it. A client will only download the
	// resource pack if it supports scripts, which, up to 1.11, only includes Windows 10.
	HasScripts bool
}

// Marshal ...
func (pack ResourcePack) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pack.UUID)
	_ = protocol.WriteString(buf, pack.Version)
	_ = binary.Write(buf, binary.LittleEndian, pack.Size)
	_ = protocol.WriteString(buf, pack.ContentKey)
	_ = protocol.WriteString(buf, pack.SubPackName)
	_ = protocol.WriteString(buf, pack.ContentIdentity)
	_ = binary.Write(buf, binary.LittleEndian, pack.HasScripts)
}

// Unmarshal ...
func (pack *ResourcePack) Unmarshal(buf *bytes.Buffer) error {
	if err := protocol.String(buf, &pack.UUID); err != nil {
		return err
	}
	if err := protocol.String(buf, &pack.Version); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &pack.Size); err != nil {
		return err
	}
	if err := protocol.String(buf, &pack.ContentKey); err != nil {
		return err
	}
	if err := protocol.String(buf, &pack.SubPackName); err != nil {
		return err
	}
	if err := protocol.String(buf, &pack.ContentIdentity); err != nil {
		return err
	}
	return binary.Read(buf, binary.LittleEndian, &pack.HasScripts)
}
