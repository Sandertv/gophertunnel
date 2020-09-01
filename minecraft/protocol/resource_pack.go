package protocol

// ResourcePackInfo represents a resource pack's info sent over network. It holds information about the
// resource pack such as its name, description and version.
type ResourcePackInfo struct {
	// UUID is the UUID of the resource pack. Each resource pack downloaded must have a different UUID in
	// order for the client to be able to handle them properly.
	UUID string
	// Version is the version of the resource pack. The client will cache resource packs sent by the server as
	// long as they carry the same version. Sending a resource pack with a different version than previously
	// will force the client to re-download it.
	Version string
	// Size is the total size in bytes that the resource pack occupies. This is the size of the compressed
	// archive (zip) of the resource pack.
	Size uint64
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

// WritePackInfo writes a ResourcePackInfo x to Writer w.
func WritePackInfo(w *Writer, x *ResourcePackInfo) {
	w.String(&x.UUID)
	w.String(&x.Version)
	w.Uint64(&x.Size)
	w.String(&x.ContentKey)
	w.String(&x.SubPackName)
	w.String(&x.ContentIdentity)
	w.Bool(&x.HasScripts)
}

// PackInfo reads a ResourcePackInfo x from Reader r.
func PackInfo(r *Reader, x *ResourcePackInfo) {
	r.String(&x.UUID)
	r.String(&x.Version)
	r.Uint64(&x.Size)
	r.String(&x.ContentKey)
	r.String(&x.SubPackName)
	r.String(&x.ContentIdentity)
	r.Bool(&x.HasScripts)
}

// WriteStackPack writes a StackResourcePack x to Writer w.
func WriteStackPack(w *Writer, x *StackResourcePack) {
	w.String(&x.UUID)
	w.String(&x.Version)
	w.String(&x.SubPackName)
}

// StackPack reads a StackResourcePack x from Reader r.
func StackPack(r *Reader, x *StackResourcePack) {
	r.String(&x.UUID)
	r.String(&x.Version)
	r.String(&x.SubPackName)
}
