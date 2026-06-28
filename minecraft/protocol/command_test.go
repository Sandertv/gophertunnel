package protocol

import "testing"

func TestCommandArgTypeWireValues(t *testing.T) {
	tests := []struct {
		name string
		got  uint32
		want uint32
	}{
		{"int", CommandArgTypeInt, 1},
		{"float", CommandArgTypeFloat, 2},
		{"value", CommandArgTypeValue, 3},
		{"rvalue", CommandArgTypeRValue, 4},
		{"wildcard_int", CommandArgTypeWildcardInt, 5},
		{"target", CommandArgTypeTarget, 8},
		{"wildcard_target", CommandArgTypeWildcardTarget, 10},
		{"filepath", CommandArgTypeFilepath, 17},
		{"integer_range", CommandArgTypeIntegerRange, 23},
		{"equipment_slots", CommandArgTypeEquipmentSlots, 47},
		{"string", CommandArgTypeString, 56},
		{"block_position", CommandArgTypeBlockPosition, 64},
		{"position", CommandArgTypePosition, 65},
		{"message", CommandArgTypeMessage, 67},
		{"raw_text", CommandArgTypeRawText, 70},
		{"json", CommandArgTypeJSON, 74},
		{"block_states", CommandArgTypeBlockStates, 84},
		{"clock_time_marker", CommandArgTypeClockTimeMarker, 86},
		{"command", CommandArgTypeCommand, 87},
		{"slash_command", CommandArgTypeSlashCommand, 88},
		{"code_builder_args", CommandArgTypeCodeBuilderArgs, 90},
		{"chained_command", CommandArgTypeChainedCommand, 0x8000000},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}
