package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CodeBuilder is an Education Edition packet sent by the server to the client to open the URL to a Code
// Builder (websocket) server.
type CodeBuilder struct {
	// URL is the url to the Code Builder (websocket) server.
	URL string
	// ShouldOpenCodeBuilder specifies if the client should automatically open the Code Builder app. If set to
	// true, the client will attempt to use the Code Builder app to connect to and interface with the server
	// running at the URL above.
	ShouldOpenCodeBuilder bool
}

// ID ...
func (*CodeBuilder) ID() uint32 {
	return IDCodeBuilder
}

// Marshal ...
func (pk *CodeBuilder) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.URL)
	_ = binary.Write(buf, binary.LittleEndian, pk.ShouldOpenCodeBuilder)
}

// Unmarshal ...
func (pk *CodeBuilder) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.URL),
		binary.Read(buf, binary.LittleEndian, &pk.ShouldOpenCodeBuilder),
	)
}
