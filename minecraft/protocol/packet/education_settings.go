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

// Marshal ...
func (pk *EducationSettings) Marshal(w *protocol.Writer) {
	w.String(&pk.CodeBuilderDefaultURI)
	w.String(&pk.CodeBuilderTitle)
	w.Bool(&pk.CanResizeCodeBuilder)
	w.Bool(&pk.DisableLegacyTitleBar)
	w.String(&pk.PostProcessFilter)
	w.String(&pk.ScreenshotBorderPath)

	protocol.OptionalFunc(w, &pk.CanModifyBlocks, w.Bool)
	protocol.OptionalFunc(w, &pk.OverrideURI, w.String)

	w.Bool(&pk.HasQuiz)

	protocol.OptionalMarshaler(w, &pk.ExternalLinkSettings)
}

// Unmarshal ...
func (pk *EducationSettings) Unmarshal(r *protocol.Reader) {
	r.String(&pk.CodeBuilderDefaultURI)
	r.String(&pk.CodeBuilderTitle)
	r.Bool(&pk.CanResizeCodeBuilder)
	r.Bool(&pk.DisableLegacyTitleBar)
	r.String(&pk.PostProcessFilter)
	r.String(&pk.ScreenshotBorderPath)

	protocol.OptionalFunc(r, &pk.CanModifyBlocks, r.Bool)
	protocol.OptionalFunc(r, &pk.OverrideURI, r.String)

	r.Bool(&pk.HasQuiz)

	protocol.OptionalMarshaler(r, &pk.ExternalLinkSettings)
}
