package protocol

const (
	ContainerAnvilInput = iota
	ContainerAnvilMaterial
	ContainerAnvilResultPreview
	ContainerSmithingTableInput
	ContainerSmithingTableMaterial
	ContainerSmithingTableResultPreview
	ContainerArmor
	ContainerLevelEntity
	ContainerBeaconPayment
	ContainerBrewingStandInput
	ContainerBrewingStandResult
	ContainerBrewingStandFuel
	ContainerCombinedHotBarAndInventory
	ContainerCraftingInput
	ContainerCraftingOutputPreview
	ContainerRecipeConstruction
	ContainerRecipeNature
	ContainerRecipeItems
	ContainerRecipeSearch
	ContainerRecipeSearchBar
	ContainerRecipeEquipment
	ContainerRecipeBook
	ContainerEnchantingInput
	ContainerEnchantingMaterial
	ContainerFurnaceFuel
	ContainerFurnaceIngredient
	ContainerFurnaceResult
	ContainerHorseEquip
	ContainerHotBar
	ContainerInventory
	ContainerShulkerBox
	ContainerTradeIngredientOne
	ContainerTradeIngredientTwo
	ContainerTradeResultPreview
	ContainerOffhand
	ContainerCompoundCreatorInput
	ContainerCompoundCreatorOutputPreview
	ContainerElementConstructorOutputPreview
	ContainerMaterialReducerInput
	ContainerMaterialReducerOutput
	ContainerLabTableInput
	ContainerLoomInput
	ContainerLoomDye
	ContainerLoomMaterial
	ContainerLoomResultPreview
	ContainerBlastFurnaceIngredient
	ContainerSmokerIngredient
	ContainerTradeTwoIngredientOne
	ContainerTradeTwoIngredientTwo
	ContainerTradeTwoResultPreview
	ContainerGrindstoneInput
	ContainerGrindstoneAdditional
	ContainerGrindstoneResultPreview
	ContainerStonecutterInput
	ContainerStonecutterResultPreview
	ContainerCartographyInput
	ContainerCartographyAdditional
	ContainerCartographyResultPreview
	ContainerBarrel
	ContainerCursor
	ContainerCreatedOutput
	ContainerSmithingTableTemplate
	ContainerCrafterLevelEntity
	ContainerDynamic
)

const (
	ContainerTypeInventory = iota - 1
	ContainerTypeContainer
	ContainerTypeWorkbench
	ContainerTypeFurnace
	ContainerTypeEnchantment
	ContainerTypeBrewingStand
	ContainerTypeAnvil
	ContainerTypeDispenser
	ContainerTypeDropper
	ContainerTypeHopper
	ContainerTypeCauldron
	ContainerTypeCartChest
	ContainerTypeCartHopper
	ContainerTypeHorse
	ContainerTypeBeacon
	ContainerTypeStructureEditor
	ContainerTypeTrade
	ContainerTypeCommandBlock
	ContainerTypeJukebox
	ContainerTypeArmour
	ContainerTypeHand
	ContainerTypeCompoundCreator
	ContainerTypeElementConstructor
	ContainerTypeMaterialReducer
	ContainerTypeLabTable
	ContainerTypeLoom
	ContainerTypeLectern
	ContainerTypeGrindstone
	ContainerTypeBlastFurnace
	ContainerTypeSmoker
	ContainerTypeStonecutter
	ContainerTypeCartography
	ContainerTypeHUD
	ContainerTypeJigsawEditor
	ContainerTypeSmithingTable
	ContainerTypeChestBoat
	ContainerTypeDecoratedPot
	ContainerTypeCrafter
)

// FullContainerName contains information required to identify a container in a StackRequestSlotInfo.
type FullContainerName struct {
	// ContainerID is the ID of the container that the slot was in.
	ContainerID byte
	// DynamicContainerID is the ID of the container if it is dynamic. If the container is not dynamic, this
	// field should be left empty. A non-optional value of 0 is assumed to be non-empty.
	DynamicContainerID Optional[uint32]
}

func (x *FullContainerName) Marshal(r IO) {
	r.Uint8(&x.ContainerID)
	OptionalFunc(r, &x.DynamicContainerID, r.Uint32)
}
