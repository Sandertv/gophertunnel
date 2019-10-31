package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// EducationSettings is a packet sent by the server to update Minecraft: Education Edition related settings.
// It is unused by the normal base game.
type EducationSettings struct {
	// CodeBuilderDefaultURI is the default URI that the code builder is ran on. Using this, a Code Builder
	// program can make code directly affect the server.
	CodeBuilderDefaultURI string
	// HasQuiz specifies if the world has a quiz connected to it.
	HasQuiz bool
}

// ID ...
func (*EducationSettings) ID() uint32 {
	return IDEducationSettings
}

// Marshal ...
func (pk *EducationSettings) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.CodeBuilderDefaultURI)
	_ = binary.Write(buf, binary.LittleEndian, pk.HasQuiz)
}

// Unmarshal ...
func (pk *EducationSettings) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.CodeBuilderDefaultURI),
		binary.Read(buf, binary.LittleEndian, &pk.HasQuiz),
	)
}
