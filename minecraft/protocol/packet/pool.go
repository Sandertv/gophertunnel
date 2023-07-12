package packet

// RegisterPacketFromClient registers a function that returns a packet for a
// specific ID. Packets with this ID coming in from connections will resolve to
// the packet returned by the function passed. noinspection
func RegisterPacketFromClient(id uint32, pk func() Packet) {
	packetsFromClient[id] = pk
}

// RegisterPacketFromServer registers a function that returns a packet for a
// specific ID. Packets with this ID coming in from connections will resolve to
// the packet returned by the function passed. noinspection
func RegisterPacketFromServer(id uint32, pk func() Packet) {
	packetsFromServer[id] = pk
}

// packetsFromClient holds packets that could be sent by the client.
var packetsFromClient = map[uint32]func() Packet{}

// packetsFromServer holds packets that could be sent by the server.
var packetsFromServer = map[uint32]func() Packet{}

// Pool is a map holding packets indexed by a packet ID.
type Pool map[uint32]func() Packet

// NewClientPool returns a new pool containing packets sent by a client.
// Packets may be retrieved from it simply by indexing it with the packet ID.
func NewClientPool() Pool {
	p := Pool{}
	for id, pk := range packetsFromClient {
		p[id] = pk
	}
	return p
}

// NewServerPool returns a new pool containing packets sent by a server.
// Packets may be retrieved from it simply by indexing it with the packet ID.
func NewServerPool() Pool {
	p := Pool{}
	for id, pk := range packetsFromServer {
		p[id] = pk
	}
	return p
}

