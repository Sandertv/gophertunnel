package packet

import (
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
	BehaviourPacks []protocol.BehaviourPackInfo
	// TexturePacks is a list of texture packs that the client needs to download before joining the server.
	// The order of these texture packs is not relevant in this packet. It is however important in the
	// ResourcePackStack packet.
	TexturePacks []protocol.TexturePackInfo
	// ForcingServerPacks is currently an unclear field.
	ForcingServerPacks bool
}

// ID ...
func (*ResourcePacksInfo) ID() uint32 {
	return IDResourcePacksInfo
}

// Marshal ...
func (pk *ResourcePacksInfo) Marshal(w *protocol.Writer) {
	w.Bool(&pk.TexturePackRequired)
	w.Bool(&pk.HasScripts)
	w.Bool(&pk.ForcingServerPacks)
	protocol.SliceUint16Length(w, &pk.BehaviourPacks)
	protocol.SliceUint16Length(w, &pk.TexturePacks)
}

// Unmarshal ...
func (pk *ResourcePacksInfo) Unmarshal(r *protocol.Reader) {
	r.Bool(&pk.TexturePackRequired)
	r.Bool(&pk.HasScripts)
	r.Bool(&pk.ForcingServerPacks)
	protocol.SliceUint16Length(r, &pk.BehaviourPacks)
	protocol.SliceUint16Length(r, &pk.TexturePacks)
}
