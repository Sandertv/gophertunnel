package protocol

const (
	MemoryCategoryUnknown = iota
	MemoryCategoryInvalidSizeUnknown
	MemoryCategoryActor
	MemoryCategoryActorAnimation
	MemoryCategoryActorRendering
	MemoryCategoryBlockTickingQueues
	MemoryCategoryBiomeStorage
	MemoryCategoryCereal
	MemoryCategoryCircuitSystem
	MemoryCategoryClient
	MemoryCategoryCommands
	MemoryCategoryDBStorage
	MemoryCategoryDebug
	MemoryCategoryDocumentation
	MemoryCategoryECSSystems
	MemoryCategoryFMOD
	MemoryCategoryFonts
	MemoryCategoryImGUI
	MemoryCategoryInput
	MemoryCategoryJsonUI
	MemoryCategoryJsonUIControlFactoryJson
	MemoryCategoryJsonUIControlTree
	MemoryCategoryJsonUIControlTreeControlElement
	MemoryCategoryJsonUIControlTreePopulateDataBinding
	MemoryCategoryJsonUIControlTreePopulateFocus
	MemoryCategoryJsonUIControlTreePopulateLayout
	MemoryCategoryJsonUIControlTreePopulateOther
	MemoryCategoryJsonUIControlTreePopulateSprite
	MemoryCategoryJsonUIControlTreePopulateText
	MemoryCategoryJsonUIControlTreePopulateTTS
	MemoryCategoryJsonUIControlTreeVisibility
	MemoryCategoryJsonUICreateUI
	MemoryCategoryJsonUIDefs
	MemoryCategoryJsonUILayoutManager
	MemoryCategoryJsonUILayoutManagerRemoveDependencies
	MemoryCategoryJsonUILayoutManagerInitVariable
	MemoryCategoryLanguages
	MemoryCategoryLevel
	MemoryCategoryLevelStructures
	MemoryCategoryLevelChunk
	MemoryCategoryLevelChunkGen
	MemoryCategoryLevelChunkGenThreadLocal
	MemoryCategoryLightVolumeManager
	MemoryCategoryNetwork
	MemoryCategoryMarketplace
	MemoryCategoryMaterialDragonCompiledDefinition
	MemoryCategoryMaterialDragonMaterial
	MemoryCategoryMaterialDragonResource
	MemoryCategoryMaterialDragonUniformMap
	MemoryCategoryMaterialRenderMaterial
	MemoryCategoryMaterialRenderMaterialGroup
	MemoryCategoryMaterialVariationManager
	MemoryCategoryMolang
	MemoryCategoryOreUI
	MemoryCategoryPersona
	MemoryCategoryPlayer
	MemoryCategoryRenderChunk
	MemoryCategoryRenderChunkIndexBuffer
	MemoryCategoryRenderChunkVertexBuffer
	MemoryCategoryRendering
	MemoryCategoryRenderingLibrary
	MemoryCategoryRequestLog
	MemoryCategoryResourcePacks
	MemoryCategorySound
	MemoryCategorySubChunkBiomeData
	MemoryCategorySubChunkBlockData
	MemoryCategorySubChunkLightData
	MemoryCategoryTextures
	MemoryCategoryVR
	MemoryCategoryWeatherRenderer
	MemoryCategoryWorldGenerator
	MemoryCategoryTasks
	MemoryCategoryTest
	MemoryCategoryScripting
	MemoryCategoryScriptingRuntime
	MemoryCategoryScriptingContext
	MemoryCategoryScriptingContextBindingsMC
	MemoryCategoryScriptingContextBindingsGT
	MemoryCategoryScriptingContextRun
	MemoryCategoryDataDrivenUI
	MemoryCategoryDataDrivenUIDefs
	MemoryCategoryGameface
	MemoryCategoryGamefaceSystem
	MemoryCategoryGamefaceDOM
	MemoryCategoryGamefaceCSS
	MemoryCategoryGamefaceDisplay
	MemoryCategoryGamefaceTempAllocator
	MemoryCategoryGamefacePoolAllocator
	MemoryCategoryGamefaceDump
	MemoryCategoryGamefaceMedia
	MemoryCategoryGamefaceJSON
	MemoryCategoryGamefaceScriptEngine
)

// MemoryCategoryCounter represents a memory usage counter for a specific category.
type MemoryCategoryCounter struct {
	// Category is the memory category. It is one of the MemoryCategory constants above.
	Category uint8
	// Bytes is the number of bytes used by this category.
	Bytes uint64
}

// Marshal encodes/decodes a MemoryCategoryCounter.
func (x *MemoryCategoryCounter) Marshal(r IO) {
	r.Uint8(&x.Category)
	r.Uint64(&x.Bytes)
}

// EntityDiagnosticTimingInfo represents diagnostics for a specific entity type.
type EntityDiagnosticTimingInfo struct {
	// DisplayName is the name to display for this timing entry.
	DisplayName string
	// Entity is the identifier of the entity that is being timed.
	Entity string
	// PercentOfTotal is the percentage of time that this timing entry has used compared to others.
	PercentOfTotal byte
	// DurationNanos is whole long the timing entry has lasted, in nanoseconds.
	DurationNanos uint64
}

// Marshal encodes/decodes a EntityDiagnosticTimingInfo.
func (x *EntityDiagnosticTimingInfo) Marshal(r IO) {
	r.String(&x.DisplayName)
	r.String(&x.Entity)
	r.Uint8(&x.PercentOfTotal)
	r.Uint64(&x.DurationNanos)
}

// SystemDiagnosticTimingInfo represents diagnostics for a specific system index.
type SystemDiagnosticTimingInfo struct {
	// DisplayName is the name to display for this timing entry.
	DisplayName string
	// SystemIndex is the index of the system that is being timed.
	SystemIndex uint64
	// PercentOfTotal is the percentage of time that this timing entry has used compared to others.
	PercentOfTotal byte
	// DurationNanos is whole long the timing entry has lasted, in nanoseconds.
	DurationNanos uint64
}

// Marshal encodes/decodes a SystemDiagnosticTimingInfo.
func (x *SystemDiagnosticTimingInfo) Marshal(r IO) {
	r.String(&x.DisplayName)
	r.Uint64(&x.SystemIndex)
	r.Uint8(&x.PercentOfTotal)
	r.Uint64(&x.DurationNanos)
}
