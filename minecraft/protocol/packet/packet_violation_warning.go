package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ViolationTypeMalformed ViolationType = iota
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
	Type ViolationType
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
func (pk *PacketViolationWarning) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint32(buf, int32(pk.Type))
	_ = protocol.WriteVarint32(buf, pk.Severity)
	_ = protocol.WriteVarint32(buf, pk.PacketID)
	_ = protocol.WriteString(buf, pk.ViolationContext)
}

// Unmarshal ...
func (pk *PacketViolationWarning) Unmarshal(buf *bytes.Buffer) error {
	var t int32
	err := chainErr(
		protocol.Varint32(buf, &t),
		protocol.Varint32(buf, &pk.Severity),
		protocol.Varint32(buf, &pk.PacketID),
		protocol.String(buf, &pk.ViolationContext),
	)
	pk.Type = ViolationType(t)
	return err
}

// ViolationType implements Stringer to convert a violation type to a string.
type ViolationType int32

// String ...
func (v ViolationType) String() string {
	switch v {
	case ViolationTypeMalformed:
		return "Malformed"
	}
	return "Unknown"
}
