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
	// CodeBuilderTitle is the title of the code builder shown when connected to the CodeBuilderDefaultURI.
	CodeBuilderTitle string
	// CanResizeCodeBuilder specifies if clients connected to the world should be able to resize the code
	// builder when it is opened.
	CanResizeCodeBuilder bool
	// OverrideURI ...
	OverrideURI string
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
	_ = protocol.WriteString(buf, pk.CodeBuilderTitle)
	_ = binary.Write(buf, binary.LittleEndian, pk.CanResizeCodeBuilder)
	_ = binary.Write(buf, binary.LittleEndian, pk.OverrideURI != "")
	if pk.OverrideURI != "" {
		_ = protocol.WriteString(buf, pk.OverrideURI)
	}
	_ = binary.Write(buf, binary.LittleEndian, pk.HasQuiz)
}

// Unmarshal ...
func (pk *EducationSettings) Unmarshal(r *protocol.Reader) {
	var hasOverrideURI bool
	r.String(&pk.CodeBuilderDefaultURI)
	r.String(&pk.CodeBuilderTitle)
	r.Bool(&pk.CanResizeCodeBuilder)
	r.Bool(&hasOverrideURI)
	if hasOverrideURI {
		r.String(&pk.OverrideURI)
	}
	r.Bool(&pk.HasQuiz)
}
