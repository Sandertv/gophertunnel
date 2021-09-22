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
	// AgentCapabilities ...
	AgentCapabilities *bool
	// OverrideURI ...
	OverrideURI string
	// HasQuiz specifies if the world has a quiz connected to it.
	HasQuiz bool
	// ExternalLinkSettings ...
	ExternalLinkSettings *bool
}

// ID ...
func (*EducationSettings) ID() uint32 {
	return IDEducationSettings
}

// Marshal ...
func (pk *EducationSettings) Marshal(w *protocol.Writer) {
	hasOverrideURI := pk.OverrideURI != ""
	w.String(&pk.CodeBuilderDefaultURI)
	w.String(&pk.CodeBuilderTitle)
	w.Bool(&pk.CanResizeCodeBuilder)
	w.Bool(&pk.DisableLegacyTitleBar)
	w.String(&pk.PostProcessFilter)
	w.String(&pk.ScreenshotBorderPath)

	if pk.AgentCapabilities == nil {
		hasAgentCapabilities := false
		w.Bool(&hasAgentCapabilities)
	} else {
		hasAgentCapabilities := true
		w.Bool(&hasAgentCapabilities)
		w.Bool(pk.AgentCapabilities)
	}

	w.Bool(&hasOverrideURI)
	if hasOverrideURI {
		w.String(&pk.OverrideURI)
	}
	w.Bool(&pk.HasQuiz)

	if pk.ExternalLinkSettings == nil {
		hasExternalLinkSettings := false
		w.Bool(&hasExternalLinkSettings)
	} else {
		hasExternalLinkSettings := true
		w.Bool(&hasExternalLinkSettings)
		w.Bool(pk.ExternalLinkSettings)
	}
}

// Unmarshal ...
func (pk *EducationSettings) Unmarshal(r *protocol.Reader) {
	var hasOverrideURI bool
	r.String(&pk.CodeBuilderDefaultURI)
	r.String(&pk.CodeBuilderTitle)
	r.Bool(&pk.CanResizeCodeBuilder)
	r.Bool(&pk.DisableLegacyTitleBar)
	r.String(&pk.PostProcessFilter)
	r.String(&pk.ScreenshotBorderPath)

	var hasAgentCapabilities bool
	r.Bool(&hasAgentCapabilities)
	if hasAgentCapabilities {
		r.Bool(pk.AgentCapabilities)
	}

	r.Bool(&hasOverrideURI)
	if hasOverrideURI {
		r.String(&pk.OverrideURI)
	}
	r.Bool(&pk.HasQuiz)

	var hasExternalLinkSettings bool
	r.Bool(&hasExternalLinkSettings)
	if hasExternalLinkSettings {
		r.Bool(pk.ExternalLinkSettings)
	}
}
