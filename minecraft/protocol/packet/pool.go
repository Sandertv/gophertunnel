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
		IDLevelEvent:           &LevelEvent{},
		IDBlockEvent:           &BlockEvent{},
		IDEntityEvent:          &EntityEvent{},
		IDMobEffect:            &MobEffect{},
		IDUpdateAttributes:     &UpdateAttributes{},
		IDInventoryTransaction: &InventoryTransaction{},
		IDMobEquipment:         &MobEquipment{},
		IDMobArmourEquipment:   &MobArmourEquipment{},
		IDInteract:             &Interact{},
		IDBlockPickRequest:     &BlockPickRequest{},
		IDEntityPickRequest:    &EntityPickRequest{},
		IDPlayerAction:         &PlayerAction{},
		IDEntityFall:           &EntityFall{},
		IDHurtArmour:           &HurtArmour{},
		IDSetEntityData:        &SetEntityData{},
		IDSetEntityMotion:      &SetEntityMotion{},
		IDSetEntityLink:        &SetEntityLink{},
		// ...
		IDFullChunkData: &FullChunkData{},
		// ...
		IDRequestChunkRadius: &RequestChunkRadius{},
		IDChunkRadiusUpdated: &ChunkRadiusUpdated{},
		// ...
		IDResourcePackDataInfo:     &ResourcePackDataInfo{},
		IDResourcePackChunkData:    &ResourcePackChunkData{},
		IDResourcePackChunkRequest: &ResourcePackChunkRequest{},
		IDTransfer:                 &Transfer{},
		// ...
		IDModalFormRequest:       &ModalFormRequest{},
		IDModalFormResponse:      &ModalFormResponse{},
		IDServerSettingsRequest:  &ServerSettingsRequest{},
		IDServerSettingsResponse: &ServerSettingsResponse{},
		// ...
		IDSetLocalPlayerAsInitialised: &SetLocalPlayerAsInitialised{},
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
