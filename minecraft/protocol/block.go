package protocol

import "github.com/sandertv/gophertunnel/minecraft/nbt"

// BlockEntry is an entry for a custom block found in the StartGame packet. The order of these specify the
// runtime ID that the blocks get.
type BlockEntry struct {
	// Name is the name of the custom block.
	Name string
	// Properties is a list of properties which, in combination with the name, specify a unique block.
	Properties map[string]interface{}
}

// Block reads a BlockEntry x from IO r.
func Block(r IO, x *BlockEntry) {
	r.String(&x.Name)
	r.NBT(&x.Properties, nbt.NetworkLittleEndian)
}
