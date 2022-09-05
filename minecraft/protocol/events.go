package protocol

// EventData represents an object that holds data specific to an event.
// The data it holds depends on the type.
type EventData interface {
	// Marshal encodes/decodes a serialised event data object.
	Marshal(r IO)
}

// AchievementAwardedEventData is the event data sent for achievements.
type AchievementAwardedEventData struct {
	// AchievementID is the ID for the achievement.
	AchievementID int32
}

// Marshal ...
func (a *AchievementAwardedEventData) Marshal(r IO) {
	r.Varint32(&a.AchievementID)
}

// EntityInteractEventData is the event data sent for entity interactions.
type EntityInteractEventData struct {
	// InteractionType ...
	InteractionType int32
	// InteractionEntityType ...
	InteractionEntityType int32
	// EntityVariant ...
	EntityVariant int32
	// EntityColour ...
	EntityColour uint8
}

// Marshal ...
func (e *EntityInteractEventData) Marshal(r IO) {
	r.Varint32(&e.InteractionType)
	r.Varint32(&e.InteractionEntityType)
	r.Varint32(&e.EntityVariant)
	r.Uint8(&e.EntityColour)
}

// PortalBuiltEventData is the event data sent when a portal is built.
type PortalBuiltEventData struct {
	// DimensionID ...
	DimensionID int32
}

// Marshal ...
func (p *PortalBuiltEventData) Marshal(r IO) {
	r.Varint32(&p.DimensionID)
}

// PortalUsedEventData is the event data sent when a portal is used.
type PortalUsedEventData struct {
	// FromDimensionID ...
	FromDimensionID int32
	// ToDimensionID ...
	ToDimensionID int32
}

// Marshal ...
func (p *PortalUsedEventData) Marshal(r IO) {
	r.Varint32(&p.FromDimensionID)
	r.Varint32(&p.ToDimensionID)
}

// MobKilledEventData is the event data sent when a mob is killed.
type MobKilledEventData struct {
	// KillerEntityUniqueID ...
	KillerEntityUniqueID int64
	// VictimEntityUniqueID ...
	VictimEntityUniqueID int64
	// KillerEntityType ...
	KillerEntityType int32
	// EntityDamageCause ...
	EntityDamageCause int32
	// VillagerTradeTier ...
	VillagerTradeTier int32
	// VillagerDisplayName ...
	VillagerDisplayName string
}

// Marshal ...
func (m *MobKilledEventData) Marshal(r IO) {
	r.Varint64(&m.KillerEntityUniqueID)
	r.Varint64(&m.VictimEntityUniqueID)
	r.Varint32(&m.KillerEntityType)
	r.Varint32(&m.EntityDamageCause)
	r.Varint32(&m.VillagerTradeTier)
	r.String(&m.VillagerDisplayName)
}

// CauldronUsedEventData is the event data sent when a cauldron is used.
type CauldronUsedEventData struct {
	// PotionID ...
	PotionID int32
	// Colour ...
	Colour int32
	// FillLevel ...
	FillLevel int32
}

// Marshal ...
func (c *CauldronUsedEventData) Marshal(r IO) {
	r.Varint32(&c.PotionID)
	r.Varint32(&c.Colour)
	r.Varint32(&c.FillLevel)
}

// PlayerDiedEventData is the event data sent when a player dies.
type PlayerDiedEventData struct {
	// AttackerEntityID ...
	AttackerEntityID int32
	// AttackerVariant ...
	AttackerVariant int32
	// EntityDamageCause ...
	EntityDamageCause int32
	// InRaid ...
	InRaid bool
}

// Marshal ...
func (p *PlayerDiedEventData) Marshal(r IO) {
	r.Varint32(&p.AttackerEntityID)
	r.Varint32(&p.AttackerVariant)
	r.Varint32(&p.EntityDamageCause)
	r.Bool(&p.InRaid)
}

// BossKilledEventData is the event data sent when a boss dies.
type BossKilledEventData struct {
	// BossEntityUniqueID ...
	BossEntityUniqueID int64
	// PlayerPartySize ...
	PlayerPartySize int32
	// InteractionEntityType ...
	InteractionEntityType int32
}

// Marshal ...
func (b *BossKilledEventData) Marshal(r IO) {
	r.Varint64(&b.BossEntityUniqueID)
	r.Varint32(&b.PlayerPartySize)
	r.Varint32(&b.InteractionEntityType)
}

