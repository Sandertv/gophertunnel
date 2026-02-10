package protocol

const (
	MemoryCategoryUnknown = iota
	MemoryCategoryInvalidSizeUnknown
	MemoryCategoryActor
	MemoryCategoryActorAnimation
	MemoryCategoryActorRendering
	MemoryCategoryBalancer
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
