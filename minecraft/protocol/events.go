package protocol

import "strings"

// EventData represents an object that holds data specific to an event.
// The data it holds depends on the type.
type EventData interface {
	// Marshal encodes the event data to its binary representation into buf.
	Marshal(w *Writer)
	// Unmarshal decodes a serialised event data object from Reader r into the
	// EventData instance.
	Unmarshal(r *Reader)
}

// AchievementAwardedEventData is the event data sent for achievements.
type AchievementAwardedEventData struct {
	// AchievementID is the ID for the achievement.
	AchievementID int32
}

// Marshal ...
func (a *AchievementAwardedEventData) Marshal(w *Writer) {
	w.Varint32(&a.AchievementID)
}

// Unmarshal ...
func (a *AchievementAwardedEventData) Unmarshal(r *Reader) {
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
func (e *EntityInteractEventData) Marshal(w *Writer) {
	w.Varint32(&e.InteractionType)
	w.Varint32(&e.InteractionEntityType)
	w.Varint32(&e.EntityVariant)
	w.Uint8(&e.EntityColour)
}

// Unmarshal ...
func (e *EntityInteractEventData) Unmarshal(r *Reader) {
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
func (p *PortalBuiltEventData) Marshal(w *Writer) {
	w.Varint32(&p.DimensionID)
}

// Unmarshal ...
func (p *PortalBuiltEventData) Unmarshal(r *Reader) {
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
func (p *PortalUsedEventData) Marshal(w *Writer) {
	w.Varint32(&p.FromDimensionID)
	w.Varint32(&p.ToDimensionID)
}

// Unmarshal ...
func (p *PortalUsedEventData) Unmarshal(r *Reader) {
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
func (m *MobKilledEventData) Marshal(w *Writer) {
	w.Varint64(&m.KillerEntityUniqueID)
	w.Varint64(&m.VictimEntityUniqueID)
	w.Varint32(&m.KillerEntityType)
	w.Varint32(&m.EntityDamageCause)
	w.Varint32(&m.VillagerTradeTier)
	w.String(&m.VillagerDisplayName)
}

// Unmarshal ...
func (m *MobKilledEventData) Unmarshal(r *Reader) {
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
func (c *CauldronUsedEventData) Marshal(w *Writer) {
	w.Varint32(&c.PotionID)
	w.Varint32(&c.Colour)
	w.Varint32(&c.FillLevel)
}

// Unmarshal ...
func (c *CauldronUsedEventData) Unmarshal(r *Reader) {
	r.Varint32(&c.PotionID)
	r.Varint32(&c.Colour)
	r.Varint32(&c.FillLevel)
}

// PlayerDiedEventData is the event data sent when a player dies.
type PlayerDiedEventData struct {
	// AttackerEntityID ...
	AttackerEntityID int32
	// EntityDamageCause ...
	EntityDamageCause int32
}

// Marshal ...
func (p *PlayerDiedEventData) Marshal(w *Writer) {
	w.Varint32(&p.AttackerEntityID)
	w.Varint32(&p.EntityDamageCause)
}

// Unmarshal ...
func (p *PlayerDiedEventData) Unmarshal(r *Reader) {
	r.Varint32(&p.AttackerEntityID)
	r.Varint32(&p.EntityDamageCause)
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
func (b *BossKilledEventData) Marshal(w *Writer) {
	w.Varint64(&b.BossEntityUniqueID)
	w.Varint32(&b.PlayerPartySize)
	w.Varint32(&b.InteractionEntityType)
}

// Unmarshal ...
func (b *BossKilledEventData) Unmarshal(r *Reader) {
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
func (a *AgentCommandEventData) Marshal(w *Writer) {
	w.Varint32(&a.AgentResult)
	w.Varint32(&a.DataValue)
	w.String(&a.Command)
	w.String(&a.DataKey)
	w.String(&a.Output)
}

// Unmarshal ...
func (a *AgentCommandEventData) Unmarshal(r *Reader) {
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
func (p *PatternRemovedEventData) Marshal(w *Writer) {
	w.Varint32(&p.ItemID)
	w.Varint32(&p.AuxValue)
	w.Varint32(&p.PatternsSize)
	w.Varint32(&p.PatternIndex)
	w.Varint32(&p.PatternColour)
}

// Unmarshal ...
func (p *PatternRemovedEventData) Unmarshal(r *Reader) {
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
	// OutputMessages ...
	OutputMessages []string
}

// Marshal ...
func (s *SlashCommandExecutedEventData) Marshal(w *Writer) {
	outputMessagesSize := int32(len(s.OutputMessages))
	outputMessagesJoined := strings.Join(s.OutputMessages, ";")

	w.Varint32(&s.SuccessCount)
	w.Varint32(&outputMessagesSize)
	w.String(&s.CommandName)
	w.String(&outputMessagesJoined)
}

// Unmarshal ...
func (s *SlashCommandExecutedEventData) Unmarshal(r *Reader) {
	var outputMessagesSize int32
	var outputMessagesJoined string

	r.Varint32(&s.SuccessCount)
	r.Varint32(&outputMessagesSize)
	r.String(&s.CommandName)
	r.String(&outputMessagesJoined)

	s.OutputMessages = strings.Split(outputMessagesJoined, ";")
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
func (f *FishBucketedEventData) Marshal(w *Writer) {
	w.Varint32(&f.Pattern)
	w.Varint32(&f.Preset)
	w.Varint32(&f.BucketedEntityType)
	w.Bool(&f.Release)
}

// Unmarshal ...
func (f *FishBucketedEventData) Unmarshal(r *Reader) {
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
func (m *MobBornEventData) Marshal(w *Writer) {
	w.Varint32(&m.EntityType)
	w.Varint32(&m.Variant)
	w.Uint8(&m.Colour)
}

// Unmarshal ...
func (m *MobBornEventData) Unmarshal(r *Reader) {
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
func (p *PetDiedEventData) Marshal(w *Writer) {
	w.Bool(&p.KilledByOwner)
	w.Varint64(&p.KillerEntityUniqueID)
	w.Varint64(&p.PetEntityUniqueID)
	w.Varint32(&p.EntityDamageCause)
	w.Varint32(&p.PetEntityType)
}

// Unmarshal ...
func (p *PetDiedEventData) Unmarshal(r *Reader) {
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
func (c *CauldronInteractEventData) Marshal(w *Writer) {
	w.Varint32(&c.BlockInteractionType)
	w.Varint32(&c.ItemID)
}

// Unmarshal ...
func (c *CauldronInteractEventData) Unmarshal(r *Reader) {
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
func (c *ComposterInteractEventData) Marshal(w *Writer) {
	w.Varint32(&c.BlockInteractionType)
	w.Varint32(&c.ItemID)
}

// Unmarshal ...
func (c *ComposterInteractEventData) Unmarshal(r *Reader) {
	r.Varint32(&c.BlockInteractionType)
	r.Varint32(&c.ItemID)
}

// BellUsedEventData is the event data sent when a bell is used.
type BellUsedEventData struct {
	// ItemID ...
	ItemID int32
}

// Marshal ...
func (b *BellUsedEventData) Marshal(w *Writer) {
	w.Varint32(&b.ItemID)
}

// Unmarshal ...
func (b *BellUsedEventData) Unmarshal(r *Reader) {
	r.Varint32(&b.ItemID)
}

// EntityDefinitionTriggerEventData is an event used for an unknown purpose.
type EntityDefinitionTriggerEventData struct {
	// EventName ...
	EventName string
}

// Marshal ...
func (e *EntityDefinitionTriggerEventData) Marshal(w *Writer) {
	w.String(&e.EventName)
}

// Unmarshal ...
func (e *EntityDefinitionTriggerEventData) Unmarshal(r *Reader) {
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
func (ra *RaidUpdateEventData) Marshal(w *Writer) {
	w.Varint32(&ra.CurrentRaidWave)
	w.Varint32(&ra.TotalRaidWaves)
	w.Bool(&ra.WonRaid)
}

// Unmarshal ...
func (ra *RaidUpdateEventData) Unmarshal(r *Reader) {
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
func (m *MovementAnomalyEventData) Marshal(w *Writer) {
	w.Uint8(&m.EventType)
	w.Float32(&m.CheatingScore)
	w.Float32(&m.AveragePositionDelta)
	w.Float32(&m.TotalPositionDelta)
	w.Float32(&m.MinPositionDelta)
	w.Float32(&m.MaxPositionDelta)
}

// Unmarshal ...
func (m *MovementAnomalyEventData) Unmarshal(r *Reader) {
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
func (m *MovementCorrectedEventData) Marshal(w *Writer) {
	w.Float32(&m.PositionDelta)
	w.Float32(&m.CheatingScore)
	w.Float32(&m.ScoreThreshold)
	w.Float32(&m.DistanceThreshold)
	w.Varint32(&m.DurationThreshold)
}

// Unmarshal ...
func (m *MovementCorrectedEventData) Unmarshal(r *Reader) {
	r.Float32(&m.PositionDelta)
	r.Float32(&m.CheatingScore)
	r.Float32(&m.ScoreThreshold)
	r.Float32(&m.DistanceThreshold)
	r.Varint32(&m.DurationThreshold)
}

// ExtractHoneyEventData is an event with no purpose.
type ExtractHoneyEventData struct{}

// Marshal ...
func (*ExtractHoneyEventData) Marshal(*Writer) {}

// Unmarshal ...
func (*ExtractHoneyEventData) Unmarshal(*Reader) {}