// AgentCommandEventData is an event used in Education Edition.
type AgentCommandEventData struct {
	// AgentResult ...
	AgentResult int32
	// DataValue ...
	DataValue int32
	// Command ...
	Command string
	// DataKey ...
	DataKey string
	// Output ...
	Output string
}

// Marshal ...
func (a *AgentCommandEventData) Marshal(r IO) {
	r.Varint32(&a.AgentResult)
	r.Varint32(&a.DataValue)
	r.String(&a.Command)
	r.String(&a.DataKey)
	r.String(&a.Output)
}

// PatternRemovedEventData is the event data sent when a pattern is removed. This is now deprecated.
type PatternRemovedEventData struct {
	// ItemID ...
	ItemID int32
	// AuxValue ...
	AuxValue int32
	// PatternsSize ...
	PatternsSize int32
	// PatternIndex ...
	PatternIndex int32
	// PatternColour ...
	PatternColour int32
}

// Marshal ...
func (p *PatternRemovedEventData) Marshal(r IO) {
	r.Varint32(&p.ItemID)
	r.Varint32(&p.AuxValue)
	r.Varint32(&p.PatternsSize)
	r.Varint32(&p.PatternIndex)
	r.Varint32(&p.PatternColour)
}

// SlashCommandExecutedEventData is the event data sent when a slash command is executed.
type SlashCommandExecutedEventData struct {
	// CommandName ...
	CommandName string
	// SuccessCount ...
	SuccessCount int32
	// MessageCount indicates the amount of OutputMessages present.
	MessageCount int32
	// OutputMessages is a list of messages joint with ;.
	OutputMessages string
}

// Marshal ...
func (s *SlashCommandExecutedEventData) Marshal(r IO) {
	r.Varint32(&s.SuccessCount)
	r.Varint32(&s.MessageCount)
	r.String(&s.CommandName)
	r.String(&s.OutputMessages)
}

// FishBucketedEventData is the event data sent when a fish is bucketed.
type FishBucketedEventData struct {
	// Pattern ...
	Pattern int32
	// Preset ...
	Preset int32
	// BucketedEntityType ...
	BucketedEntityType int32
	// Release ...
	Release bool
}

// Marshal ...
func (f *FishBucketedEventData) Marshal(r IO) {
	r.Varint32(&f.Pattern)
	r.Varint32(&f.Preset)
	r.Varint32(&f.BucketedEntityType)
	r.Bool(&f.Release)
}

// MobBornEventData is the event data sent when a mob is born.
type MobBornEventData struct {
	// EntityType ...
	EntityType int32
	// Variant ...
	Variant int32
	// Colour ...
	Colour uint8
}

// Marshal ...
func (m *MobBornEventData) Marshal(r IO) {
	r.Varint32(&m.EntityType)
	r.Varint32(&m.Variant)
	r.Uint8(&m.Colour)
}

// PetDiedEventData is the event data sent when a pet dies. This is now deprecated.
type PetDiedEventData struct {
	// KilledByOwner ...
	KilledByOwner bool
	// KillerEntityUniqueID ...
	KillerEntityUniqueID int64
	// PetEntityUniqueID ...
	PetEntityUniqueID int64
	// EntityDamageCause ...
	EntityDamageCause int32
	// PetEntityType ...
	PetEntityType int32
}

// Marshal ...
func (p *PetDiedEventData) Marshal(r IO) {
	r.Bool(&p.KilledByOwner)
	r.Varint64(&p.KillerEntityUniqueID)
	r.Varint64(&p.PetEntityUniqueID)
	r.Varint32(&p.EntityDamageCause)
	r.Varint32(&p.PetEntityType)
}

// CauldronInteractEventData is the event data sent when a cauldron is interacted with.
type CauldronInteractEventData struct {
	// BlockInteractionType ...
	BlockInteractionType int32
	// ItemID ...
	ItemID int32
}

// Marshal ...
func (c *CauldronInteractEventData) Marshal(r IO) {
	r.Varint32(&c.BlockInteractionType)
	r.Varint32(&c.ItemID)
}

// ComposterInteractEventData is the event data sent when a composter is interacted with.
type ComposterInteractEventData struct {
	// BlockInteractionType ...
	BlockInteractionType int32
	// ItemID ...
	ItemID int32
}

