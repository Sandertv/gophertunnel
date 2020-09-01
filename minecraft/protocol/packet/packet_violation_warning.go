package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ViolationTypeMalformed = iota
)

const (
	ViolationSeverityWarning = iota
	ViolationSeverityFinalWarning
	ViolationSeverityTerminatingConnection
)

// PacketViolationWarning is sent by the client when it receives an invalid packet from the server. It holds
// some information on the error that occurred.
//noinspection GoNameStartsWithPackageName
type PacketViolationWarning struct {
	// Type is the type of violation. It is one of the constants above.
	Type int32
	// Severity specifies the severity of the packet violation. The action the client takes after this
	// violation depends on the severity sent.
	Severity int32
	// PacketID is the ID of the invalid packet that was received.
	PacketID int32
	// ViolationContext holds a description on the violation of the packet.
	ViolationContext string
}

// ID ...
func (*PacketViolationWarning) ID() uint32 {
	return IDPacketViolationWarning
}

// Marshal ...
func (pk *PacketViolationWarning) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.Type)
	w.Varint32(&pk.Severity)
	w.Varint32(&pk.PacketID)
	w.String(&pk.ViolationContext)
}

// Unmarshal ...
func (pk *PacketViolationWarning) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.Type)
	r.Varint32(&pk.Severity)
	r.Varint32(&pk.PacketID)
	r.String(&pk.ViolationContext)
}
