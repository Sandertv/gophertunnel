package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// entityMetadataIDKeys holds the entity metadata keys whose values are entity unique IDs.
var entityMetadataIDKeys = []uint32{
	protocol.EntityDataKeyOwner,
	protocol.EntityDataKeyTarget,
	protocol.EntityDataKeyLeashHolder,
	protocol.EntityDataKeyTargetA,
	protocol.EntityDataKeyTargetB,
	protocol.EntityDataKeyTargetC,
	protocol.EntityDataKeyTradeTarget,
	protocol.EntityDataKeyBalloonAnchor,
	protocol.EntityDataKeyAgent,
	protocol.EntityDataKeyAimAssistPriorityActorID,
	protocol.EntityDataKeyArrowShooterID,
	protocol.EntityDataKeyFireworkShooterID,
}

// TranslateEntityIDs passes every entity runtime ID and entity unique ID carried by pk,
// including IDs nested in structures such as entity links or metadata, through runtimeID
// and uniqueID respectively, storing the results. Either function may be nil to leave IDs
// of that kind untouched. The callbacks decide the mapping policy and must return their
// argument unchanged for IDs they do not translate, such as sentinel values like
// math.MaxInt64. Runtime IDs held in uint32 fields are truncated to their width.
func TranslateEntityIDs(pk Packet, runtimeID func(uint64) uint64, uniqueID func(int64) int64) {
	rid := runtimeID
	if rid == nil {
		rid = func(id uint64) uint64 { return id }
	}
	uid := uniqueID
	if uid == nil {
		uid = func(id int64) int64 { return id }
	}
	entityLink := func(link protocol.EntityLink) protocol.EntityLink {
		link.RiddenEntityUniqueID = uid(link.RiddenEntityUniqueID)
		link.RiderEntityUniqueID = uid(link.RiderEntityUniqueID)
		return link
	}
	metadata := func(values map[uint32]any) {
		for _, key := range entityMetadataIDKeys {
			switch id := values[key].(type) {
			case int64:
				values[key] = uid(id)
			case uint64:
				values[key] = uint64(uid(int64(id)))
			}
		}
	}

	switch pk := pk.(type) {
	case *ActorEvent:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *ActorPickRequest:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
	case *AgentAnimation:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *AddActor:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
		metadata(pk.EntityMetadata)
		for i := range pk.EntityLinks {
			pk.EntityLinks[i] = entityLink(pk.EntityLinks[i])
		}
	case *AddItemActor:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
		metadata(pk.EntityMetadata)
	case *AddPainting:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *AddPlayer:
		pk.AbilityData.EntityUniqueID = uid(pk.AbilityData.EntityUniqueID)
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
		metadata(pk.EntityMetadata)
		for i := range pk.EntityLinks {
			pk.EntityLinks[i] = entityLink(pk.EntityLinks[i])
		}
	case *AddVolumeEntity:
		pk.EntityRuntimeID = uint32(rid(uint64(pk.EntityRuntimeID)))
	case *AdventureSettings:
		pk.PlayerUniqueID = uid(pk.PlayerUniqueID)
	case *Animate:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *AnimateEntity:
		for i := range pk.EntityRuntimeIDs {
			pk.EntityRuntimeIDs[i] = rid(pk.EntityRuntimeIDs[i])
		}
	case *BossEvent:
		pk.BossEntityUniqueID = uid(pk.BossEntityUniqueID)
		pk.PlayerUniqueID = uid(pk.PlayerUniqueID)
	case *Camera:
		pk.CameraEntityUniqueID = uid(pk.CameraEntityUniqueID)
		pk.TargetPlayerUniqueID = uid(pk.TargetPlayerUniqueID)
	case *CameraInstruction:
		if target, ok := pk.Target.Value(); ok {
			target.EntityUniqueID = uid(target.EntityUniqueID)
			pk.Target = protocol.Option(target)
		}
		if attached, ok := pk.AttachToEntity.Value(); ok {
			pk.AttachToEntity = protocol.Option(uid(attached))
		}
	case *ChangeMobProperty:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
	case *ClientBoundMapItemData:
		for i := range pk.TrackedObjects {
			if pk.TrackedObjects[i].Type == protocol.MapObjectTypeEntity {
				pk.TrackedObjects[i].EntityUniqueID = uid(pk.TrackedObjects[i].EntityUniqueID)
			}
		}
	case *ClientCheatAbility:
		pk.AbilityData.EntityUniqueID = uid(pk.AbilityData.EntityUniqueID)
	case *ClientMovementPredictionSync:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
	case *CommandBlockUpdate:
		if !pk.Block {
			pk.MinecartEntityRuntimeID = rid(pk.MinecartEntityRuntimeID)
		}
	case *CommandOutput:
		pk.CommandOrigin.PlayerUniqueID = uid(pk.CommandOrigin.PlayerUniqueID)
	case *CommandRequest:
		pk.CommandOrigin.PlayerUniqueID = uid(pk.CommandOrigin.PlayerUniqueID)
	case *ContainerOpen:
		pk.ContainerEntityUniqueID = uid(pk.ContainerEntityUniqueID)
	case *CreatePhoto:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
	case *DebugInfo:
		pk.PlayerUniqueID = uid(pk.PlayerUniqueID)
	case *Emote:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *EmoteList:
		pk.PlayerRuntimeID = rid(pk.PlayerRuntimeID)
	case *Event:
		pk.EntityRuntimeID = int64(rid(uint64(pk.EntityRuntimeID)))
		switch event := pk.Event.(type) {
		case *protocol.MobKilledEvent:
			event.KillerEntityUniqueID = uid(event.KillerEntityUniqueID)
			event.VictimEntityUniqueID = uid(event.VictimEntityUniqueID)
		case *protocol.BossKilledEvent:
			event.BossEntityUniqueID = uid(event.BossEntityUniqueID)
		case *protocol.EntityInteractEvent:
			event.InteractedEntityID = uid(event.InteractedEntityID)
		}
	case *Interact:
		pk.TargetEntityRuntimeID = rid(pk.TargetEntityRuntimeID)
	case *InventoryTransaction:
		if data, ok := pk.TransactionData.(*protocol.UseItemOnEntityTransactionData); ok {
			data.TargetEntityRuntimeID = rid(data.TargetEntityRuntimeID)
		}
	case *LevelSoundEvent:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
	case *LocatorBar:
		for i := range pk.Waypoints {
			if id, ok := pk.Waypoints[i].Waypoint.ActorUniqueID.Value(); ok {
				pk.Waypoints[i].Waypoint.ActorUniqueID = protocol.Option(uid(id))
			}
		}
	case *MobArmourEquipment:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *MobEffect:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *MobEquipment:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *MotionPredictionHints:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *MovementEffect:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *MoveActorAbsolute:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *MoveActorDelta:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *MovePlayer:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
		pk.RiddenEntityRuntimeID = rid(pk.RiddenEntityRuntimeID)
	case *NPCDialogue:
		pk.EntityUniqueID = uint64(uid(int64(pk.EntityUniqueID)))
	case *NPCRequest:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *PhotoTransfer:
		pk.OwnerEntityUniqueID = uid(pk.OwnerEntityUniqueID)
	case *PlayerAction:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *PlayerAuthInput:
		if pk.InputData.Load(InputFlagClientPredictedVehicle) {
			pk.ClientPredictedVehicle = uid(pk.ClientPredictedVehicle)
		}
	case *PlayerLocation:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
	case *PlayerList:
		for i := range pk.Entries {
			pk.Entries[i].EntityUniqueID = uid(pk.Entries[i].EntityUniqueID)
		}
	case *PlayerUpdateEntityOverrides:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
	case *PrimitiveShapes:
		for i := range pk.Shapes {
			if id, ok := pk.Shapes[i].AttachedToEntityID.Value(); ok {
				pk.Shapes[i].AttachedToEntityID = protocol.Option(uid(id))
			}
		}
	case *RemoveActor:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
	case *RemoveVolumeEntity:
		pk.EntityRuntimeID = uint32(rid(uint64(pk.EntityRuntimeID)))
	case *Respawn:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *RequestPermissions:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
	case *SetActorData:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
		metadata(pk.EntityMetadata)
	case *SetActorLink:
		pk.EntityLink = entityLink(pk.EntityLink)
	case *SetActorMotion:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *SetLocalPlayerAsInitialised:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *SetScore:
		for i := range pk.Entries {
			if pk.Entries[i].IdentityType != protocol.ScoreboardIdentityFakePlayer {
				pk.Entries[i].EntityUniqueID = uid(pk.Entries[i].EntityUniqueID)
			}
		}
	case *SetScoreboardIdentity:
		if pk.ActionType != ScoreboardIdentityActionClear {
			for i := range pk.Entries {
				pk.Entries[i].EntityUniqueID = uid(pk.Entries[i].EntityUniqueID)
			}
		}
	case *ShowCredits:
		pk.PlayerRuntimeID = rid(pk.PlayerRuntimeID)
	case *SpawnParticleEffect:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
	case *StartGame:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *StructureBlockUpdate:
		pk.Settings.LastEditingPlayerUniqueID = uid(pk.Settings.LastEditingPlayerUniqueID)
	case *StructureTemplateDataRequest:
		pk.Settings.LastEditingPlayerUniqueID = uid(pk.Settings.LastEditingPlayerUniqueID)
	case *TakeItemActor:
		pk.ItemEntityRuntimeID = rid(pk.ItemEntityRuntimeID)
		pk.TakerEntityRuntimeID = rid(pk.TakerEntityRuntimeID)
	case *UpdateAttributes:
		pk.EntityRuntimeID = rid(pk.EntityRuntimeID)
	case *UpdateAbilities:
		pk.AbilityData.EntityUniqueID = uid(pk.AbilityData.EntityUniqueID)
	case *UpdateBlockSynced:
		pk.EntityUniqueID = uint64(uid(int64(pk.EntityUniqueID)))
	case *UpdateEquip:
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
	case *UpdatePlayerGameType:
		pk.PlayerUniqueID = uid(pk.PlayerUniqueID)
	case *UpdateSubChunkBlocks:
		for i := range pk.Blocks {
			pk.Blocks[i].SyncedUpdateEntityUniqueID =
				uint64(uid(int64(pk.Blocks[i].SyncedUpdateEntityUniqueID)))
		}
		for i := range pk.Extra {
			pk.Extra[i].SyncedUpdateEntityUniqueID =
				uint64(uid(int64(pk.Extra[i].SyncedUpdateEntityUniqueID)))
		}
	case *UpdateTrade:
		pk.VillagerUniqueID = uid(pk.VillagerUniqueID)
		pk.EntityUniqueID = uid(pk.EntityUniqueID)
	}
}
