package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// EducationSettings is a packet sent by the server to update Minecraft: Education Edition related settings.
// It is unused by the normal base game.
type EducationSettings struct {
	// CodeBuilderDefaultURI is the default URI that the code builder is ran on. Using this, a Code Builder program can
	// make code directly affect the server.
	CodeBuilderDefaultURI string
	// CodeBuilderTitle is the title of the code builder shown when connected to the CodeBuilderDefaultURI.
	CodeBuilderTitle string
	// CanResizeCodeBuilder specifies if clients connected to the world should be able to resize the code
	// builder when it is opened.
	CanResizeCodeBuilder bool
	// DisableLegacyTitleBar ...
	DisableLegacyTitleBar bool
	// PostProcessFilter ...
	PostProcessFilter string
	// ScreenshotBorderPath ...
	ScreenshotBorderPath string
	// CanModifyBlocks ...
	CanModifyBlocks protocol.Optional[bool]
	// OverrideURI ...
	OverrideURI protocol.Optional[string]
	// HasQuiz specifies if the world has a quiz connected to it.
	HasQuiz bool
	// ExternalLinkSettings ...
	ExternalLinkSettings protocol.Optional[protocol.EducationExternalLinkSettings]
}

// ID ...
func (*EducationSettings) ID() uint32 {
	return IDEducationSettings
}

func (pk *EducationSettings) Marshal(io protocol.IO) {
	io.String(&pk.CodeBuilderDefaultURI)
	io.String(&pk.CodeBuilderTitle)
	io.Bool(&pk.CanResizeCodeBuilder)
	io.Bool(&pk.DisableLegacyTitleBar)
	io.String(&pk.PostProcessFilter)
	io.String(&pk.ScreenshotBorderPath)
	protocol.OptionalFunc(io, &pk.CanModifyBlocks, io.Bool)
	protocol.OptionalFunc(io, &pk.OverrideURI, io.String)
	io.Bool(&pk.HasQuiz)
	protocol.OptionalMarshaler(io, &pk.ExternalLinkSettings)
}
