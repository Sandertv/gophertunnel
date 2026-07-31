package packet

import (
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// idFieldPattern matches struct field names that suggest the field carries an entity
// runtime or unique ID. Fields named after their meaning rather than their ID kind are
// listed explicitly.
var idFieldPattern = regexp.MustCompile(`RuntimeID|UniqueID|ActorID|EntityID|ClientPredictedVehicle|AttachToEntity`)

// translatedIDFields lists every entity ID field, as a path from the packet struct, that
// TranslateEntityIDs rewrites and that idFieldPattern can discover. IDs behind interface
// fields and entity metadata map values are translated but not discoverable.
var translatedIDFields = []string{
	"ActorEvent.EntityRuntimeID",
	"ActorPickRequest.EntityUniqueID",
	"AddActor.EntityLinks[].RiddenEntityUniqueID",
	"AddActor.EntityLinks[].RiderEntityUniqueID",
	"AddActor.EntityRuntimeID",
	"AddActor.EntityUniqueID",
	"AddItemActor.EntityRuntimeID",
	"AddItemActor.EntityUniqueID",
	"AddPainting.EntityRuntimeID",
	"AddPainting.EntityUniqueID",
	"AddPlayer.AbilityData.EntityUniqueID",
	"AddPlayer.EntityLinks[].RiddenEntityUniqueID",
	"AddPlayer.EntityLinks[].RiderEntityUniqueID",
	"AddPlayer.EntityRuntimeID",
	"AddVolumeEntity.EntityRuntimeID",
	"AdventureSettings.PlayerUniqueID",
	"AgentAnimation.EntityRuntimeID",
	"Animate.EntityRuntimeID",
	"AnimateEntity.EntityRuntimeIDs",
	"BossEvent.BossEntityUniqueID",
	"BossEvent.PlayerUniqueID",
	"Camera.CameraEntityUniqueID",
	"Camera.TargetPlayerUniqueID",
	"CameraInstruction.AttachToEntity",
	"CameraInstruction.Target.EntityUniqueID",
	"ChangeMobProperty.EntityUniqueID",
	"ClientBoundMapItemData.TrackedObjects[].EntityUniqueID",
	"ClientCheatAbility.AbilityData.EntityUniqueID",
	"ClientMovementPredictionSync.EntityUniqueID",
	"CommandBlockUpdate.MinecartEntityRuntimeID",
	"CommandOutput.CommandOrigin.PlayerUniqueID",
	"CommandRequest.CommandOrigin.PlayerUniqueID",
	"ContainerOpen.ContainerEntityUniqueID",
	"CreatePhoto.EntityUniqueID",
	"DebugInfo.PlayerUniqueID",
	"Emote.EntityRuntimeID",
	"EmoteList.PlayerRuntimeID",
	"Event.EntityRuntimeID",
	"Interact.TargetEntityRuntimeID",
	"LevelSoundEvent.EntityUniqueID",
	"LocatorBar.Waypoints[].Waypoint.ActorUniqueID",
	"MobArmourEquipment.EntityRuntimeID",
	"MobEffect.EntityRuntimeID",
	"MobEquipment.EntityRuntimeID",
	"MotionPredictionHints.EntityRuntimeID",
	"MoveActorAbsolute.EntityRuntimeID",
	"MoveActorDelta.EntityRuntimeID",
	"MovePlayer.EntityRuntimeID",
	"MovePlayer.RiddenEntityRuntimeID",
	"MovementEffect.EntityRuntimeID",
	"NPCDialogue.EntityUniqueID",
	"NPCRequest.EntityRuntimeID",
	"PhotoTransfer.OwnerEntityUniqueID",
	"PlayerAction.EntityRuntimeID",
	"PlayerAuthInput.ClientPredictedVehicle",
	"PlayerList.Entries[].EntityUniqueID",
	"PlayerLocation.EntityUniqueID",
	"PlayerUpdateEntityOverrides.EntityUniqueID",
	"PrimitiveShapes.Shapes[].AttachedToEntityID",
	"RemoveActor.EntityUniqueID",
	"RemoveVolumeEntity.EntityRuntimeID",
	"RequestPermissions.EntityUniqueID",
	"Respawn.EntityRuntimeID",
	"SetActorData.EntityRuntimeID",
	"SetActorLink.EntityLink.RiddenEntityUniqueID",
	"SetActorLink.EntityLink.RiderEntityUniqueID",
	"SetActorMotion.EntityRuntimeID",
	"SetLocalPlayerAsInitialised.EntityRuntimeID",
	"SetScore.Entries[].EntityUniqueID",
	"SetScoreboardIdentity.Entries[].EntityUniqueID",
	"ShowCredits.PlayerRuntimeID",
	"SpawnParticleEffect.EntityUniqueID",
	"StartGame.EntityRuntimeID",
	"StartGame.EntityUniqueID",
	"StructureBlockUpdate.Settings.LastEditingPlayerUniqueID",
	"StructureTemplateDataRequest.Settings.LastEditingPlayerUniqueID",
	"TakeItemActor.ItemEntityRuntimeID",
	"TakeItemActor.TakerEntityRuntimeID",
	"UpdateAbilities.AbilityData.EntityUniqueID",
	"UpdateAttributes.EntityRuntimeID",
	"UpdateBlockSynced.EntityUniqueID",
	"UpdateEquip.EntityUniqueID",
	"UpdatePlayerGameType.PlayerUniqueID",
	"UpdateSubChunkBlocks.Blocks[].SyncedUpdateEntityUniqueID",
	"UpdateSubChunkBlocks.Extra[].SyncedUpdateEntityUniqueID",
	"UpdateTrade.EntityUniqueID",
	"UpdateTrade.VillagerUniqueID",
}

// ignoredIDFields lists fields matched by the naming heuristic that TranslateEntityIDs
// deliberately leaves untouched, with the reason.
var ignoredIDFields = map[string]string{
	"ItemRegistry.Items[].RuntimeID": "item type network ID, not an entity ID",
}

// TestTranslateEntityIDsCoverage fails when a packet field that looks like an entity ID
// is neither translated nor deliberately ignored, so that a protocol update adding one
// breaks this test until TranslateEntityIDs is brought back in sync. Block runtime IDs
// are out of scope; interface implementations carrying entity IDs must be covered by
// hand.
func TestTranslateEntityIDsCoverage(t *testing.T) {
	discovered := map[string]bool{}
	seen := map[reflect.Type]bool{}
	for _, pool := range []Pool{NewClientPool(), NewServerPool()} {
		for _, newPk := range pool {
			typ := reflect.TypeOf(newPk()).Elem()
			if seen[typ] {
				continue
			}
			seen[typ] = true
			walkIDFields(typ, typ.Name(), discovered, map[reflect.Type]bool{})
		}
	}

	expected := map[string]bool{}
	for _, path := range translatedIDFields {
		expected[path] = true
	}
	for path := range ignoredIDFields {
		expected[path] = true
	}

	var missing, stale []string
	for path := range discovered {
		if !expected[path] {
			missing = append(missing, path)
		}
	}
	for path := range expected {
		if !discovered[path] {
			stale = append(stale, path)
		}
	}
	sort.Strings(missing)
	sort.Strings(stale)
	for _, path := range missing {
		t.Errorf("entity ID field %v is neither translated by TranslateEntityIDs nor listed as ignored", path)
	}
	for _, path := range stale {
		t.Errorf("listed entity ID field %v no longer exists in the packet package", path)
	}
}

// walkIDFields recursively collects the paths of ID-like fields of a packet struct type,
// reporting fields of protocol.Optional types as the optional field itself.
func walkIDFields(typ reflect.Type, path string, found map[string]bool, visiting map[reflect.Type]bool) {
	switch typ.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		walkIDFields(typ.Elem(), path+"[]", found, visiting)
	case reflect.Map:
		walkIDFields(typ.Elem(), path+"{}", found, visiting)
	case reflect.Struct:
		if visiting[typ] {
			return
		}
		visiting[typ] = true
		defer delete(visiting, typ)
		optional := strings.HasPrefix(typ.Name(), "Optional[")
		for field := range typ.Fields() {
			fieldPath := path + "." + field.Name
			if optional {
				fieldPath = path
			}
			if !optional && idFieldPattern.MatchString(field.Name) && !strings.Contains(field.Name, "BlockRuntimeID") {
				found[fieldPath] = true
				continue
			}
			walkIDFields(field.Type, fieldPath, found, visiting)
		}
	}
}

