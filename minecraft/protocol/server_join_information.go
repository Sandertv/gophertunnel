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
	// ExperienceName is the optional name of the experience.
	ExperienceName Optional[string]
	// WorldName is the optional name of the world.
	WorldName Optional[string]
	// RichPresenceID is the rich presence ID overriding the client-driven
	// rich presence.
	RichPresenceID string
}

// Marshal encodes/decodes a PresenceInfo.
func (x *PresenceInfo) Marshal(r IO) {
	OptionalFunc(r, &x.ExperienceName, r.String)
	OptionalFunc(r, &x.WorldName, r.String)
	r.String(&x.RichPresenceID)
}

// ServerJoinInformation contains optional information about the server the player is joining.
type ServerJoinInformation struct {
	// GatheringJoinInfo is optional information about the gathering being joined.
	GatheringJoinInfo Optional[GatheringJoinInfo]
	// StoreEntryPointInfo is optional information about the store entry point.
	StoreEntryPointInfo Optional[StoreEntryPointInfo]
	// PresenceInfo is optional presence information.
	PresenceInfo Optional[PresenceInfo]
}

// Marshal encodes/decodes a ServerJoinInformation.
func (x *ServerJoinInformation) Marshal(r IO) {
	OptionalMarshaler(r, &x.GatheringJoinInfo)
	OptionalMarshaler(r, &x.StoreEntryPointInfo)
	OptionalMarshaler(r, &x.PresenceInfo)
}
