package login

import (
	"encoding/json"
	"testing"
)

func TestClientDataEditorConnectionFieldsDecode(t *testing.T) {
	var data ClientData
	if err := json.Unmarshal([]byte(`{"ClientIsEditorCapable":true,"ClientEditorConnectionIntent":2}`), &data); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !data.ClientIsEditorCapable {
		t.Fatal("ClientIsEditorCapable = false, want true")
	}
	if got := data.ClientEditorConnectionIntent; got != 2 {
		t.Fatalf("ClientEditorConnectionIntent = %d, want 2", got)
	}
}
