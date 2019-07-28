package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PlayerSkin is sent by the client to the server when it updates its own skin using the in-game skin picker.
// It is relayed by the server, or sent if the server changes the skin of a player on its own accord. Note
// that the packet can only be sent for players that are in the player list at the time of sending.
type PlayerSkin struct {
	// UUID is the UUID of the player as sent in the Login packet when the client joined the server. It must
	// match this UUID exactly for the skin to show up on the player.
	UUID uuid.UUID
	// SkinID is a unique ID produced for the skin, for example 'c18e65aa-7b21-4637-9b63-8ad63622ef01_Alex'
	// for the default Alex skin.
	SkinID string
	// NewSkinName no longer has a function: The field can be left empty at all times.
	NewSkinName string
	// OldSkinName no longer has a function: The field can be left empty at all times.
	OldSkinName string
	// SkinData is a byte slice of 64*32*4, 64*64*4 or 128*128*4 bytes. It is a RGBA ordered byte
	// representation of the skin colours.
	SkinData []byte
	// CapeData is a byte slice of 64*32*4 bytes. It is a RGBA ordered byte representation of the cape
	// colours, much like the SkinData.
	CapeData []byte
	// SkinGeometryName is the geometry name of the skin geometry above. This name must be equal to one of the
	// outer names found in the SkinGeometry, so that the client can find the correct geometry data.
	SkinGeometryName string
	// SkinGeometry is a base64 JSON encoded structure of the geometry data of a skin, containing properties
	// such as bones, uv, pivot etc.
	SkinGeometry []byte
	// PremiumSkin specifies if the skin equipped was a premium skin, meaning a payment was required in the
	// marketplace to get access to it.
	PremiumSkin bool
}

// ID ...
func (*PlayerSkin) ID() uint32 {
	return IDPlayerSkin
}

// Marshal ...
func (pk *PlayerSkin) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteUUID(buf, pk.UUID)
	_ = protocol.WriteString(buf, pk.SkinID)
	_ = protocol.WriteString(buf, pk.NewSkinName)
	_ = protocol.WriteString(buf, pk.OldSkinName)
	_ = protocol.WriteByteSlice(buf, pk.SkinData)
	_ = protocol.WriteByteSlice(buf, pk.CapeData)
	_ = protocol.WriteString(buf, pk.SkinGeometryName)
	_ = protocol.WriteByteSlice(buf, pk.SkinGeometry)
	_ = binary.Write(buf, binary.LittleEndian, pk.PremiumSkin)
}

// Unmarshal ...
func (pk *PlayerSkin) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.UUID(buf, &pk.UUID),
		protocol.String(buf, &pk.SkinID),
		protocol.String(buf, &pk.NewSkinName),
		protocol.String(buf, &pk.OldSkinName),
		protocol.ByteSlice(buf, &pk.SkinData),
		protocol.ByteSlice(buf, &pk.CapeData),
		protocol.String(buf, &pk.SkinGeometryName),
		protocol.ByteSlice(buf, &pk.SkinGeometry),
		binary.Read(buf, binary.LittleEndian, &pk.PremiumSkin),
	)
}
