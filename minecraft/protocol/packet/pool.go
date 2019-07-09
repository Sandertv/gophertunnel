package packet

import (
	"reflect"
)

// Pool is a map holding packets indexed by a packet ID.
type Pool map[uint32]Packet

// NewPool returns a new pool with all supported packets sent. Packets may be retrieved from it simply by
// indexing it with the packet ID.
func NewPool() Pool {
	return Pool{
		IDLogin:                      &Login{},
		IDPlayStatus:                 &PlayStatus{},
		IDServerToClientHandshake:    &ServerToClientHandshake{},
		IDClientToServerHandshake:    &ClientToServerHandshake{},
		IDDisconnect:                 &Disconnect{},
		IDResourcePacksInfo:          &ResourcePacksInfo{},
		IDResourcePackStack:          &ResourcePackStack{},
		IDResourcePackClientResponse: &ResourcePackClientResponse{},
		IDText:                       &Text{},
		IDSetTime:                    &SetTime{},
		IDStartGame:                  &StartGame{},
		IDAddPlayer:                  &AddPlayer{},
		IDAddEntity:                  &AddEntity{},
		IDRemoveEntity:               &RemoveEntity{},
		IDAddItemEntity:              &AddItemEntity{},
		// ---
		IDTakeItemEntity:     &TakeItemEntity{},
		IDMoveEntityAbsolute: &MoveEntityAbsolute{},
		IDMovePlayer:         &MovePlayer{},
		IDRiderJump:          &RiderJump{},
		IDUpdateBlock:        &UpdateBlock{},
		IDAddPainting:        &AddPainting{},
		IDExplode:            &Explode{},
		// ---
		IDLevelEvent:                  &LevelEvent{},
		IDBlockEvent:                  &BlockEvent{},
		IDEntityEvent:                 &EntityEvent{},
		IDMobEffect:                   &MobEffect{},
		IDUpdateAttributes:            &UpdateAttributes{},
		IDInventoryTransaction:        &InventoryTransaction{},
		IDMobEquipment:                &MobEquipment{},
		IDMobArmourEquipment:          &MobArmourEquipment{},
		IDInteract:                    &Interact{},
		IDBlockPickRequest:            &BlockPickRequest{},
		IDEntityPickRequest:           &EntityPickRequest{},
		IDPlayerAction:                &PlayerAction{},
		IDEntityFall:                  &EntityFall{},
		IDHurtArmour:                  &HurtArmour{},
		IDSetEntityData:               &SetEntityData{},
		IDSetEntityMotion:             &SetEntityMotion{},
		IDSetEntityLink:               &SetEntityLink{},
		IDSetHealth:                   &SetHealth{},
		IDSetSpawnPosition:            &SetSpawnPosition{},
		IDAnimate:                     &Animate{},
		IDRespawn:                     &Respawn{},
		IDContainerOpen:               &ContainerOpen{},
		IDContainerClose:              &ContainerClose{},
		IDPlayerHotBar:                &PlayerHotBar{},
		IDInventoryContent:            &InventoryContent{},
		IDInventorySlot:               &InventorySlot{},
		IDContainerSetData:            &ContainerSetData{},
		IDCraftingData:                &CraftingData{},
		IDCraftingEvent:               &CraftingEvent{},
		IDGUIDataPickItem:             &GUIDataPickItem{},
		IDAdventureSettings:           &AdventureSettings{},
		IDBlockEntityData:             &BlockEntityData{},
		IDPlayerInput:                 &PlayerInput{},
		IDFullChunkData:               &FullChunkData{},
		IDSetCommandsEnabled:          &SetCommandsEnabled{},
		IDSetDifficulty:               &SetDifficulty{},
		IDChangeDimension:             &ChangeDimension{},
		IDSetPlayerGameType:           &SetPlayerGameType{},
		IDPlayerList:                  &PlayerList{},
		IDSimpleEvent:                 &SimpleEvent{},
		IDEvent:                       &Event{},
		IDSpawnExperienceOrb:          &SpawnExperienceOrb{},
		IDClientBoundMapItemData:      &ClientBoundMapItemData{},
		IDMapInfoRequest:              &MapInfoRequest{},
		IDRequestChunkRadius:          &RequestChunkRadius{},
		IDChunkRadiusUpdated:          &ChunkRadiusUpdated{},
		IDItemFrameDropItem:           &ItemFrameDropItem{},
		IDGameRulesChanged:            &GameRulesChanged{},
		IDCamera:                      &Camera{},
		IDBossEvent:                   &BossEvent{},
		IDShowCredits:                 &ShowCredits{},
		IDAvailableCommands:           &AvailableCommands{},
		IDCommandRequest:              &CommandRequest{},
		IDCommandBlockUpdate:          &CommandBlockUpdate{},
		IDCommandOutput:               &CommandOutput{},
		IDUpdateTrade:                 &UpdateTrade{},
		IDUpdateEquip:                 &UpdateEquip{},
		IDResourcePackDataInfo:        &ResourcePackDataInfo{},
		IDResourcePackChunkData:       &ResourcePackChunkData{},
		IDResourcePackChunkRequest:    &ResourcePackChunkRequest{},
		IDTransfer:                    &Transfer{},
		IDPlaySound:                   &PlaySound{},
		IDStopSound:                   &StopSound{},
		IDSetTitle:                    &SetTitle{},
		IDAddBehaviourTree:            &AddBehaviourTree{},
		IDStructureBlockUpdate:        &StructureBlockUpdate{},
		IDShowStoreOffer:              &ShowStoreOffer{},
		IDPurchaseReceipt:             &PurchaseReceipt{},
		IDPlayerSkin:                  &PlayerSkin{},
		IDSubClientLogin:              &SubClientLogin{},
		IDAutomationClientConnect:     &AutomationClientConnect{},
		IDSetLastHurtBy:               &SetLastHurtBy{},
		IDBookEdit:                    &BookEdit{},
		IDNPCRequest:                  &NPCRequest{},
		IDPhotoTransfer:               &PhotoTransfer{},
		IDModalFormRequest:            &ModalFormRequest{},
		IDModalFormResponse:           &ModalFormResponse{},
		IDServerSettingsRequest:       &ServerSettingsRequest{},
		IDServerSettingsResponse:      &ServerSettingsResponse{},
		IDShowProfile:                 &ShowProfile{},
		IDSetDefaultGameType:          &SetDefaultGameType{},
		IDRemoveObjective:             &RemoveObjective{},
		IDSetDisplayObjective:         &SetDisplayObjective{},
		IDSetScore:                    &SetScore{},
		IDLabTable:                    &LabTable{},
		IDUpdateBlockSynced:           &UpdateBlockSynced{},
		IDMoveEntityDelta:             &MoveEntityDelta{},
		IDSetScoreboardIdentity:       &SetScoreboardIdentity{},
		IDSetLocalPlayerAsInitialised: &SetLocalPlayerAsInitialised{},
		IDUpdateSoftEnum:              &UpdateSoftEnum{},
		IDNetworkStackLatency:         &NetworkStackLatency{},
		// ---
		IDScriptCustomEvent:           &ScriptCustomEvent{},
		IDSpawnParticleEffect:         &SpawnParticleEffect{},
		IDAvailableEntityIdentifiers:  &AvailableEntityIdentifiers{},
		IDNetworkChunkPublisherUpdate: &NetworkChunkPublisherUpdate{},
		IDBiomeDefinitionList:         &BiomeDefinitionList{},
		IDLevelSoundEvent:             &LevelSoundEvent{},
		IDLecternUpdate:               &LecternUpdate{},
	}
}

// PacketsByName is a map holding a function to create a new packet for each packet registered in Pool. These
// functions are indexed using the exact packet name they return.
var PacketsByName = map[string]func() Packet{}

func init() {
	for _, packet := range NewPool() {
		pk := packet
		PacketsByName[reflect.TypeOf(pk).Elem().Name()] = func() Packet {
			return reflect.New(reflect.TypeOf(pk).Elem()).Interface().(Packet)
		}
	}
}
