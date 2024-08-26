package discovery

import "io"

type RequestPacket struct{}

func (*RequestPacket) ID() uint16 { return IDRequestPacket }

func (*RequestPacket) Read(io.Reader) error { return nil }

func (*RequestPacket) Write(io.Writer) {}
