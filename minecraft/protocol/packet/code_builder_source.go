package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	CodeBuilderCategoryNone = iota
	CodeBuilderCategoryStatus
	CodeBuilderCategoryInstantiation
)

const (
	CodeBuilderOperationNone = iota
	CodeBuilderOperationGet
	CodeBuilderOperationSet
	CodeBuilderOperationReset
)

// CodeBuilderSource is an Education Edition packet sent by the client to the server to run an operation with a
// code builder.
type CodeBuilderSource struct {
	// Operation is used to distinguish the operation performed. It is always one of the constants listed above.
	Operation byte
	// Category is used to distinguish the category of the operation performed. It is always one of the constants
	// listed above.
	Category byte
	// Value contains extra data about the operation performed. It is always empty unless the operation is
	// CodeBuilderOperationSet.
	Value []byte
}

// ID ...
func (pk *CodeBuilderSource) ID() uint32 {
	return IDCodeBuilderSource
}

func (pk *CodeBuilderSource) Marshal(io protocol.IO) {
	io.Uint8(&pk.Operation)
	io.Uint8(&pk.Category)
	io.ByteSlice(&pk.Value)
}
