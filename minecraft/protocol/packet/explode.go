package packet

import (
	"bytes"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Explode is sent by the server to make an explosion happen. The explosion will break all blocks in the
// packet client-side.
type Explode struct {
	// Position is the centre position of the explosion that is taking place. It has no effect on the blocks
	// broken client-side, and seems to have no other functionality.
	Position mgl32.Vec3
	// Radius is the radius of the explosion. It seems to have no functionality.
	Radius float32
	// BlocksBroken is a list of all block positions that are to be destroyed by the explosion. This is the
	// only field that has functionality in the packet.
	BlocksBroken []protocol.BlockPos
}

// ID ...
func (*Explode) ID() uint32 {
	return IDExplode
}

// Marshal ...
func (pk *Explode) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVec3(buf, pk.Position)
	_ = protocol.WriteVarint32(buf, int32(pk.Radius*32))
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.BlocksBroken)))
	for _, blockBroken := range pk.BlocksBroken {
		_ = protocol.WriteBlockPosition(buf, blockBroken)
	}
}

// Unmarshal ...
func (pk *Explode) Unmarshal(buf *bytes.Buffer) error {
	var blockCount uint32
	var radiusInt32 int32
	if err := ChainErr(
		protocol.Vec3(buf, &pk.Position),
		protocol.Varint32(buf, &radiusInt32),
		protocol.Varuint32(buf, &blockCount),
	); err != nil {
		return err
	}
	pk.Radius = float32(radiusInt32) / 32.0
	pk.BlocksBroken = make([]protocol.BlockPos, blockCount)
	for i := uint32(0); i < blockCount; i++ {
		if err := protocol.BlockPosition(buf, &pk.BlocksBroken[i]); err != nil {
			return err
		}
	}
	return nil
}
