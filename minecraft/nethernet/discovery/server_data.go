package discovery

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ServerData struct {
	Version        uint8
	ServerName     string
	LevelName      string
	GameType       int32
	PlayerCount    int32
	MaxPlayerCount int32
	IsEditorWorld  bool
	TransportLayer int32
}

func (d *ServerData) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	_ = binary.Write(buf, binary.LittleEndian, d.Version)
	writeBytes[uint8](buf, []byte(d.ServerName))
	writeBytes[uint8](buf, []byte(d.LevelName))
	_ = binary.Write(buf, binary.LittleEndian, d.GameType)
	_ = binary.Write(buf, binary.LittleEndian, d.PlayerCount)
	_ = binary.Write(buf, binary.LittleEndian, d.MaxPlayerCount)
	_ = binary.Write(buf, binary.LittleEndian, d.IsEditorWorld)
	_ = binary.Write(buf, binary.LittleEndian, d.TransportLayer)

	return buf.Bytes(), nil
}

func (d *ServerData) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)

	if err := binary.Read(buf, binary.LittleEndian, &d.Version); err != nil {
		return fmt.Errorf("read version: %w", err)
	}
	serverName, err := readBytes[uint8](buf)
	if err != nil {
		return fmt.Errorf("read server name: %w", err)
	}
	d.ServerName = string(serverName)
	levelName, err := readBytes[uint8](buf)
	if err != nil {
		return fmt.Errorf("read level name: %w", err)
	}
	d.LevelName = string(levelName)
	if err := binary.Read(buf, binary.LittleEndian, &d.GameType); err != nil {
		return fmt.Errorf("read game type: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &d.PlayerCount); err != nil {
		return fmt.Errorf("read player count: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &d.MaxPlayerCount); err != nil {
		return fmt.Errorf("read max player count: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &d.IsEditorWorld); err != nil {
		return fmt.Errorf("read editor world: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &d.TransportLayer); err != nil {
		return fmt.Errorf("read transport layer: %w", err)
	}

	return nil
}