// translateSwap swaps between the (10, 100) and (20, 200) (unique, runtime) identities.
func translateSwap(pk Packet) {
	rid := func(id uint64) uint64 {
		switch id {
		case 100:
			return 200
		case 200:
			return 100
		}
		return id
	}
	uid := func(id int64) int64 {
		switch id {
		case 10:
			return 20
		case 20:
			return 10
		}
		return id
	}
	TranslateEntityIDs(pk, rid, uid)
}

func TestTranslateEntityIDsSimpleFields(t *testing.T) {
	pk := &StartGame{EntityUniqueID: 10, EntityRuntimeID: 200}
	translateSwap(pk)
	if pk.EntityUniqueID != 20 || pk.EntityRuntimeID != 100 {
		t.Errorf("unexpected IDs after translation: unique %v, runtime %v", pk.EntityUniqueID, pk.EntityRuntimeID)
	}
	move := &MovePlayer{EntityRuntimeID: 100, RiddenEntityRuntimeID: 300}
	translateSwap(move)
	if move.EntityRuntimeID != 200 || move.RiddenEntityRuntimeID != 300 {
		t.Errorf("unexpected IDs after translation: runtime %v, ridden %v", move.EntityRuntimeID, move.RiddenEntityRuntimeID)
	}
}

func TestTranslateEntityIDsNilCallbacks(t *testing.T) {
	pk := &StartGame{EntityUniqueID: 10, EntityRuntimeID: 100}
	TranslateEntityIDs(pk, nil, nil)
	if pk.EntityUniqueID != 10 || pk.EntityRuntimeID != 100 {
		t.Errorf("nil callbacks changed IDs: unique %v, runtime %v", pk.EntityUniqueID, pk.EntityRuntimeID)
	}
}