// Marshal ...
func (c *ComposterInteractEventData) Marshal(r IO) {
	r.Varint32(&c.BlockInteractionType)
	r.Varint32(&c.ItemID)
}

// BellUsedEventData is the event data sent when a bell is used.
type BellUsedEventData struct {
	// ItemID ...
	ItemID int32
}

// Marshal ...
func (b *BellUsedEventData) Marshal(r IO) {
	r.Varint32(&b.ItemID)
}

// EntityDefinitionTriggerEventData is an event used for an unknown purpose.
type EntityDefinitionTriggerEventData struct {
	// EventName ...
	EventName string
}

// Marshal ...
func (e *EntityDefinitionTriggerEventData) Marshal(r IO) {
	r.String(&e.EventName)
}

// RaidUpdateEventData is an event used to update a raids progress client side.
type RaidUpdateEventData struct {
	// CurrentRaidWave ...
	CurrentRaidWave int32
	// TotalRaidWaves ...
	TotalRaidWaves int32
	// WonRaid ...
	WonRaid bool
}

// Marshal ...
func (ra *RaidUpdateEventData) Marshal(r IO) {
	r.Varint32(&ra.CurrentRaidWave)
	r.Varint32(&ra.TotalRaidWaves)
	r.Bool(&ra.WonRaid)
}

// MovementAnomalyEventData is an event used for updating the other party on movement data.
type MovementAnomalyEventData struct {
	// EventType ...
	EventType uint8
	// CheatingScore ...
	CheatingScore float32
	// AveragePositionDelta ...
	AveragePositionDelta float32
	// TotalPositionDelta ...
	TotalPositionDelta float32
	// MinPositionDelta ...
	MinPositionDelta float32
	// MaxPositionDelta ...
	MaxPositionDelta float32
}

// Marshal ...
func (m *MovementAnomalyEventData) Marshal(r IO) {
	r.Uint8(&m.EventType)
	r.Float32(&m.CheatingScore)
	r.Float32(&m.AveragePositionDelta)
	r.Float32(&m.TotalPositionDelta)
	r.Float32(&m.MinPositionDelta)
	r.Float32(&m.MaxPositionDelta)
}

// MovementCorrectedEventData is an event sent by the server to correct movement client side.
type MovementCorrectedEventData struct {
	// PositionDelta ...
	PositionDelta float32
	// CheatingScore ...
	CheatingScore float32
	// ScoreThreshold ...
	ScoreThreshold float32
	// DistanceThreshold ...
	DistanceThreshold float32
	// DurationThreshold ...
	DurationThreshold int32
}

// Marshal ...
func (m *MovementCorrectedEventData) Marshal(r IO) {
	r.Float32(&m.PositionDelta)
	r.Float32(&m.CheatingScore)
	r.Float32(&m.ScoreThreshold)
	r.Float32(&m.DistanceThreshold)
	r.Varint32(&m.DurationThreshold)
}

// ExtractHoneyEventData is an event with no purpose.
type ExtractHoneyEventData struct{}

// Marshal ...
func (*ExtractHoneyEventData) Marshal(IO) {}

const (
	WaxNotOxidised   = uint16(0xa609)
	WaxExposed       = uint16(0xa809)
	WaxWeathered     = uint16(0xaa09)
	WaxOxidised      = uint16(0xac09)
	UnWaxNotOxidised = uint16(0xae09)
	UnWaxExposed     = uint16(0xb009)
	UnWaxWeathered   = uint16(0xb209)
	UnWaxOxidised    = uint16(0xfa0a)
)

// WaxedOrUnwaxedCopperEventData is an event sent by the server when a copper block is waxed or unwaxed.
type WaxedOrUnwaxedCopperEventData struct {
	Type uint16
}

// Marshal ...
func (w *WaxedOrUnwaxedCopperEventData) Marshal(r IO) {
	r.Uint16(&w.Type)
}

// SneakCloseToSculkSensorEventData is an event sent by the server when a player sneaks close to an sculk block.
type SneakCloseToSculkSensorEventData struct{}

// Marshal ...
func (u *SneakCloseToSculkSensorEventData) Marshal(r IO) {}

// UnknownEventData is an unimplemented event
type UnknownEventData struct{}

// Marshal ...
func (u *UnknownEventData) Marshal(r IO) {}
