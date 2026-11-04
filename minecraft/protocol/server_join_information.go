package protocol

import "github.com/google/uuid"

// GatheringJoinInfo contains information about the gathering (experience) the player is joining.
type GatheringJoinInfo struct {
	// ExperienceID is the UUID of the experience.
	ExperienceID uuid.UUID
	// ExperienceName is the name of the experience.
	ExperienceName string
	// ExperienceWorldID is the UUID of the experience world.
	ExperienceWorldID Optional[uuid.UUID]
	// ExperienceWorldName is the world name of the experience.
	ExperienceWorldName Optional[string]
	// CreatorID is the ID of the creator.
	CreatorID string
	// TargetID is the session ID of the experience.
	TargetID Optional[uuid.UUID]
	// ScenarioID is the scenario ID of experience.
	ScenarioID Optional[string]
	// ServerID is the server identifier.
	ServerID Optional[string]
}

// Marshal encodes/decodes a GatheringJoinInfo.
func (x *GatheringJoinInfo) Marshal(r IO) {
	r.UUID(&x.ExperienceID)
	r.String(&x.ExperienceName)
	OptionalFunc(r, &x.ExperienceWorldID, r.UUID)
	OptionalFunc(r, &x.ExperienceWorldName, r.String)
	r.String(&x.CreatorID)
	OptionalFunc(r, &x.TargetID, r.UUID)
	OptionalFunc(r, &x.ScenarioID, r.String)
	OptionalFunc(r, &x.ServerID, r.String)
}

// StoreEntryPointInfo contains information about the store entry point.
type StoreEntryPointInfo struct {
	// StoreID is the store identifier.
	StoreID string
	// StoreName is the store name.
	StoreName string
}

// Marshal encodes/decodes a StoreEntryPointInfo.
func (x *StoreEntryPointInfo) Marshal(r IO) {
	r.String(&x.StoreID)
	r.String(&x.StoreName)
}

// PresenceInfo contains presence information about the experience.
type PresenceInfo struct {
	// RichPresenceID is the rich presence ID overriding the client-driven
	// rich presence.
	RichPresenceID Optional[string]
}

// Marshal encodes/decodes a PresenceInfo.
func (x *PresenceInfo) Marshal(r IO) {
	OptionalFunc(r, &x.RichPresenceID, r.String)
}

// ManagedChatJoinInfo contains information about Mojang Managed Chat on the server.
type ManagedChatJoinInfo struct {
	// Enabled specifies if the server routes chat through Mojang Managed Chat.
	Enabled bool
}

// Marshal encodes/decodes a ManagedChatJoinInfo.
func (x *ManagedChatJoinInfo) Marshal(r IO) {
	r.Bool(&x.Enabled)
}

// ServerJoinInformation contains optional information about the server the player is joining.
type ServerJoinInformation struct {
	// GatheringJoinInfo is optional information about the gathering being joined.
	GatheringJoinInfo Optional[GatheringJoinInfo]
	// StoreEntryPointInfo is optional information about the store entry point.
	StoreEntryPointInfo Optional[StoreEntryPointInfo]
	// PresenceInfo is optional presence information.
	PresenceInfo Optional[PresenceInfo]
	// ManagedChatJoinInfo is optional information about Mojang Managed Chat.
	ManagedChatJoinInfo Optional[ManagedChatJoinInfo]
}

// Marshal encodes/decodes a ServerJoinInformation.
func (x *ServerJoinInformation) Marshal(r IO) {
	OptionalMarshaler(r, &x.GatheringJoinInfo)
	OptionalMarshaler(r, &x.StoreEntryPointInfo)
	OptionalMarshaler(r, &x.PresenceInfo)
	OptionalMarshaler(r, &x.ManagedChatJoinInfo)
}

// TransferUIOptions holds options for the loading screen shown when a player is transferred to another server.
type TransferUIOptions struct {
	// TitleText is the title shown on the loading screen.
	TitleText Optional[string]
	// BodyText is the body text shown on the loading screen.
	BodyText Optional[string]
	// TitleVisible specifies if the title is visible.
	TitleVisible Optional[bool]
	// BodyVisible specifies if the body text is visible.
	BodyVisible Optional[bool]
	// CancelButtonText is the text of the cancel button.
	CancelButtonText Optional[string]
	// CancelButtonVisible specifies if the cancel button is visible.
	CancelButtonVisible Optional[bool]
	// CancelButtonDelaySeconds is the delay, in seconds, before the cancel button is shown.
	CancelButtonDelaySeconds Optional[float32]
}

// Marshal encodes/decodes a TransferUIOptions.
func (x *TransferUIOptions) Marshal(r IO) {
	OptionalFunc(r, &x.TitleText, r.String)
	OptionalFunc(r, &x.BodyText, r.String)
	OptionalFunc(r, &x.TitleVisible, r.Bool)
	OptionalFunc(r, &x.BodyVisible, r.Bool)
	OptionalFunc(r, &x.CancelButtonText, r.String)
	OptionalFunc(r, &x.CancelButtonVisible, r.Bool)
	OptionalFunc(r, &x.CancelButtonDelaySeconds, r.Float32)
}