func TestTranslateEntityIDsMetadataAndLinks(t *testing.T) {
	pk := &AddActor{
		EntityUniqueID:  10,
		EntityRuntimeID: 100,
		EntityMetadata: map[uint32]any{
			protocol.EntityDataKeyOwner:                    int64(10),
			protocol.EntityDataKeyTarget:                   uint64(20),
			protocol.EntityDataKeyLeashHolder:              int64(30),
			protocol.EntityDataKeyAimAssistPriorityActorID: int64(20),
			protocol.EntityDataKeyName:                     "unrelated",
		},
		EntityLinks: []protocol.EntityLink{{RiddenEntityUniqueID: 20, RiderEntityUniqueID: 10}},
	}
	translateSwap(pk)
	if pk.EntityUniqueID != 20 || pk.EntityRuntimeID != 200 {
		t.Errorf("unexpected IDs after translation: unique %v, runtime %v", pk.EntityUniqueID, pk.EntityRuntimeID)
	}
	if owner := pk.EntityMetadata[protocol.EntityDataKeyOwner]; owner != int64(20) {
		t.Errorf("unexpected owner metadata after translation: %v", owner)
	}
	if target := pk.EntityMetadata[protocol.EntityDataKeyTarget]; target != uint64(10) {
		t.Errorf("unexpected target metadata after translation: %v", target)
	}
	if holder := pk.EntityMetadata[protocol.EntityDataKeyLeashHolder]; holder != int64(30) {
		t.Errorf("unexpected leash holder metadata after translation: %v", holder)
	}
	if aim := pk.EntityMetadata[protocol.EntityDataKeyAimAssistPriorityActorID]; aim != int64(10) {
		t.Errorf("unexpected aim assist actor metadata after translation: %v", aim)
	}
	if link := pk.EntityLinks[0]; link.RiddenEntityUniqueID != 10 || link.RiderEntityUniqueID != 20 {
		t.Errorf("unexpected entity link after translation: %+v", link)
	}
}