func init() {
	// TODO: Remove packets from this list that are not sent by the server.
	serverOriginating := map[uint32]func() Packet{
		IDLogin:                      func() Packet { return &Login{} },
		IDPlayStatus:                 func() Packet { return &PlayStatus{} },
		IDServerToClientHandshake:    func() Packet { return &ServerToClientHandshake{} },
		IDClientToServerHandshake:    func() Packet { return &ClientToServerHandshake{} },
		IDDisconnect:                 func() Packet { return &Disconnect{} },
		IDResourcePacksInfo:          func() Packet { return &ResourcePacksInfo{} },
		IDResourcePackStack:          func() Packet { return &ResourcePackStack{} },
		IDResourcePackClientResponse: func() Packet { return &ResourcePackClientResponse{} },
		IDText:                       func() Packet { return &Text{} },
		IDSetTime:                    func() Packet { return &SetTime{} },
		IDStartGame:                  func() Packet { return &StartGame{} },
		IDAddPlayer:                  func() Packet { return &AddPlayer{} },
		IDAddActor:                   func() Packet { return &AddActor{} },
		IDRemoveActor:                func() Packet { return &RemoveActor{} },
		IDAddItemActor:               func() Packet { return &AddItemActor{} },
		// ---
		IDTakeItemActor:     func() Packet { return &TakeItemActor{} },
		IDMoveActorAbsolute: func() Packet { return &MoveActorAbsolute{} },
		IDMovePlayer:        func() Packet { return &MovePlayer{} },
		IDPassengerJump:     func() Packet { return &PassengerJump{} },
		IDUpdateBlock:       func() Packet { return &UpdateBlock{} },
		IDAddPainting:       func() Packet { return &AddPainting{} },
		IDTickSync:          func() Packet { return &TickSync{} },
		// ---
		IDLevelEvent:           func() Packet { return &LevelEvent{} },
		IDBlockEvent:           func() Packet { return &BlockEvent{} },
		IDActorEvent:           func() Packet { return &ActorEvent{} },
		IDMobEffect:            func() Packet { return &MobEffect{} },
		IDUpdateAttributes:     func() Packet { return &UpdateAttributes{} },
		IDInventoryTransaction: func() Packet { return &InventoryTransaction{} },
		IDMobEquipment:         func() Packet { return &MobEquipment{} },
		IDMobArmourEquipment:   func() Packet { return &MobArmourEquipment{} },
		IDInteract:             func() Packet { return &Interact{} },
		IDBlockPickRequest:     func() Packet { return &BlockPickRequest{} },
		IDActorPickRequest:     func() Packet { return &ActorPickRequest{} },
		IDPlayerAction:         func() Packet { return &PlayerAction{} },
		// ---
		IDHurtArmour:                  func() Packet { return &HurtArmour{} },
		IDSetActorData:                func() Packet { return &SetActorData{} },
		IDSetActorMotion:              func() Packet { return &SetActorMotion{} },
		IDSetActorLink:                func() Packet { return &SetActorLink{} },
		IDSetHealth:                   func() Packet { return &SetHealth{} },
		IDSetSpawnPosition:            func() Packet { return &SetSpawnPosition{} },
		IDAnimate:                     func() Packet { return &Animate{} },
		IDRespawn:                     func() Packet { return &Respawn{} },
		IDContainerOpen:               func() Packet { return &ContainerOpen{} },
		IDContainerClose:              func() Packet { return &ContainerClose{} },
		IDPlayerHotBar:                func() Packet { return &PlayerHotBar{} },
		IDInventoryContent:            func() Packet { return &InventoryContent{} },
		IDInventorySlot:               func() Packet { return &InventorySlot{} },
		IDContainerSetData:            func() Packet { return &ContainerSetData{} },
		IDCraftingData:                func() Packet { return &CraftingData{} },
		IDCraftingEvent:               func() Packet { return &CraftingEvent{} },
		IDGUIDataPickItem:             func() Packet { return &GUIDataPickItem{} },
		IDAdventureSettings:           func() Packet { return &AdventureSettings{} },
		IDBlockActorData:              func() Packet { return &BlockActorData{} },
		IDPlayerInput:                 func() Packet { return &PlayerInput{} },
		IDLevelChunk:                  func() Packet { return &LevelChunk{} },
		IDSetCommandsEnabled:          func() Packet { return &SetCommandsEnabled{} },
		IDSetDifficulty:               func() Packet { return &SetDifficulty{} },
		IDChangeDimension:             func() Packet { return &ChangeDimension{} },
		IDSetPlayerGameType:           func() Packet { return &SetPlayerGameType{} },
		IDPlayerList:                  func() Packet { return &PlayerList{} },
		IDSimpleEvent:                 func() Packet { return &SimpleEvent{} },
		IDEvent:                       func() Packet { return &Event{} },
		IDSpawnExperienceOrb:          func() Packet { return &SpawnExperienceOrb{} },
		IDClientBoundMapItemData:      func() Packet { return &ClientBoundMapItemData{} },
		IDMapInfoRequest:              func() Packet { return &MapInfoRequest{} },
		IDRequestChunkRadius:          func() Packet { return &RequestChunkRadius{} },
		IDChunkRadiusUpdated:          func() Packet { return &ChunkRadiusUpdated{} },
		IDItemFrameDropItem:           func() Packet { return &ItemFrameDropItem{} },
		IDGameRulesChanged:            func() Packet { return &GameRulesChanged{} },
		IDCamera:                      func() Packet { return &Camera{} },
		IDBossEvent:                   func() Packet { return &BossEvent{} },
		IDShowCredits:                 func() Packet { return &ShowCredits{} },
		IDAvailableCommands:           func() Packet { return &AvailableCommands{} },
		IDCommandRequest:              func() Packet { return &CommandRequest{} },
		IDCommandBlockUpdate:          func() Packet { return &CommandBlockUpdate{} },
		IDCommandOutput:               func() Packet { return &CommandOutput{} },
		IDUpdateTrade:                 func() Packet { return &UpdateTrade{} },
		IDUpdateEquip:                 func() Packet { return &UpdateEquip{} },
		IDResourcePackDataInfo:        func() Packet { return &ResourcePackDataInfo{} },
		IDResourcePackChunkData:       func() Packet { return &ResourcePackChunkData{} },
		IDResourcePackChunkRequest:    func() Packet { return &ResourcePackChunkRequest{} },
		IDTransfer:                    func() Packet { return &Transfer{} },
		IDPlaySound:                   func() Packet { return &PlaySound{} },
		IDStopSound:                   func() Packet { return &StopSound{} },
		IDSetTitle:                    func() Packet { return &SetTitle{} },
		IDAddBehaviourTree:            func() Packet { return &AddBehaviourTree{} },
		IDStructureBlockUpdate:        func() Packet { return &StructureBlockUpdate{} },
		IDShowStoreOffer:              func() Packet { return &ShowStoreOffer{} },
		IDPurchaseReceipt:             func() Packet { return &PurchaseReceipt{} },
		IDPlayerSkin:                  func() Packet { return &PlayerSkin{} },
		IDSubClientLogin:              func() Packet { return &SubClientLogin{} },
		IDAutomationClientConnect:     func() Packet { return &AutomationClientConnect{} },
		IDSetLastHurtBy:               func() Packet { return &SetLastHurtBy{} },
		IDBookEdit:                    func() Packet { return &BookEdit{} },
		IDNPCRequest:                  func() Packet { return &NPCRequest{} },
		IDPhotoTransfer:               func() Packet { return &PhotoTransfer{} },
		IDModalFormRequest:            func() Packet { return &ModalFormRequest{} },
		IDModalFormResponse:           func() Packet { return &ModalFormResponse{} },
		IDServerSettingsRequest:       func() Packet { return &ServerSettingsRequest{} },
		IDServerSettingsResponse:      func() Packet { return &ServerSettingsResponse{} },
		IDShowProfile:                 func() Packet { return &ShowProfile{} },
		IDSetDefaultGameType:          func() Packet { return &SetDefaultGameType{} },
		IDRemoveObjective:             func() Packet { return &RemoveObjective{} },
		IDSetDisplayObjective:         func() Packet { return &SetDisplayObjective{} },
		IDSetScore:                    func() Packet { return &SetScore{} },
		IDLabTable:                    func() Packet { return &LabTable{} },
		IDUpdateBlockSynced:           func() Packet { return &UpdateBlockSynced{} },
		IDMoveActorDelta:              func() Packet { return &MoveActorDelta{} },
		IDSetScoreboardIdentity:       func() Packet { return &SetScoreboardIdentity{} },
		IDSetLocalPlayerAsInitialised: func() Packet { return &SetLocalPlayerAsInitialised{} },
		IDUpdateSoftEnum:              func() Packet { return &UpdateSoftEnum{} },
		IDNetworkStackLatency:         func() Packet { return &NetworkStackLatency{} },
		// ---
		IDScriptCustomEvent:           func() Packet { return &ScriptCustomEvent{} },
		IDSpawnParticleEffect:         func() Packet { return &SpawnParticleEffect{} },
		IDAvailableActorIdentifiers:   func() Packet { return &AvailableActorIdentifiers{} },
		IDNetworkChunkPublisherUpdate: func() Packet { return &NetworkChunkPublisherUpdate{} },
		IDBiomeDefinitionList:         func() Packet { return &BiomeDefinitionList{} },
		IDLevelSoundEvent:             func() Packet { return &LevelSoundEvent{} },
		IDLevelEventGeneric:           func() Packet { return &LevelEventGeneric{} },
		IDLecternUpdate:               func() Packet { return &LecternUpdate{} },
		// ---
		IDAddEntity:                     func() Packet { return &AddEntity{} },
		IDRemoveEntity:                  func() Packet { return &RemoveEntity{} },
		IDClientCacheStatus:             func() Packet { return &ClientCacheStatus{} },
		IDOnScreenTextureAnimation:      func() Packet { return &OnScreenTextureAnimation{} },
		IDMapCreateLockedCopy:           func() Packet { return &MapCreateLockedCopy{} },
		IDStructureTemplateDataRequest:  func() Packet { return &StructureTemplateDataRequest{} },
		IDStructureTemplateDataResponse: func() Packet { return &StructureTemplateDataResponse{} },
		// ---
		IDClientCacheBlobStatus:             func() Packet { return &ClientCacheBlobStatus{} },
		IDClientCacheMissResponse:           func() Packet { return &ClientCacheMissResponse{} },
		IDEducationSettings:                 func() Packet { return &EducationSettings{} },
		IDEmote:                             func() Packet { return &Emote{} },
		IDMultiPlayerSettings:               func() Packet { return &MultiPlayerSettings{} },
		IDSettingsCommand:                   func() Packet { return &SettingsCommand{} },
		IDAnvilDamage:                       func() Packet { return &AnvilDamage{} },
		IDCompletedUsingItem:                func() Packet { return &CompletedUsingItem{} },
		IDNetworkSettings:                   func() Packet { return &NetworkSettings{} },
		IDPlayerAuthInput:                   func() Packet { return &PlayerAuthInput{} },
		IDCreativeContent:                   func() Packet { return &CreativeContent{} },
		IDPlayerEnchantOptions:              func() Packet { return &PlayerEnchantOptions{} },
		IDItemStackRequest:                  func() Packet { return &ItemStackRequest{} },
		IDItemStackResponse:                 func() Packet { return &ItemStackResponse{} },
		IDPlayerArmourDamage:                func() Packet { return &PlayerArmourDamage{} },
		IDCodeBuilder:                       func() Packet { return &CodeBuilder{} },
		IDUpdatePlayerGameType:              func() Packet { return &UpdatePlayerGameType{} },
		IDEmoteList:                         func() Packet { return &EmoteList{} },
		IDPositionTrackingDBServerBroadcast: func() Packet { return &PositionTrackingDBServerBroadcast{} },
		IDPositionTrackingDBClientRequest:   func() Packet { return &PositionTrackingDBClientRequest{} },
		IDDebugInfo:                         func() Packet { return &DebugInfo{} },
		IDPacketViolationWarning:            func() Packet { return &PacketViolationWarning{} },
		IDMotionPredictionHints:             func() Packet { return &MotionPredictionHints{} },
		IDAnimateEntity:                     func() Packet { return &AnimateEntity{} },
		IDCameraShake:                       func() Packet { return &CameraShake{} },
		IDPlayerFog:                         func() Packet { return &PlayerFog{} },
		IDCorrectPlayerMovePrediction:       func() Packet { return &CorrectPlayerMovePrediction{} },
		IDItemComponent:                     func() Packet { return &ItemComponent{} },
		IDFilterText:                        func() Packet { return &FilterText{} },
		IDClientBoundDebugRenderer:          func() Packet { return &ClientBoundDebugRenderer{} },
		IDSyncActorProperty:                 func() Packet { return &SyncActorProperty{} },
		IDAddVolumeEntity:                   func() Packet { return &AddVolumeEntity{} },
		IDRemoveVolumeEntity:                func() Packet { return &RemoveVolumeEntity{} },
		IDSimulationType:                    func() Packet { return &SimulationType{} },
		IDNPCDialogue:                       func() Packet { return &NPCDialogue{} },
		IDEducationResourceURI:              func() Packet { return &EducationResourceURI{} },
		IDCreatePhoto:                       func() Packet { return &CreatePhoto{} },
		IDUpdateSubChunkBlocks:              func() Packet { return &UpdateSubChunkBlocks{} },
		IDPhotoInfoRequest:                  func() Packet { return &PhotoInfoRequest{} },
		IDSubChunk:                          func() Packet { return &SubChunk{} },
		IDSubChunkRequest:                   func() Packet { return &SubChunkRequest{} },
		IDClientStartItemCooldown:           func() Packet { return &ClientStartItemCooldown{} },
		IDScriptMessage:                     func() Packet { return &ScriptMessage{} },
		IDCodeBuilderSource:                 func() Packet { return &CodeBuilderSource{} },
		IDTickingAreasLoadStatus:            func() Packet { return &TickingAreasLoadStatus{} },
		IDDimensionData:                     func() Packet { return &DimensionData{} },
		IDAgentAction:                       func() Packet { return &AgentAction{} },
		IDChangeMobProperty:                 func() Packet { return &ChangeMobProperty{} },
		IDLessonProgress:                    func() Packet { return &LessonProgress{} },
		IDRequestAbility:                    func() Packet { return &RequestAbility{} },
		IDRequestPermissions:                func() Packet { return &RequestPermissions{} },
		IDToastRequest:                      func() Packet { return &ToastRequest{} },
		IDUpdateAbilities:                   func() Packet { return &UpdateAbilities{} },
		IDUpdateAdventureSettings:           func() Packet { return &UpdateAdventureSettings{} },
		IDDeathInfo:                         func() Packet { return &DeathInfo{} },
		IDEditorNetwork:                     func() Packet { return &EditorNetwork{} },
		IDFeatureRegistry:                   func() Packet { return &FeatureRegistry{} },
		IDServerStats:                       func() Packet { return &ServerStats{} },
		IDRequestNetworkSettings:            func() Packet { return &RequestNetworkSettings{} },
		IDGameTestRequest:                   func() Packet { return &GameTestRequest{} },
		IDGameTestResults:                   func() Packet { return &GameTestResults{} },
		IDUpdateClientInputLocks:            func() Packet { return &UpdateClientInputLocks{} },
		IDClientCheatAbility:                func() Packet { return &ClientCheatAbility{} },
		IDCameraPresets:                     func() Packet { return &CameraPresets{} },
		IDUnlockedRecipes:                   func() Packet { return &UnlockedRecipes{} },
		// ---
		IDCameraInstruction:             func() Packet { return &CameraInstruction{} },
		IDCompressedBiomeDefinitionList: func() Packet { return &CompressedBiomeDefinitionList{} },
		IDTrimData:                      func() Packet { return &TrimData{} },
		IDOpenSign:                      func() Packet { return &OpenSign{} },
		IDAgentAnimation:                func() Packet { return &AgentAnimation{} },
	}
	for id, pk := range serverOriginating {
		RegisterPacketFromServer(id, pk)
	}
}

