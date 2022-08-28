package protocol

// BehaviourPackInfo represents a behaviour pack's info sent over network. It holds information about the
// behaviour pack such as its name, description and version.
type BehaviourPackInfo struct {
	// UUID is the UUID of the behaviour pack. Each behaviour pack downloaded must have a different UUID in
	// order for the client to be able to handle them properly.
	UUID string
	// Version is the version of the behaviour pack. The client will cache behaviour packs sent by the server as
	// long as they carry the same version. Sending a behaviour pack with a different version than previously
	// will force the client to re-download it.
	Version string
	// Size is the total size in bytes that the behaviour pack occupies. This is the size of the compressed
	// archive (zip) of the behaviour pack.
	Size uint64
	// ContentKey is the key used to decrypt the behaviour pack if it is encrypted. This is generally the case
	// for marketplace behaviour packs.
	ContentKey string
	// SubPackName ...
	SubPackName string
	// ContentIdentity ...
	ContentIdentity string
	// HasScripts specifies if the behaviour packs has any scripts in it. A client will only download the
	// behaviour pack if it supports scripts, which, up to 1.11, only includes Windows 10.
	HasScripts bool
}

// Marshal encodes/decodes a BehaviourPackInfo.
func (x *BehaviourPackInfo) Marshal(r IO) {
	r.String(&x.UUID)
	r.String(&x.Version)
	r.Uint64(&x.Size)
	r.String(&x.ContentKey)
	r.String(&x.SubPackName)
	r.String(&x.ContentIdentity)
	r.Bool(&x.HasScripts)
}

// TexturePackInfo represents a texture pack's info sent over network. It holds information about the
// texture pack such as its name, description and version.
type TexturePackInfo struct {
	// UUID is the UUID of the texture pack. Each texture pack downloaded must have a different UUID in
	// order for the client to be able to handle them properly.
	UUID string
	// Version is the version of the texture pack. The client will cache texture packs sent by the server as
	// long as they carry the same version. Sending a texture pack with a different version than previously
	// will force the client to re-download it.
	Version string
	// Size is the total size in bytes that the texture pack occupies. This is the size of the compressed
	// archive (zip) of the texture pack.
	Size uint64
	// ContentKey is the key used to decrypt the behaviour pack if it is encrypted. This is generally the case
	// for marketplace texture packs.
	ContentKey string
	// SubPackName ...
	SubPackName string
	// ContentIdentity is another UUID for the resource pack, and is generally set for marketplace texture
	// packs. It is also required for client-side validations when the resource pack is encrypted.
	ContentIdentity string
	// HasScripts specifies if the texture packs has any scripts in it. A client will only download the
	// behaviour pack if it supports scripts, which, up to 1.11, only includes Windows 10.
	HasScripts bool
	// RTXEnabled specifies if the texture pack uses the raytracing technology introduced in 1.16.200.
	RTXEnabled bool
}

// Marshal encodes/decodes a TexturePackInfo.
func (x *TexturePackInfo) Marshal(r IO) {
	r.String(&x.UUID)
	r.String(&x.Version)
	r.Uint64(&x.Size)
	r.String(&x.ContentKey)
	r.String(&x.SubPackName)
	r.String(&x.ContentIdentity)
	r.Bool(&x.HasScripts)
	r.Bool(&x.RTXEnabled)
}

// StackResourcePack represents a resource pack sent on the stack of the client. When sent, the client will
// apply them in the order of the stack sent.
type StackResourcePack struct {
	// UUID is the UUID of the resource pack. Each resource pack downloaded must have a different UUID in
	// order for the client to be able to handle them properly.
	UUID string
	// Version is the version of the resource pack. The client will cache resource packs sent by the server as
	// long as they carry the same version. Sending a resource pack with a different version than previously
	// will force the client to re-download it.
	Version string
	// SubPackName ...
	SubPackName string
}

// Marshal encodes/decodes a StackResourcePack.
func (x *StackResourcePack) Marshal(r IO) {
	r.String(&x.UUID)
	r.String(&x.Version)
	r.String(&x.SubPackName)
}