func TestTranslateEntityIDsConditionalFields(t *testing.T) {
	block := &CommandBlockUpdate{Block: true, MinecartEntityRuntimeID: 100}
	translateSwap(block)
	if block.MinecartEntityRuntimeID != 100 {
		t.Errorf("block-mode CommandBlockUpdate minecart ID translated to %v", block.MinecartEntityRuntimeID)
	}
	minecart := &CommandBlockUpdate{Block: false, MinecartEntityRuntimeID: 100}
	translateSwap(minecart)
	if minecart.MinecartEntityRuntimeID != 200 {
		t.Errorf("minecart CommandBlockUpdate ID not translated: %v", minecart.MinecartEntityRuntimeID)
	}

	input := &PlayerAuthInput{ClientPredictedVehicle: 10, InputData: protocol.NewBitset(PlayerAuthInputBitsetSize)}
	translateSwap(input)
	if input.ClientPredictedVehicle != 10 {
		t.Errorf("predicted vehicle translated without its input flag: %v", input.ClientPredictedVehicle)
	}
	input.InputData.Set(InputFlagClientPredictedVehicle)
	translateSwap(input)
	if input.ClientPredictedVehicle != 20 {
		t.Errorf("predicted vehicle not translated with its input flag: %v", input.ClientPredictedVehicle)
	}

	score := &SetScore{Entries: []protocol.ScoreboardEntry{
		{IdentityType: protocol.ScoreboardIdentityPlayer, EntityUniqueID: 10},
		{IdentityType: protocol.ScoreboardIdentityFakePlayer, EntityUniqueID: 10},
	}}
	translateSwap(score)
	if score.Entries[0].EntityUniqueID != 20 {
		t.Errorf("player scoreboard entry not translated: %v", score.Entries[0].EntityUniqueID)
	}
	if score.Entries[1].EntityUniqueID != 10 {
		t.Errorf("fake player scoreboard entry translated: %v", score.Entries[1].EntityUniqueID)
	}
}

func TestTranslateEntityIDsInterfaceFields(t *testing.T) {
	tx := &InventoryTransaction{TransactionData: &protocol.UseItemOnEntityTransactionData{TargetEntityRuntimeID: 100}}
	translateSwap(tx)
	if id := tx.TransactionData.(*protocol.UseItemOnEntityTransactionData).TargetEntityRuntimeID; id != 200 {
		t.Errorf("use item on entity target not translated: %v", id)
	}

	event := &Event{EntityRuntimeID: 100, Event: &protocol.MobKilledEvent{KillerEntityUniqueID: 10, VictimEntityUniqueID: 20}}
	translateSwap(event)
	if event.EntityRuntimeID != 200 {
		t.Errorf("event runtime ID not translated: %v", event.EntityRuntimeID)
	}
	killed := event.Event.(*protocol.MobKilledEvent)
	if killed.KillerEntityUniqueID != 20 || killed.VictimEntityUniqueID != 10 {
		t.Errorf("mob killed event not translated: %+v", killed)
	}

	interact := &Event{Event: &protocol.EntityInteractEvent{InteractedEntityID: 10}}
	translateSwap(interact)
	if id := interact.Event.(*protocol.EntityInteractEvent).InteractedEntityID; id != 20 {
		t.Errorf("entity interact event not translated: %v", id)
	}
}

func TestTranslateEntityIDsOptionalFields(t *testing.T) {
	camera := &CameraInstruction{
		Target:         protocol.Option(protocol.CameraInstructionTarget{EntityUniqueID: 10}),
		AttachToEntity: protocol.Option(int64(20)),
	}
	translateSwap(camera)
	if target, _ := camera.Target.Value(); target.EntityUniqueID != 20 {
		t.Errorf("camera target not translated: %v", target.EntityUniqueID)
	}
	if attached, _ := camera.AttachToEntity.Value(); attached != 10 {
		t.Errorf("camera attach entity not translated: %v", attached)
	}

	empty := &CameraInstruction{}
	translateSwap(empty)
	if _, ok := empty.Target.Value(); ok {
		t.Error("empty camera target became set")
	}
}