// Packets sent by the client:
func init() {
	clientOriginating := map[uint32]func() Packet{
		IDLogin:                           func() Packet { return &Login{} },
		IDClientToServerHandshake:         func() Packet { return &ClientToServerHandshake{} },
		IDResourcePackClientResponse:      func() Packet { return &ResourcePackClientResponse{} },
		IDText:                            func() Packet { return &Text{} },
		IDMovePlayer:                      func() Packet { return &MovePlayer{} },
		IDPassengerJump:                   func() Packet { return &PassengerJump{} },
		IDTickSync:                        func() Packet { return &TickSync{} },
		IDInventoryTransaction:            func() Packet { return &InventoryTransaction{} },
		IDMobEquipment:                    func() Packet { return &MobEquipment{} },
		IDInteract:                        func() Packet { return &Interact{} },
		IDBlockPickRequest:                func() Packet { return &BlockPickRequest{} },
		IDActorPickRequest:                func() Packet { return &ActorPickRequest{} },
		IDPlayerAction:                    func() Packet { return &PlayerAction{} },
		IDRespawn:                         func() Packet { return &Respawn{} },
		IDContainerOpen:                   func() Packet { return &ContainerOpen{} },
		IDContainerClose:                  func() Packet { return &ContainerClose{} },
		IDCraftingEvent:                   func() Packet { return &CraftingEvent{} },
		IDAdventureSettings:               func() Packet { return &AdventureSettings{} },
		IDPlayerInput:                     func() Packet { return &PlayerInput{} },
		IDSetPlayerGameType:               func() Packet { return &SetPlayerGameType{} },
		IDMapInfoRequest:                  func() Packet { return &MapInfoRequest{} },
		IDRequestChunkRadius:              func() Packet { return &RequestChunkRadius{} },
		IDCommandRequest:                  func() Packet { return &CommandRequest{} },
		IDCommandBlockUpdate:              func() Packet { return &CommandBlockUpdate{} },
		IDResourcePackChunkRequest:        func() Packet { return &ResourcePackChunkRequest{} },
		IDStructureBlockUpdate:            func() Packet { return &StructureBlockUpdate{} },
		IDPurchaseReceipt:                 func() Packet { return &PurchaseReceipt{} },
		IDPlayerSkin:                      func() Packet { return &PlayerSkin{} },
		IDSubClientLogin:                  func() Packet { return &SubClientLogin{} },
		IDAutomationClientConnect:         func() Packet { return &AutomationClientConnect{} },
		IDBookEdit:                        func() Packet { return &BookEdit{} },
		IDNPCRequest:                      func() Packet { return &NPCRequest{} },
		IDModalFormRequest:                func() Packet { return &ModalFormRequest{} },
		IDServerSettingsRequest:           func() Packet { return &ServerSettingsRequest{} },
		IDSetDefaultGameType:              func() Packet { return &SetDefaultGameType{} },
		IDLabTable:                        func() Packet { return &LabTable{} },
		IDSetLocalPlayerAsInitialised:     func() Packet { return &SetLocalPlayerAsInitialised{} },
		IDNetworkStackLatency:             func() Packet { return &NetworkStackLatency{} },
		IDScriptCustomEvent:               func() Packet { return &ScriptCustomEvent{} },
		IDLecternUpdate:                   func() Packet { return &LecternUpdate{} },
		IDClientCacheStatus:               func() Packet { return &ClientCacheStatus{} },
		IDMapCreateLockedCopy:             func() Packet { return &MapCreateLockedCopy{} },
		IDStructureTemplateDataResponse:   func() Packet { return &StructureTemplateDataResponse{} },
		IDClientCacheBlobStatus:           func() Packet { return &ClientCacheBlobStatus{} },
		IDEmote:                           func() Packet { return &Emote{} },
		IDMultiPlayerSettings:             func() Packet { return &MultiPlayerSettings{} },
		IDSettingsCommand:                 func() Packet { return &SettingsCommand{} },
		IDAnvilDamage:                     func() Packet { return &AnvilDamage{} },
		IDPlayerAuthInput:                 func() Packet { return &PlayerAuthInput{} },
		IDItemStackRequest:                func() Packet { return &ItemStackRequest{} },
		IDUpdatePlayerGameType:            func() Packet { return &UpdatePlayerGameType{} },
		IDEmoteList:                       func() Packet { return &EmoteList{} },
		IDPositionTrackingDBClientRequest: func() Packet { return &PositionTrackingDBClientRequest{} },
		IDDebugInfo:                       func() Packet { return &DebugInfo{} },
		IDPacketViolationWarning:          func() Packet { return &PacketViolationWarning{} },
		IDFilterText:                      func() Packet { return &FilterText{} },
		IDCreatePhoto:                     func() Packet { return &CreatePhoto{} },
		IDPhotoInfoRequest:                func() Packet { return &PhotoInfoRequest{} },
		IDSubChunkRequest:                 func() Packet { return &SubChunkRequest{} },
		IDScriptMessage:                   func() Packet { return &ScriptMessage{} },
		IDCodeBuilderSource:               func() Packet { return &CodeBuilderSource{} },
		IDRequestAbility:                  func() Packet { return &RequestAbility{} },
		IDRequestPermissions:              func() Packet { return &RequestPermissions{} },
		IDEditorNetwork:                   func() Packet { return &EditorNetwork{} },
		IDRequestNetworkSettings:          func() Packet { return &RequestNetworkSettings{} },
		IDGameTestResults:                 func() Packet { return &GameTestResults{} },
		IDOpenSign:                        func() Packet { return &OpenSign{} },
	}
	for id, pk := range clientOriginating {
		RegisterPacketFromClient(id, pk)
	}
}
