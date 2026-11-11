package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SetPassengerOfBlock is sent by the server to make an entity start or stop riding a block.
type SetPassengerOfBlock struct {
	// PassengerUniqueID is the unique ID of the entity that starts or stops riding the block.
	PassengerUniqueID int64
	// Data holds the data of the block that the entity starts riding. If not set, the entity stops riding
	// the block.
	Data protocol.Optional[protocol.PassengerOfBlockData]
}

// ID ...
func (*SetPassengerOfBlock) ID() uint32 {
	return IDSetPassengerOfBlock
}

func (pk *SetPassengerOfBlock) Marshal(io protocol.IO) {
	io.ActorUniqueID(&pk.PassengerUniqueID)
	protocol.OptionalMarshaler(io, &pk.Data)
}
