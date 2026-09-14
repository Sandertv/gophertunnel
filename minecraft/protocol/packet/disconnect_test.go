package packet

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type disconnectSelectorIO struct {
	protocol.IO
	varuint32Calls int
	boolCalls      int
	selector       uint32
}

func (io *disconnectSelectorIO) Varint32(*int32) {}

func (io *disconnectSelectorIO) Varuint32(x *uint32) {
	io.varuint32Calls++
	io.selector = *x
}

func (io *disconnectSelectorIO) Bool(*bool) {
	io.boolCalls++
}

func (io *disconnectSelectorIO) String(*string) {}

func TestDisconnectUsesVaruint32MessageSelector(t *testing.T) {
	pk := &Disconnect{HideDisconnectionScreen: true}
	io := new(disconnectSelectorIO)

	pk.Marshal(io)

	if io.varuint32Calls != 1 {
		t.Fatalf("Varuint32 calls = %d, want 1", io.varuint32Calls)
	}
	if io.selector != 1 {
		t.Fatalf("selector = %d, want 1", io.selector)
	}
	if io.boolCalls != 0 {
		t.Fatalf("Bool calls = %d, want 0", io.boolCalls)
	}
}
